package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{
		Level:   LevelInfo,
		Output:  &buf,
		Enabled: true,
	})

	assert.NotNil(t, log)
}

func TestDefault(t *testing.T) {
	log := Default()
	assert.NotNil(t, log)
}

func TestNoop(t *testing.T) {
	log := Noop()
	assert.NotNil(t, log)
}

func TestLogger_Logging(t *testing.T) {
	tests := []struct {
		name      string
		level     Level
		logFunc   func(*Logger, string, map[string]any)
		message   string
		fields    map[string]any
		shouldLog bool
	}{
		{
			name:      "debug logs at debug level",
			level:     LevelDebug,
			logFunc:   (*Logger).Debug,
			message:   "debug message",
			fields:    map[string]any{"key": "value"},
			shouldLog: true,
		},
		{
			name:      "debug does not log at info level",
			level:     LevelInfo,
			logFunc:   (*Logger).Debug,
			message:   "debug message",
			fields:    nil,
			shouldLog: false,
		},
		{
			name:      "info logs at info level",
			level:     LevelInfo,
			logFunc:   (*Logger).Info,
			message:   "info message",
			fields:    nil,
			shouldLog: true,
		},
		{
			name:      "warn logs at warn level",
			level:     LevelWarn,
			logFunc:   (*Logger).Warn,
			message:   "warn message",
			fields:    nil,
			shouldLog: true,
		},
		{
			name:      "error logs at error level",
			level:     LevelError,
			logFunc:   (*Logger).Error,
			message:   "error message",
			fields:    map[string]any{"error": "something went wrong"},
			shouldLog: true,
		},
		{
			name:      "info does not log at error level",
			level:     LevelError,
			logFunc:   (*Logger).Info,
			message:   "info message",
			fields:    nil,
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(Config{
				Level:   tt.level,
				Output:  &buf,
				Enabled: true,
			})

			tt.logFunc(log, tt.message, tt.fields)

			output := buf.String()
			if tt.shouldLog {
				assert.NotEmpty(t, output)
				assert.Contains(t, output, tt.message)

				// Verify JSON structure
				var logEntry map[string]any
				err := json.Unmarshal([]byte(output), &logEntry)
				require.NoError(t, err)

				assert.Contains(t, logEntry, "timestamp")
				assert.Contains(t, logEntry, "level")
				assert.Contains(t, logEntry, "message")
				assert.Equal(t, tt.message, logEntry["message"])

				// Check fields if provided
				if tt.fields != nil {
					fields, ok := logEntry["fields"].(map[string]any)
					require.True(t, ok)
					for k, v := range tt.fields {
						assert.Equal(t, v, fields[k])
					}
				}
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_NoopDoesNotLog(t *testing.T) {
	var buf bytes.Buffer
	// Noop logger uses io.Discard and Enabled=false, so even with custom output it won't log
	log := Noop()

	log.Debug("debug", nil)
	log.Info("info", nil)
	log.Warn("warn", nil)
	log.Error("error", nil)

	// Verify noop logger doesn't write to provided buffer
	log2 := New(Config{
		Level:   LevelDebug,
		Output:  &buf,
		Enabled: false, // Explicitly disabled
	})
	log2.Info("test", nil)
	assert.Empty(t, buf.String())
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{
		Level:   LevelWarn,
		Output:  &buf,
		Enabled: true,
	})

	// Debug and Info should not log
	log.Debug("debug", nil)
	log.Info("info", nil)
	assert.Empty(t, buf.String())

	// Warn should log
	log.Warn("warn", nil)
	assert.Contains(t, buf.String(), "warn")

	buf.Reset()

	// Error should log
	log.Error("error", nil)
	assert.Contains(t, buf.String(), "error")
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{
		Level:   LevelInfo,
		Output:  &buf,
		Enabled: true,
	})

	fields := map[string]any{
		"request_id": "123",
		"duration":   42,
		"success":    true,
	}

	log.Info("test message", fields)

	output := buf.String()
	var logEntry map[string]any
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	// Verify required fields
	assert.Contains(t, logEntry, "timestamp")
	assert.Contains(t, logEntry, "level")
	assert.Contains(t, logEntry, "message")
	assert.Equal(t, "test message", logEntry["message"])

	// Verify custom fields
	logFields, ok := logEntry["fields"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "123", logFields["request_id"])
	assert.Equal(t, float64(42), logFields["duration"])
	assert.Equal(t, true, logFields["success"])
}

func TestLogger_ThreadSafety(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{
		Level:   LevelInfo,
		Output:  &buf,
		Enabled: true,
	})

	// Run concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			log.Info("concurrent message", map[string]any{"id": id})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 log entries
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 10, len(lines))

	// Each line should be valid JSON
	for _, line := range lines {
		var logEntry map[string]any
		err := json.Unmarshal([]byte(line), &logEntry)
		assert.NoError(t, err)
	}
}
