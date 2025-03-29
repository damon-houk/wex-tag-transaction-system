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
}
