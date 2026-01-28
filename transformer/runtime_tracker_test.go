package transformer

import "testing"

func TestRuntimeTracker(t *testing.T) {
	// Test initial state after reset
	t.Run("InitialState", func(t *testing.T) {
		ResetRuntimeComponentTracking()
		if HasRuntimeComponents() {
			t.Error("Expected HasRuntimeComponents() to be false after reset")
		}
	})

	// Test marking runtime component
	t.Run("MarkRuntimeComponent", func(t *testing.T) {
		ResetRuntimeComponentTracking()
		MarkRuntimeComponentUsed()
		if !HasRuntimeComponents() {
			t.Error("Expected HasRuntimeComponents() to be true after marking")
		}
	})

	// Test multiple marks
	t.Run("MultipleMarks", func(t *testing.T) {
		ResetRuntimeComponentTracking()
		MarkRuntimeComponentUsed()
		MarkRuntimeComponentUsed()
		MarkRuntimeComponentUsed()
		if !HasRuntimeComponents() {
			t.Error("Expected HasRuntimeComponents() to be true after multiple marks")
		}
	})

	// Test reset after mark
	t.Run("ResetAfterMark", func(t *testing.T) {
		ResetRuntimeComponentTracking()
		MarkRuntimeComponentUsed()
		ResetRuntimeComponentTracking()
		if HasRuntimeComponents() {
			t.Error("Expected HasRuntimeComponents() to be false after reset")
		}
	})
}
