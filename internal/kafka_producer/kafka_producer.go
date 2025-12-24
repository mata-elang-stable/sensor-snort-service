package kafka_producer

import (
	"fmt"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
	"github.com/mata-elang-stable/sensor-snort-service/internal/pb"
)

var log = logger.GetLogger()

type Producer struct {
	p          *kafka.Producer
	serializer *protobuf.Serializer
	topic      string
}

func NewKafkaProducer(brokers string, schemaRegistryUrl string, topic string, securityProtocol string, pathToCA string, pathToClientKeystore string, clientKeystorePassword string) (*Producer, error) {
	// Determine and validate security protocol and TLS assets.
	effectiveProtocol, err := validateAndDetermineProtocol(securityProtocol, pathToCA, pathToClientKeystore)
	if err != nil {
		return nil, err
	}

	// Build the producer config inline (minimal changes):
	config := &kafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"enable.idempotence":       true,
		"linger.ms":                10,
		"batch.size":               65536,
		"acks":                     "all",
		"message.send.max.retries": 5,
		"message.max.bytes":        1000000000,
		"retry.backoff.ms":         100,
		"socket.keepalive.enable":  true,
		"go.batch.producer":        true,
		"security.protocol":        effectiveProtocol,
	}

	// Only set ssl.ca.location if a CA path is provided and TLS is in use.
	if pathToCA != "" && effectiveProtocol != "PLAINTEXT" {
		(*config)["ssl.ca.location"] = pathToCA
	}

	// Add client keystore when provided to enable mTLS (PKCS#12)
	if pathToClientKeystore != "" {
		if effectiveProtocol != "PLAINTEXT" {
			(*config)["ssl.keystore.location"] = pathToClientKeystore
			if clientKeystorePassword != "" {
				(*config)["ssl.keystore.password"] = clientKeystorePassword
			}
		} else {
			log.Warnf("Client keystore provided but TLS not enabled; ignoring client keystore: %s", pathToClientKeystore)
		}
	}

	p, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Start a go routine to handle delivery reports
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Errorf("Failed to deliver message %s: %v\n", ev.Key, ev.TopicPartition.Error)
				} else {
					log.Tracef("Delivered message to topic %s [%d] at offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			case kafka.Error:
				log.Errorf("Kafka error: %v\n", ev)
			default:
				log.Debugf("Ignored event: %s\n", ev)
			}
		}
	}()

	registryConfig := schemaregistry.NewConfig(schemaRegistryUrl)
	if pathToCA != "" {
		registryConfig.SslCaLocation = pathToCA
	}

	client, err := schemaregistry.NewClient(registryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema registry client: %w", err)
	}

	serializer, err := protobuf.NewSerializer(client, serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create serializer: %w", err)
	}

	log.Infof("Created Kafka producer with brokers: %s, schema registry URL: %s, and topic: %s", brokers, schemaRegistryUrl, topic)

	return &Producer{
		p:          p,
		serializer: serializer,
		topic:      topic,
	}, nil
}

// validateAndDetermineProtocol validates the provided TLS assets and returns the
// effective security protocol (uppercased). If TLS is requested but no TLS material
// (CA or keystore) is provided, we return an error to avoid runtime failures.
func validateAndDetermineProtocol(securityProtocol, pathToCA, pathToClientKeystore string) (string, error) {
	protocol := strings.ToUpper(strings.TrimSpace(securityProtocol))
	if protocol == "" {
		return "PLAINTEXT", nil
	}

	// Only validate TLS requirements for SSL-like protocols
	if protocol == "SSL" || protocol == "SASL_SSL" {
		if pathToCA == "" && pathToClientKeystore == "" {
			return "", fmt.Errorf("security.protocol %s requires a path_to_ca or path_to_client_keystore", protocol)
		}

		// If CA path provided, check existence
		if pathToCA != "" {
			if fi, err := os.Stat(pathToCA); err != nil {
				return "", fmt.Errorf("failed to stat CA file: %w", err)
			} else if fi.IsDir() {
				return "", fmt.Errorf("path_to_ca points to a directory, not a file: %s", pathToCA)
			}
		}

		// If keystore provided, check existence
		if pathToClientKeystore != "" {
			if fi, err := os.Stat(pathToClientKeystore); err != nil {
				return "", fmt.Errorf("failed to stat client keystore: %w", err)
			} else if fi.IsDir() {
				return "", fmt.Errorf("path_to_client_keystore points to a directory, not a file: %s", pathToClientKeystore)
			}
		}
	}

	return protocol, nil
}

// (No helper; inline config above to keep changes minimal.)

func createKafkaMessages(serializer *protobuf.Serializer, topic string, value *pb.SensorEvent) (*kafka.Message, error) {
	payload, err := serializer.Serialize(topic, value)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(value.EventHashSha256),
		Value:          payload,
		Headers: []kafka.Header{
			{Key: "hash_sha256", Value: []byte(value.EventHashSha256)},
			{Key: "sensor_id", Value: []byte(value.SensorId)},
			{Key: "sensor_read_at", Value: []byte(fmt.Sprintf("%d", value.EventReadAt))},
			{Key: "sensor_sent_at", Value: []byte(fmt.Sprintf("%d", value.EventSentAt))},
			{Key: "dc_received_at", Value: []byte(fmt.Sprintf("%d", value.EventReceivedAt))},
		},
	}, nil
}

func (k *Producer) Serialize(value *pb.SensorEvent) ([]byte, error) {
	return k.serializer.Serialize(k.topic, value)
}

func (k *Producer) Produce(value *pb.SensorEvent) error {
	log.Tracef("Producing message: %v\n", value.EventHashSha256)

	// Serialize message
	payload, err := createKafkaMessages(k.serializer, k.topic, value)
	if err != nil {
		return err
	}

	if err := k.p.Produce(payload, nil); err != nil {
		log.Errorf("Failed to produce message with size %d: %v\n", value.EventMetricsCount, err)
		return err
	}

	log.Tracef("Produced message: %v\n", value.EventHashSha256)
	return nil
}

func (k *Producer) Flush(timeoutMs int) int {
	return k.p.Flush(timeoutMs)
}

func (k *Producer) Close() {
	k.p.Close()
}
