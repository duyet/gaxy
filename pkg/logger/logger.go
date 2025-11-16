package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the log level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is a structured logger
type Logger struct {
	mu     sync.Mutex
	level  Level
	format string // "json" or "text"
	output io.Writer
	fields map[string]interface{}
}

// New creates a new logger
func New(levelStr, format string) *Logger {
	level := parseLevel(levelStr)
	return &Logger{
		level:  level,
		format: format,
		output: os.Stdout,
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// ErrorWithErr logs an error with an error object
func (l *Logger) ErrorWithErr(msg string, err error) {
	fields := make(map[string]interface{})
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(ErrorLevel, msg, fields)
}

// log is the internal logging function
func (l *Logger) log(level Level, msg string, extraFields map[string]interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Merge fields
	fields := make(map[string]interface{})
	for k, v := range l.fields {
		fields[k] = v
	}
	for k, v := range extraFields {
		fields[k] = v
	}

	// Add standard fields
	fields["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	fields["level"] = level.String()
	fields["message"] = msg

	var output string
	if l.format == "json" {
		data, _ := json.Marshal(fields)
		output = string(data)
	} else {
		// Text format
		output = fmt.Sprintf("[%s] %s: %s", fields["timestamp"], fields["level"], msg)

		// Add extra fields
		for k, v := range fields {
			if k != "timestamp" && k != "level" && k != "message" {
				output += fmt.Sprintf(" %s=%v", k, v)
			}
		}
	}

	fmt.Fprintln(l.output, output)
}

// parseLevel parses a log level string
func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
