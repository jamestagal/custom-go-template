package integration_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRuntimeComponentResolution_EndToEnd(t *testing.T) {
	baseURL := "http://localhost:3333"

	t.Run("Homepage renders with runtime wrappers", func(t *testing.T) {
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

		// Check runtime-components.js included
		if !strings.Contains(html, "/js/runtime-components.js") {
			t.Error("runtime-components.js not found in HTML")
		}

		// Check for runtime wrappers
		if !strings.Contains(html, "dyn-comp-runtime") {
			t.Error("No runtime component wrappers found")
		}

		if !strings.Contains(html, "x-data") {
			t.Error("No x-data attributes found")
		}

		if !strings.Contains(html, "$renderDynamicComponent") {
			t.Error("No $renderDynamicComponent calls found")
		}

		// Log sample output for inspection
		t.Logf("\n=== HTML Sample ===\n%s\n", getSampleOutput(html, "dyn-comp-runtime", 500))
	})

	t.Run("Component registry is accessible", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/js/component-registry.js")
		if err != nil {
			t.Fatalf("Failed to fetch component registry: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		registry := string(body)

		// Check ES module format
		if !strings.HasPrefix(registry, "export default {") {
			t.Error("Registry doesn't start with 'export default {'")
		}

		// Check Hero2436 component exists
		if !strings.Contains(registry, "'Hero2436':") {
			t.Error("Hero2436 component not found in registry")
		}

		// Check Services2437 component exists
		if !strings.Contains(registry, "'Services2437':") {
			t.Error("Services2437 component not found in registry")
		}

		// Log first 1000 chars of registry
		t.Logf("\n=== Component Registry Sample ===\n%s\n", truncate(registry, 1000))
	})

	t.Run("Runtime components script is accessible", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/js/runtime-components.js")
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

		// Check for x-for loop over components array
		if !strings.Contains(html, "x-for=\"(component,") {
			t.Error("x-for loop over components not found")
		}

		if !strings.Contains(html, "in components") {
			t.Error("'in components' clause not found in x-for")
		}

		// Check for hero2436 in data
		if !strings.Contains(html, "hero2436") {
			t.Error("hero2436 component name not found in HTML")
		}

		// Check for services2437 in data
		if !strings.Contains(html, "services2437") {
			t.Error("services2437 component name not found in HTML")
		}

		// Check for components array structure in x-data
		if !strings.Contains(html, "components:[") {
			t.Error("components array not found in x-data")
		}

		// Check for props structure
		if !strings.Contains(html, "compName") {
			t.Error("compName property not found in x-data")
		}

		if !strings.Contains(html, "compProps") {
			t.Error("compProps property not found in x-data")
		}

		// Check for the runtime wrapper div inside the template
		if !strings.Contains(html, "dyn-comp-runtime") {
			t.Error("dyn-comp-runtime class not found")
		}

		// Check for x-init with $renderDynamicComponent
		if !strings.Contains(html, "x-init=\"$renderDynamicComponent") {
			t.Error("x-init with $renderDynamicComponent not found")
		}

		// IMPORTANT: The HTML contains ONE <template x-for> wrapper which will
		// create MULTIPLE runtime wrappers when Alpine.js executes in the browser.
		// We're testing the server-generated HTML, not the runtime result.
		wrapperCount := strings.Count(html, "dyn-comp-runtime")
		if wrapperCount != 1 {
			t.Errorf("Expected 1 x-for template wrapper in HTML, found %d", wrapperCount)
		}

		// Both component names should be in the data
		heroCount := strings.Count(html, "hero2436")
		servicesCount := strings.Count(html, "services2437")

		if heroCount < 1 {
			t.Error("hero2436 should appear at least once in HTML")
		}

		if servicesCount < 1 {
			t.Error("services2437 should appear at least once in HTML")
		}

		t.Logf("\n=== Runtime Wrapper Details ===\nFound %d x-for template wrapper\nComponents in data: hero2436(%d), services2437(%d)\n",
			wrapperCount, heroCount, servicesCount)
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
