package kafka_producer

import (
	"fmt"

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

func NewKafkaProducer(brokers string, schemaRegistryUrl string, topic string) (*Producer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"message.send.max.retries": 5,
		"retry.backoff.ms":         100,
		"enable.idempotence":       true,
		"acks":                     "all",
		"socket.keepalive.enable":  true,
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

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(schemaRegistryUrl))
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
		return err
	}

	log.Debugf("Produced message: %v\n", value.EventHashSha256)
	return nil
}

func (k *Producer) Flush(timeoutMs int) int {
	return k.p.Flush(timeoutMs)
}

func (k *Producer) Close() {
	k.p.Close()
}
