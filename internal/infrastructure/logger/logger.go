// Package logger internal/infrastructure/logger/logger.go
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// Level represents the severity level of a log message
type Level string

const (
	// DebugLevel is used for development messages
	DebugLevel Level = "DEBUG"
	// InfoLevel is used for general operational information
	InfoLevel Level = "INFO"
	// WarnLevel is used for warnings and potential issues
	WarnLevel Level = "WARN"
	// ErrorLevel is used for errors and unexpected events
	ErrorLevel Level = "ERROR"
	// FatalLevel is used for critical errors that require termination
	FatalLevel Level = "FATAL"
)

// Logger defines the interface for the application logger
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// JSONLogger is a logger that outputs structured JSON logs
type JSONLogger struct {
	output io.Writer
	level  Level
	fields map[string]interface{}
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(output io.Writer, level Level) *JSONLogger {
	if output == nil {
		output = os.Stdout
	}

	return &JSONLogger{
		output: output,
		level:  level,
		fields: make(map[string]interface{}),
	}
}

// WithField returns a new logger with the field added to the log context
func (l *JSONLogger) WithField(key string, value interface{}) Logger {
	newFields := make(map[string]interface{}, len(l.fields)+1)

	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}

	// Add new field
	newFields[key] = value

	return &JSONLogger{
		output: l.output,
		level:  l.level,
		fields: newFields,
	}
}

// WithFields returns a new logger with the fields added to the log context
func (l *JSONLogger) WithFields(fields map[string]interface{}) Logger {
	if len(fields) == 0 {
		return l
	}

	newFields := make(map[string]interface{}, len(l.fields)+len(fields))

	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newFields[k] = v
	}

	return &JSONLogger{
		output: l.output,
		level:  l.level,
		fields: newFields,
	}
}

// Debug logs a message at debug level
func (l *JSONLogger) Debug(msg string, fields map[string]interface{}) {
	if l.shouldLog(DebugLevel) {
		l.log(DebugLevel, msg, fields)
	}
}

// Info logs a message at info level
func (l *JSONLogger) Info(msg string, fields map[string]interface{}) {
	if l.shouldLog(InfoLevel) {
		l.log(InfoLevel, msg, fields)
	}
}

// Warn logs a message at warn level
func (l *JSONLogger) Warn(msg string, fields map[string]interface{}) {
	if l.shouldLog(WarnLevel) {
		l.log(WarnLevel, msg, fields)
	}
}

// Error logs a message at error level
func (l *JSONLogger) Error(msg string, fields map[string]interface{}) {
	if l.shouldLog(ErrorLevel) {
		l.log(ErrorLevel, msg, fields)
	}
}

// Fatal logs a message at fatal level and then terminates the program
func (l *JSONLogger) Fatal(msg string, fields map[string]interface{}) {
	if l.shouldLog(FatalLevel) {
		l.log(FatalLevel, msg, fields)
	}
	os.Exit(1)
}

// shouldLog determines if a message at the given level should be logged
func (l *JSONLogger) shouldLog(level Level) bool {
	// Order of severity: DEBUG < INFO < WARN < ERROR < FATAL
	switch l.level {
	case DebugLevel:
		return true
	case InfoLevel:
		return level != DebugLevel
	case WarnLevel:
		return level != DebugLevel && level != InfoLevel
	case ErrorLevel:
		return level == ErrorLevel || level == FatalLevel
	case FatalLevel:
		return level == FatalLevel
	default:
		return true
	}
}

// log outputs a log message with the given level, message, and fields
func (l *JSONLogger) log(level Level, msg string, fields map[string]interface{}) {
	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Create log record
	record := make(map[string]interface{})

	// Add base fields
	record["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	record["level"] = level
	record["message"] = msg
	record["file"] = file
	record["line"] = line

	// Add context fields
	for k, v := range l.fields {
		record[k] = v
	}

	// Add message-specific fields
	for k, v := range fields {
		record[k] = v
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(record)
	if err != nil {
		// If we can't marshal, at least try to output something
		fmt.Fprintf(l.output, "{\"level\":\"ERROR\",\"message\":\"Failed to marshal log entry\",\"error\":\"%s\"}\n", err)
		return
	}

	// Write to output
	jsonData = append(jsonData, '\n')
	_, err = l.output.Write(jsonData)
	if err != nil {
		// Not much we can do if writing fails, but print to stderr as a last resort
		fmt.Fprintf(os.Stderr, "Failed to write log entry: %s\n", err)
	}
}

// Default logger instances
var (
	defaultLogger = NewJSONLogger(os.Stdout, InfoLevel)
)

// GetDefaultLogger returns the default logger
func GetDefaultLogger() Logger {
	return defaultLogger
}

// SetDefaultLogger sets the default logger
func SetDefaultLogger(logger Logger) {
	if l, ok := logger.(*JSONLogger); ok {
		defaultLogger = l
	}
}

// Debug Global logger functions
func Debug(msg string, fields map[string]interface{}) {
	defaultLogger.Debug(msg, fields)
}

func Info(msg string, fields map[string]interface{}) {
	defaultLogger.Info(msg, fields)
}

func Warn(msg string, fields map[string]interface{}) {
	defaultLogger.Warn(msg, fields)
}

func Error(msg string, fields map[string]interface{}) {
	defaultLogger.Error(msg, fields)
}

func Fatal(msg string, fields map[string]interface{}) {
	defaultLogger.Fatal(msg, fields)
}
