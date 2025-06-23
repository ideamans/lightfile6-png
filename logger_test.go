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

func TestOptimizeWithLogger(t *testing.T) {
	testLogger := &TestLogger{}

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