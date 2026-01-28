package integration_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRuntimeComponentResolution_EndToEnd(t *testing.T) {
	baseURL := "http://localhost:3333"

	// Skip if server is not running
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		t.Skip("Skipping E2E test: server not running on port 3333. Start server with: go run cmd/server/main.go")
	}
	resp.Body.Close()

	t.Run("Homepage renders with build-time expanded components", func(t *testing.T) {
		// Updated 2025-01-25: Components now expand at build-time instead of using runtime wrappers
		// This is the correct Plenti/Svelte-like behavior for static content
		resp, err := http.Get(baseURL + "/")
		if err != nil {
			t.Fatalf("Failed to fetch homepage: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		html := string(body)

		// Check Alpine.js included
		if !strings.Contains(html, "alpinejs") {
			t.Error("Alpine.js CDN not found in HTML")
		}

		// Check runtime-components.js included (still needed for dynamic cases)
		// Path is now /core/runtime-components.js for Plenti compatibility
		if !strings.Contains(html, "/core/runtime-components.js") && !strings.Contains(html, "/js/runtime-components.js") {
			t.Error("runtime-components.js not found in HTML (checked /core/ and /js/)")
		}

		// With build-time expansion, components are inlined directly
		// No runtime wrappers should be present for static content
		if strings.Contains(html, "dyn-comp-runtime") {
			t.Logf("Note: Found runtime wrappers - some components may require runtime resolution")
		}

		if !strings.Contains(html, "x-data") {
			t.Error("No x-data attributes found")
		}

		// Log sample output for inspection
		t.Logf("\n=== HTML Sample ===\n%s\n", getSampleOutput(html, "x-data", 500))
	})

	t.Run("Component registry is accessible", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/generated/layouts.js")
		if err != nil {
			t.Fatalf("Failed to fetch component registry: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		registry := string(body)

		// Check ES module format - registry uses const registry = { ... } and export default registry
		if !strings.Contains(registry, "const registry = {") {
			t.Error("Registry doesn't contain 'const registry = {'")
		}
		if !strings.Contains(registry, "export default registry") {
			t.Error("Registry doesn't contain 'export default registry'")
		}

		// Check that registry contains component entries (using relative path format)
		// Components are registered with relative paths like '../components/userprofile.html'
		if !strings.Contains(registry, "(props) =>") {
			t.Error("No component template functions found in registry")
		}

		// Check registry has some content
		if len(registry) < 100 {
			t.Error("Registry appears to be empty or too small")
		}

		// Log first 1000 chars of registry
		t.Logf("\n=== Component Registry Sample ===\n%s\n", truncate(registry, 1000))
	})

	t.Run("Runtime components script is accessible", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/core/runtime-components.js")
		if err != nil {
			t.Fatalf("Failed to fetch runtime components: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		script := string(body)

		// Check Alpine magic function
		if !strings.Contains(script, "Alpine.magic('renderDynamicComponent'") {
			t.Error("Alpine magic function not found")
		}

		// Check registry loading
		if !strings.Contains(script, "loadComponentRegistry") {
			t.Error("loadComponentRegistry function not found")
		}

		// Log first 1000 chars of script
		t.Logf("\n=== Runtime Components Script Sample ===\n%s\n", truncate(script, 1000))
	})

	t.Run("Runtime wrappers contain correct data", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/")
		if err != nil {
			t.Fatalf("Failed to fetch homepage: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		html := string(body)

		// UPDATED EXPECTATION: The homepage (_index.html) uses STATIC imports:
		//   import Hero2436 from '../components/hero2436.html'
		//   <Hero2436 ... />
		//
		// This is NOT dynamic component resolution! Static imports are inlined at build-time.
		// With build-time expansion, there's no x-for loop or runtime wrappers.
		//
		// For runtime dynamic components, a template would need:
		//   {for component in components}
		//     <Component:dynamic name={component.name} {...component.fields} />
		//   {/for}
		//
		// The homepage doesn't use this pattern, so we check for inlined components instead.

		// Check for hero2436 component content (inlined via id="hero-2436")
		// The component uses id="hero-2436" (with hyphen) in the rendered HTML
		if !strings.Contains(html, "hero-2436") {
			t.Error("hero-2436 component ID not found in HTML")
		}

		// Check for services2437 component content (inlined via id="services-2437")
		// The component uses id="services-2437" (with hyphen) in the rendered HTML
		if !strings.Contains(html, "services-2437") {
			t.Error("services-2437 component ID not found in HTML")
		}

		// Log component occurrences
		heroCount := strings.Count(html, "hero-2436")
		servicesCount := strings.Count(html, "services-2437")

		t.Logf("\n=== Static Component Inlining ===\nComponents are statically imported and inlined (not runtime resolved)\nhero-2436 appears: %d times\nservices-2437 appears: %d times\n",
			heroCount, servicesCount)

		// NOTE: To test actual runtime component resolution, create a separate test page
		// that uses the {for component in components}<Component:dynamic .../>{/for} pattern
		// with content that cannot be resolved at build-time (e.g., using Alpine stores).
	})
}

// Helper function to get sample output around a search term
func getSampleOutput(content, searchTerm string, contextSize int) string {
	idx := strings.Index(content, searchTerm)
	if idx == -1 {
		return "[Search term not found]"
	}

	start := idx - contextSize/2
	if start < 0 {
		start = 0
	}

	end := idx + contextSize/2
	if end > len(content) {
		end = len(content)
	}

	return content[start:end]
}

// Helper function to truncate string
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
