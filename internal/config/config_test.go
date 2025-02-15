package config

import (
	"testing"

	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
)

func Test_GetConfig(t *testing.T) {
	config := GetConfig()

	if config == nil {
		t.Error("Expected config instance, got nil")
	}

	if len(config.ClientConfig.FieldsToSkip) != 3 {
		t.Errorf("Expected 3 fields to skip, got %d", len(config.ClientConfig.FieldsToSkip))
	}
}

func Test_SetupLogging(t *testing.T) {
	config := GetConfig()

	// Test default logging level
	config.SetupLogging()
	if log.GetLevel() != logger.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", log.GetLevel())
	}

	// Test DebugLevel logging
	config.VerboseCount = 1
	config.SetupLogging()
	if log.GetLevel() != logger.DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", log.GetLevel())
	}

	// Test TraceLevel logging
	config.VerboseCount = 2
	config.SetupLogging()
	if log.GetLevel() != logger.TraceLevel {
		t.Errorf("Expected TraceLevel, got %v", log.GetLevel())
	}

	// Test TestingMode
	config.ClientConfig.TestingMode = true
	config.SetupLogging()
	if config.ClientConfig.GRPCSecure != false {
		t.Errorf("Expected GRPCSecure to be false, got %v", config.ClientConfig.GRPCSecure)
	}
	if config.ClientConfig.GRPCServerName != "" {
		t.Errorf("Expected GRPCServerName to be empty, got %v", config.ClientConfig.GRPCServerName)
	}
}
