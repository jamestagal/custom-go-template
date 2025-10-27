package transformer

import "log"

// OptimizeXData controls x-data optimization feature
// Set to false to revert to legacy behavior
//
// Phase 1: Removes redundant root-level x-data wrapper (body already provides scope)
// Phase 2: Smart scope diffing to only add x-data when components introduce new variables
//
// Default: true (optimization enabled)
var OptimizeXData = true

// SetOptimization allows runtime control (useful for testing)
// This function provides a clean API for toggling the optimization flag
//
// Example:
//   SetOptimization(false) // Revert to legacy behavior
//   SetOptimization(true)  // Enable optimization
func SetOptimization(enabled bool) {
	OptimizeXData = enabled
	status := "DISABLED"
	if enabled {
		status = "ENABLED"
	}
	log.Printf("[X-Data Config] Optimization %s", status)
}
