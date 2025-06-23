package png

import (
	"fmt"
	"strings"
	"testing"
)

// TestLogger is a test implementation of the Logger interface
type TestLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
}

func (l *TestLogger) Debug(format string, args ...interface{}) {
	l.DebugMessages = append(l.DebugMessages, fmt.Sprintf(format, args...))
}

func (l *TestLogger) Info(format string, args ...interface{}) {
	l.InfoMessages = append(l.InfoMessages, fmt.Sprintf(format, args...))
}

func (l *TestLogger) Warn(format string, args ...interface{}) {
	l.WarnMessages = append(l.WarnMessages, fmt.Sprintf(format, args...))
}

func (l *TestLogger) Error(format string, args ...interface{}) {
	l.ErrorMessages = append(l.ErrorMessages, fmt.Sprintf(format, args...))
}

func TestSetLogger(t *testing.T) {
	// Save original logger
	originalLogger := logger
	defer func() { logger = originalLogger }()

	// Test with nil logger (should not panic)
	SetLogger(nil)

	// These should not panic even with nil logger
	logDebug("debug message")
	logInfo("info message")
	logWarn("warn message")
	logError("error message")

	// Test with custom logger
	testLogger := &TestLogger{}
	SetLogger(testLogger)

	// Test logging functions
	logDebug("test debug %d", 1)
	logInfo("test info %s", "message")
	logWarn("test warn %.2f", 3.14)
	logError("test error %v", fmt.Errorf("error"))

	// Verify messages were captured
	if len(testLogger.DebugMessages) != 1 || testLogger.DebugMessages[0] != "test debug 1" {
		t.Errorf("Debug message not captured correctly: %v", testLogger.DebugMessages)
	}
	if len(testLogger.InfoMessages) != 1 || testLogger.InfoMessages[0] != "test info message" {
		t.Errorf("Info message not captured correctly: %v", testLogger.InfoMessages)
	}
	if len(testLogger.WarnMessages) != 1 || testLogger.WarnMessages[0] != "test warn 3.14" {
		t.Errorf("Warn message not captured correctly: %v", testLogger.WarnMessages)
	}
	if len(testLogger.ErrorMessages) != 1 || testLogger.ErrorMessages[0] != "test error error" {
		t.Errorf("Error message not captured correctly: %v", testLogger.ErrorMessages)
	}
}

func TestOptimizeWithLogger(t *testing.T) {
	// Save original logger
	originalLogger := logger
	defer func() { logger = originalLogger }()

	testLogger := &TestLogger{}
	SetLogger(testLogger)

	// Run optimization on a regular PNG file
	srcPath := "./testdata/optimize/me2020.png"
	destPath := "./testdata/temp/optimized_test.png"
	quality := ""

	optimizer := NewOptimizer(quality)
	optimizer.SetLogger(testLogger)
	_, err := optimizer.Run(srcPath, destPath)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	// Verify logger captured the messages
	hasStartMessage := false
	hasSizeMessage := false

	for _, msg := range testLogger.InfoMessages {
		if strings.Contains(msg, "Starting PNG optimization") {
			hasStartMessage = true
		}
		if strings.Contains(msg, "Optimization completed") || strings.Contains(msg, "Cannot optimize") {
			// Check for human-readable bytes (e.g., "1.2 MB -> 900 kB")
			if strings.Contains(msg, "kB") || strings.Contains(msg, "MB") || strings.Contains(msg, "B") {
				hasSizeMessage = true
			}
		}
	}

	if !hasStartMessage {
		t.Error("Expected 'Starting PNG optimization' message")
		t.Logf("Info messages: %v", testLogger.InfoMessages)
	}

	// Check debug messages for human-readable sizes
	hasDebugSizeMessage := false
	for _, msg := range testLogger.DebugMessages {
		if strings.Contains(msg, "size:") && (strings.Contains(msg, "kB") || strings.Contains(msg, "MB") || strings.Contains(msg, "B")) {
			hasDebugSizeMessage = true
			break
		}
	}

	if !hasDebugSizeMessage && !hasSizeMessage {
		t.Error("Expected human-readable size in messages")
		t.Logf("Debug messages: %v", testLogger.DebugMessages)
		t.Logf("Info messages: %v", testLogger.InfoMessages)
	}
}
