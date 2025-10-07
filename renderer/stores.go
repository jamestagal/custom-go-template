package renderer

import (
	"sort"
	"strings"
)

// renderStoreInitializations generates Alpine.js store initialization script
// Input: stores map where key=storeName, value=JS object literal
// Output: <script> tag with Alpine.store() calls wrapped in alpine:init listener
//
// Example:
//   Input:  {"auth": "{ isLoggedIn: false }", "cart": "{ items: [] }"}
//   Output: <script>
//           document.addEventListener('alpine:init', () => {
//               Alpine.store('auth', { isLoggedIn: false });
//               Alpine.store('cart', { items: [] });
//           });
//           </script>
//
// Cognitive Load: 8
// - Empty check: 1
// - String builder: 1
// - Sort for determinism: 2
// - Loop + append: 2
// - String formatting: 2
func renderStoreInitializations(stores map[string]string) string {
	// Handle empty stores - return empty string
	if len(stores) == 0 {
		return ""
	}

	var sb strings.Builder

	// Start script tag
	sb.WriteString("<script>\n")

	// Alpine.js initialization event listener
	sb.WriteString("document.addEventListener('alpine:init', () => {\n")

	// Sort store names for deterministic output
	// This is important for testing and reproducible builds
	storeNames := make([]string, 0, len(stores))
	for name := range stores {
		storeNames = append(storeNames, name)
	}
	sort.Strings(storeNames)

	// Generate Alpine.store() call for each store
	// Indented with 4 spaces to be inside the event listener
	for _, name := range storeNames {
		definition := stores[name]
		// The definition is already a valid JS object literal
		// No need to parse or escape - just embed it
		sb.WriteString("    Alpine.store('")
		sb.WriteString(name)
		sb.WriteString("', ")
		sb.WriteString(definition)
		sb.WriteString(");\n")
	}

	// Close event listener
	sb.WriteString("});\n")

	// Close script tag
	sb.WriteString("</script>")

	return sb.String()
}
