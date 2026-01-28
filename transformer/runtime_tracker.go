// Package transformer provides AST transformation utilities.
// This file contains runtime component tracking for conditional registry generation.
package transformer

import "sync"

// Runtime component tracking for conditional registry generation.
// When a runtime component is detected (component name only known at runtime),
// we mark it so the registry generator knows to generate layouts.js.
//
// This eliminates the need to generate the full component registry for pages
// that only use build-time resolved components.

var (
	hasRuntimeComponents bool
	runtimeTrackerMu     sync.RWMutex
)

// MarkRuntimeComponentUsed marks that a runtime component was detected.
// This should be called when emitting a runtime wrapper in dynamic_component_by_name.go.
func MarkRuntimeComponentUsed() {
	runtimeTrackerMu.Lock()
	defer runtimeTrackerMu.Unlock()
	hasRuntimeComponents = true
}

// HasRuntimeComponents returns true if any runtime components were detected.
// This is used by the registry generator to decide whether to generate layouts.js.
func HasRuntimeComponents() bool {
	runtimeTrackerMu.RLock()
	defer runtimeTrackerMu.RUnlock()
	return hasRuntimeComponents
}

// ResetRuntimeComponentTracking resets the runtime component tracking state.
// This should be called at the start of each build to ensure fresh state.
func ResetRuntimeComponentTracking() {
	runtimeTrackerMu.Lock()
	defer runtimeTrackerMu.Unlock()
	hasRuntimeComponents = false
}
