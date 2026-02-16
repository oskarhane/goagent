// Package logger provides structured logging and optional OpenTelemetry tracing
// for agent execution monitoring and debugging.
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Level represents the severity level of a log entry.
type Level int

const (
	// LevelDebug is for detailed debugging information.
	LevelDebug Level = iota
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning messages.
	LevelWarn
	// LevelError is for error messages.
	LevelError
)

// String returns the string representation of a log level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) Level {
	switch s {
	case "DEBUG", "debug":
		return LevelDebug
	case "INFO", "info":
		return LevelInfo
	case "WARN", "warn":
		return LevelWarn
	case "ERROR", "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger provides structured logging with configurable output and levels.
type Logger struct {
	level    Level
	output   io.Writer
	mu       sync.Mutex
	tracer   trace.Tracer
	enabled  bool
	tracerOn bool
}

// Config contains configuration for creating a logger.
type Config struct {
	// Level is the minimum log level to output. Defaults to LevelInfo.
	Level Level

	// Output is the destination for log messages. Defaults to os.Stderr.
	Output io.Writer

	// Enabled controls whether logging is active. Defaults to true.
	Enabled bool

	// TracerName enables OpenTelemetry tracing with the given name.
	// If empty, tracing is disabled.
	TracerName string
}

// New creates a new logger with the given configuration.
func New(cfg Config) *Logger {
	output := cfg.Output
	if output == nil {
		output = os.Stderr
	}

	l := &Logger{
		level:    cfg.Level,
		output:   output,
		enabled:  cfg.Enabled,
		tracerOn: cfg.TracerName != "",
	}

	if cfg.TracerName != "" {
		l.tracer = otel.Tracer(cfg.TracerName)
	}

	return l
}

// Default returns a logger with default settings (Info level, stderr output).
func Default() *Logger {
	return New(Config{
		Level:   LevelInfo,
		Output:  os.Stderr,
		Enabled: true,
	})
}

// Noop returns a logger that discards all output.
func Noop() *Logger {
	return New(Config{
		Level:   LevelError,
		Output:  io.Discard,
		Enabled: false,
	})
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Entry represents a structured log entry.
type Entry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// log writes a structured log entry at the given level.
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if !l.enabled || level < l.level {
		return
	}

	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   msg,
		Fields:    fields,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple output if JSON marshaling fails
		_, _ = fmt.Fprintf(l.output, "[%s] %s: %s\n", entry.Timestamp, entry.Level, entry.Message)
		return
	}

	_, _ = l.output.Write(data)
	_, _ = l.output.Write([]byte("\n"))
}

// Debug logs a debug-level message with optional fields.
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(LevelDebug, msg, fields)
}

// Info logs an info-level message with optional fields.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, fields)
}

// Warn logs a warn-level message with optional fields.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, fields)
}

// Error logs an error-level message with optional fields.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(LevelError, msg, fields)
}

// StartSpan starts a new OpenTelemetry span if tracing is enabled.
// Returns the span and a modified context. If tracing is disabled, returns nil span.
func (l *Logger) StartSpan(
	ctx context.Context, name string, attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	if !l.tracerOn || l.tracer == nil {
		return ctx, nil
	}

	ctx, span := l.tracer.Start(ctx, name)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	return ctx, span
}

// EndSpan ends a span with optional error recording.
func (l *Logger) EndSpan(span trace.Span, err error) {
	if span == nil {
		return
	}

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
}

// RecordSpanEvent records an event on the given span.
func (l *Logger) RecordSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if span == nil {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}
