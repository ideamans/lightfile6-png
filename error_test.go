package png

import (
	"errors"
	"fmt"
	"testing"
)

func returnError(isDataError bool) error {
	if isDataError {
		return NewDataError("test error")
	}
	return errors.New("test error")
}

func TestDataError(t *testing.T) {
	err := returnError(false)
	dataErr := returnError(true)

	if _, ok := dataErr.(*DataError); !ok {
		t.Errorf("NewDataError should return a *DataError")
	}

	if _, ok := err.(*DataError); ok {
		t.Errorf("err should not be a *DataError")
	}

	if AsDataError(dataErr) == nil {
		t.Errorf("AsDataError should return a *DataError")
	}

	if AsDataError(err) != nil {
		t.Errorf("AsDataError should return nil")
	}

	// Test AsDataError with wrapped DataError using %w
	wrappedDataErr := fmt.Errorf("wrapper: %w", dataErr)
	if AsDataError(wrappedDataErr) == nil {
		t.Error("AsDataError should find DataError in error chain")
	}

	// Test multiple levels of wrapping
	doubleWrappedErr := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", dataErr))
	if AsDataError(doubleWrappedErr) == nil {
		t.Error("AsDataError should find DataError in deeply nested error chain")
	}

	// Test that errors.Is works correctly with DataError
	if !errors.Is(wrappedDataErr, dataErr) {
		t.Error("errors.Is should work with wrapped DataError")
	}

	// Test wrapped regular error should not be detected as DataError
	wrappedRegularErr := fmt.Errorf("wrapper: %w", err)
	if AsDataError(wrappedRegularErr) != nil {
		t.Error("AsDataError should return nil for wrapped regular error")
	}
}
