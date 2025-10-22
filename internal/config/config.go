package config

import (
	"io"
	"sync"
	"time"

	"github.com/nxadm/tail"

	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
)

type ClientConfig struct {
	// SnortAlertFile is the file to read the Snort alert from.
	SnortAlertFile string `mapstructure:"file"`

	// GRPCServer is the server to connect to.
	GRPCServer string `mapstructure:"server"`

	// GRPCPort is the port to connect to the server.
	GRPCPort int `mapstructure:"port"`

	// GRPCSecure is a flag to determine whether the connection is secure or not.
	GRPCSecure bool `mapstructure:"secure"`

	// GRPCInterval is the interval to send the data to the server.
	GRPCInterval time.Duration `mapstructure:"interval"`

	// SensorID is the ID of the sensor.
	SensorID string `mapstructure:"sensor_id"`

	// GRPCCertFile is the certificate file to connect to the server.
	GRPCCertFile string

	// GRPCServerName is the name of the server.
	GRPCServerName string

	// FieldsToSkip is the fields to skip in the log.
	FieldsToSkip []string

	// MaxClients is the maximum number of clients.
	MaxClients int `mapstructure:"max_clients"`

	// TestingMode is the flag to determine whether the application is in testing mode or not.
	TestingMode bool `mapstructure:"testing_mode"`
}

type ServerConfig struct {
	// GRPCHost is the host to listen to.
	GRPCHost string `mapstructure:"host"`

	// GRPCPort is the port to listen to.
	GRPCPort int `mapstructure:"port"`

	// GRPCSecure is a flag to determine whether the connection is secure or not.
	GRPCSecure bool `mapstructure:"secure"`

	// SchemaRegistryUrl is the schema registry URL.
	SchemaRegistryUrl string `mapstructure:"schema_registry_url"`

	// KafkaBrokers is the Kafka broker to connect to.
	KafkaBrokers string `mapstructure:"kafka_brokers"`

	// KafkaTopic is the Kafka topic.
	KafkaTopic string `mapstructure:"kafka_topic"`
}

type Config struct {
	ClientConfig ClientConfig `mapstructure:"client"`
	ServerConfig ServerConfig `mapstructure:"server"`

	// GRPCMaxMsgSize is the maximum message size.
	GRPCMaxMsgSize int `mapstructure:"max_message_size"`

	// VerboseCount is the verbose level.
	VerboseCount int `mapstructure:"verbose"`
}

var log = logger.GetLogger()

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			ClientConfig: ClientConfig{
				FieldsToSkip: []string{"Seconds", "Timestamp", ""},
			},
			ServerConfig: ServerConfig{},
		}
	})

	return instance
}

func (c *Config) Client() *ClientConfig {
	return &instance.ClientConfig
}

func (c *Config) Server() *ServerConfig {
	return &instance.ServerConfig
}

func (c *Config) GetTailConfig() tail.Config {
	if c.ClientConfig.TestingMode {
		log.Infof("Testing mode enabled.")

		return tail.Config{
			ReOpen:    false,
			Follow:    false,
			MustExist: true,
			Poll:      false,
			CompleteLines: true,
			Logger:    log,
			Location:  &tail.SeekInfo{Whence: io.SeekStart},
		}
	}

	return tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: true,
		Poll:      true,
		CompleteLines: true,
		Logger:    log,
		Location:  &tail.SeekInfo{Whence: io.SeekStart},
	}
}

func (c *Config) SetupLogging() {
	switch instance.VerboseCount {
	case 0:
		log.SetLevel(logger.InfoLevel)
	case 1:
		log.SetLevel(logger.DebugLevel)
	default:
		log.SetLevel(logger.TraceLevel)
	}
	log.WithFields(logger.Fields{
		"LOG_LEVEL": log.GetLevel().String(),
	}).Infoln("Logging level set.")

	// Enable insecure connection if testing mode is enabled
	if instance.ClientConfig.TestingMode {
		instance.ClientConfig.GRPCSecure = false
		instance.ClientConfig.GRPCServerName = ""
	}
}
