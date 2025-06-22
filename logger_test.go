package png

import (
	"fmt"
	"strings"
	"testing"
)

// TestLogger はテスト用のLoggerインターフェース実装です。
type TestLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
}

func (tl *TestLogger) Debug(format string, args ...interface{}) {
	tl.DebugMessages = append(tl.DebugMessages, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Info(format string, args ...interface{}) {
	tl.InfoMessages = append(tl.InfoMessages, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Warn(format string, args ...interface{}) {
	tl.WarnMessages = append(tl.WarnMessages, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Error(format string, args ...interface{}) {
	tl.ErrorMessages = append(tl.ErrorMessages, fmt.Sprintf(format, args...))
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

	// Log some messages
	logDebug("debug message %d", 1)
	logInfo("info message %s", "test")
	logWarn("warn message %v", true)
	logError("error message %f", 3.14)

	// Verify messages were logged
	if len(testLogger.DebugMessages) != 1 || testLogger.DebugMessages[0] != "debug message 1" {
		t.Errorf("Expected debug message 'debug message 1', got %v", testLogger.DebugMessages)
	}
	if len(testLogger.InfoMessages) != 1 || testLogger.InfoMessages[0] != "info message test" {
		t.Errorf("Expected info message 'info message test', got %v", testLogger.InfoMessages)
	}
	if len(testLogger.WarnMessages) != 1 || testLogger.WarnMessages[0] != "warn message true" {
		t.Errorf("Expected warn message 'warn message true', got %v", testLogger.WarnMessages)
	}
	if len(testLogger.ErrorMessages) != 1 || testLogger.ErrorMessages[0] != "error message 3.140000" {
		t.Errorf("Expected error message 'error message 3.140000', got %v", testLogger.ErrorMessages)
	}
}

func TestOptimizeWithLogger(t *testing.T) {
	// Save original logger
	originalLogger := logger
	defer func() { logger = originalLogger }()

	testLogger := &TestLogger{}
	SetLogger(testLogger)

	// Run optimization on a regular PNG file
	input := OptimizePngInput{
		SrcPath:  "./testdata/optimize/me2020.png",
		DestPath: "./testdata/temp/optimized_test.png",
		Quality:  "",
	}

	_, err := Optimize(input)
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