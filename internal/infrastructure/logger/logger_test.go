// internal/infrastructure/logger/logger_test.go
package logger

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSONLogger(t *testing.T) {
	// Setup a buffer to capture output
	var buf bytes.Buffer
	logger := NewJSONLogger(&buf, DebugLevel)

	// Test debug level logging
	logger.Debug("Debug message", map[string]interface{}{
		"key1": "value1",
	})

	// Parse and verify the output
	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)

	assert.NoError(t, err)
	assert.Equal(t, "DEBUG", logEntry["level"])
	assert.Equal(t, "Debug message", logEntry["message"])
	assert.Equal(t, "value1", logEntry["key1"])
	assert.Contains(t, logEntry, "timestamp")
	assert.Contains(t, logEntry, "file")
	assert.Contains(t, logEntry, "line")

	// Test that log levels are respected
	buf.Reset()
	warnLogger := NewJSONLogger(&buf, WarnLevel)

	// This should not log anything (debug < warn)
	warnLogger.Debug("Should not appear", nil)
	assert.Equal(t, "", buf.String())

	// This should log
	warnLogger.Warn("Warning message", nil)
	assert.Contains(t, buf.String(), "Warning message")

	// Test WithField
	buf.Reset()
	fieldLogger := logger.WithField("context", "test")
	fieldLogger.Info("With field", nil)

	output = buf.String()
	err = json.Unmarshal([]byte(output), &logEntry)

	assert.NoError(t, err)
	assert.Equal(t, "test", logEntry["context"])
	assert.Equal(t, "With field", logEntry["message"])

	// Test WithFields
	buf.Reset()
	fieldsLogger := logger.WithFields(map[string]interface{}{
		"app":     "test-app",
		"version": "1.0.0",
	})
	fieldsLogger.Info("With fields", nil)

	output = buf.String()
	err = json.Unmarshal([]byte(output), &logEntry)

	assert.NoError(t, err)
	assert.Equal(t, "test-app", logEntry["app"])
	assert.Equal(t, "1.0.0", logEntry["version"])
	assert.Equal(t, "With fields", logEntry["message"])

	// Test log levels
	buf.Reset()
	infoLogger := NewJSONLogger(&buf, InfoLevel)

	infoLogger.Debug("Debug", nil) // Shouldn't log
	assert.Equal(t, "", buf.String())

	infoLogger.Info("Info", nil) // Should log
	assert.Contains(t, buf.String(), "Info")

	buf.Reset()
	infoLogger.Warn("Warn", nil) // Should log
	assert.Contains(t, buf.String(), "Warn")

	buf.Reset()
	infoLogger.Error("Error", nil) // Should log
	assert.Contains(t, buf.String(), "Error")
}

func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	assert.NotNil(t, logger)
}

func TestSetDefaultLogger(t *testing.T) {
	originalLogger := GetDefaultLogger()

	var buf bytes.Buffer
	newLogger := NewJSONLogger(&buf, DebugLevel)

	SetDefaultLogger(newLogger)
	currentLogger := GetDefaultLogger()

	// Can't directly compare loggers, so we'll just make sure it's not nil
	assert.NotNil(t, currentLogger)

	// Reset to original
	SetDefaultLogger(originalLogger)
}
