package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestMagicVariables_ComponentsArrayPassedToTemplate tests that components array is accessible
func TestMagicVariables_ComponentsArrayPassedToTemplate(t *testing.T) {
	templateContent := `---
export let components
---

<div>
	{if components}
		<p>Components array exists</p>
	{/if}
</div>`

	// Parse template
	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Build props with components magic variable
	props := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"name": "Hero2436",
				"fields": map[string]interface{}{
					"title": "Welcome",
				},
			},
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: RenderWithStores signature is (originalAST, transformedAST, storeDefinitions, templatePath, dynamicLayoutName)
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify components array is accessible
	if !strings.Contains(markup, "Components array exists") {
		t.Error("Expected components array to be accessible in template")
		t.Logf("Markup: %s", markup)
	}
}

// TestMagicVariables_ContentObjectPassedToTemplate tests content object access
func TestMagicVariables_ContentObjectPassedToTemplate(t *testing.T) {
	templateContent := `---
export let content
---

<div>
	{if content}
		<h1>{content.title}</h1>
	{/if}
</div>`

	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Build props with content magic variable
	props := map[string]interface{}{
		"content": map[string]interface{}{
			"title": "Test Page",
			"description": "Description",
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Remove extra parameter
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify content object is accessible
	if !strings.Contains(markup, "x-text=\"content.title\"") {
		t.Error("Expected content.title to be accessible")
		t.Logf("Markup: %s", markup)
	}
}

// TestMagicVariables_AllContentMapPassedToTemplate tests allContent map access
func TestMagicVariables_AllContentMapPassedToTemplate(t *testing.T) {
	templateContent := `---
export let allContent
---

<div>
	{if allContent}
		<p>All content map exists</p>
	{/if}
</div>`

	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Build props with allContent magic variable
	props := map[string]interface{}{
		"allContent": map[string]interface{}{
			"pages/_index": map[string]interface{}{
				"title": "Home",
			},
			"pages/about": map[string]interface{}{
				"title": "About",
			},
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Remove extra parameter
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify allContent map is accessible
	if !strings.Contains(markup, "All content map exists") {
		t.Error("Expected allContent map to be accessible")
		t.Logf("Markup: %s", markup)
	}
}

// TestMagicVariables_AllLayoutsRegistryPassedToTemplate tests allLayouts access
func TestMagicVariables_AllLayoutsRegistryPassedToTemplate(t *testing.T) {
	templateContent := `---
export let allLayouts
---

<div>
	{if allLayouts}
		<p>All layouts registry exists</p>
	{/if}
</div>`

	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Build props with allLayouts magic variable
	props := map[string]interface{}{
		"allLayouts": map[string]bool{
			"Hero2436": true,
			"Footer": true,
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Remove extra parameter
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify allLayouts registry is accessible
	if !strings.Contains(markup, "All layouts registry exists") {
		t.Error("Expected allLayouts registry to be accessible")
		t.Logf("Markup: %s", markup)
	}
}

// TestMagicVariables_AllFourInLoopContext tests magic variables in Component:dynamic loop
func TestMagicVariables_AllFourInLoopContext(t *testing.T) {
	// First, register a test component
	componentContent := `---
export let title, content, allContent
---

<div class="test-component">
	<h2>{title}</h2>
	{if content}
		<p>Content: {content.description}</p>
	{/if}
	{if allContent}
		<p>Has all content</p>
	{/if}
</div>`

	componentAST, err := parser.ParseTemplate(componentContent)
	if err != nil {
		t.Fatalf("Failed to parse component: %v", err)
	}

	// Register the component
	transformer.RegisterComponent("TestComponent", componentAST, []string{"title", "content", "allContent"})

	// Cleanup after test
	defer transformer.UnregisterComponent("TestComponent")

	// Template with Component:dynamic in loop
	templateContent := `---
export let components, content, allContent, allLayouts
---

<div>
	{for component in components}
		<Component:dynamic name={component.name} {...component.fields} content={content} allContent={allContent} />
	{/for}
</div>`

	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Build props with all magic variables
	props := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"name": "TestComponent",
				"fields": map[string]interface{}{
					"title": "Test Title",
				},
			},
		},
		"content": map[string]interface{}{
			"description": "Page description",
		},
		"allContent": map[string]interface{}{
			"pages/_index": map[string]interface{}{
				"title": "Home",
			},
		},
		"allLayouts": map[string]bool{
			"TestComponent": true,
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Remove extra parameter
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify component rendered with magic variables
	if !strings.Contains(markup, "test-component") {
		t.Error("Expected component to render")
		t.Logf("Markup: %s", markup)
	}

	if !strings.Contains(markup, "Test Title") {
		t.Error("Expected title from component.fields")
		t.Logf("Markup: %s", markup)
	}

	// UPDATED: Check for runtime conditional structure with x-if
	// Magic variables are passed as Alpine variable references (content: content)
	// so conditionals are evaluated at runtime, not build-time
	if !strings.Contains(markup, `x-if="content"`) {
		t.Error("Expected content conditional to be present")
		t.Logf("Markup: %s", markup)
	}

	if !strings.Contains(markup, `x-if="allContent"`) {
		t.Error("Expected allContent conditional to be present")
		t.Logf("Markup: %s", markup)
	}

	// Verify the conditional content structure is present
	if !strings.Contains(markup, "content.description") {
		t.Error("Expected content.description reference in conditional")
		t.Logf("Markup: %s", markup)
	}
}

// TestMagicVariables_ComponentsLoopWithMultipleComponents tests iteration over multiple components
func TestMagicVariables_ComponentsLoopWithMultipleComponents(t *testing.T) {
	// Register test components
	comp1Content := `---
export let name
---
<div class="comp1">{name}</div>`

	comp2Content := `---
export let title
---
<div class="comp2">{title}</div>`

	comp1AST, _ := parser.ParseTemplate(comp1Content)
	comp2AST, _ := parser.ParseTemplate(comp2Content)

	transformer.RegisterComponent("Comp1", comp1AST, []string{"name"})
	transformer.RegisterComponent("Comp2", comp2AST, []string{"title"})

	// Cleanup after test
	defer func() {
		transformer.UnregisterComponent("Comp1")
		transformer.UnregisterComponent("Comp2")
	}()

	// Template
	templateContent := `---
export let components
---

<div>
	{for component in components}
		<Component:dynamic name={component.name} {...component.fields} />
	{/for}
</div>`

	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Props with multiple components
	props := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"name": "Comp1",
				"fields": map[string]interface{}{
					"name": "First",
				},
			},
			map[string]interface{}{
				"name": "Comp2",
				"fields": map[string]interface{}{
					"title": "Second",
				},
			},
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Remove extra parameter
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify both components rendered
	if !strings.Contains(markup, "comp1") {
		t.Error("Expected first component to render")
	}

	if !strings.Contains(markup, "comp2") {
		t.Error("Expected second component to render")
	}

	if !strings.Contains(markup, "First") {
		t.Error("Expected first component's data")
	}

	if !strings.Contains(markup, "Second") {
		t.Error("Expected second component's data")
	}
}

// TestMagicVariables_IntegrationWithRealFiles tests magic variables with actual file system
func TestMagicVariables_IntegrationWithRealFiles(t *testing.T) {
	// This test verifies that the magic variables system works with real component files
	// Skip if we're not in the project root
	if _, err := os.Stat("layouts/components"); os.IsNotExist(err) {
		t.Skip("Skipping integration test - not in project root")
	}

	// Check if Hero2436 component exists
	hero2436Path := filepath.Join("layouts", "components", "Hero2436.html")
	if _, err := os.Stat(hero2436Path); os.IsNotExist(err) {
		t.Skip("Skipping - Hero2436 component not found")
	}

	// Read and parse Hero2436 component
	heroContent, err := os.ReadFile(hero2436Path)
	if err != nil {
		t.Fatalf("Failed to read Hero2436 component: %v", err)
	}

	heroAST, err := parser.ParseTemplate(string(heroContent))
	if err != nil {
		t.Fatalf("Failed to parse Hero2436 component: %v", err)
	}

	// Register the component
	transformer.RegisterComponent("Hero2436", heroAST, []string{"topper", "title", "description", "buttonText", "buttonLink"})

	// Cleanup after test
	defer transformer.UnregisterComponent("Hero2436")

	// Template using Component:dynamic
	templateContent := `---
export let components, content
---

<!DOCTYPE html>
<html>
<body>
	{for component in components}
		<Component:dynamic name={component.name} {...component.fields} content={content} />
	{/for}
</body>
</html>`

	tmpl, err := parser.ParseTemplate(templateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Props matching real content JSON structure
	props := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"name": "Hero2436",
				"fields": map[string]interface{}{
					"topper": "Welcome",
					"title": "Test Title",
					"description": "Test description",
					"buttonText": "Click",
					"buttonLink": "/test",
				},
			},
		},
		"content": map[string]interface{}{
			"title": "Page Title",
		},
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Remove extra parameter
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "")

	// Verify Hero2436 rendered with props
	if !strings.Contains(markup, "hero-2436") {
		t.Error("Expected Hero2436 component to render (id: hero-2436)")
		t.Logf("Markup (first 500 chars): %s", markup[:min(500, len(markup))])
	}

	if !strings.Contains(markup, "Welcome") {
		t.Error("Expected topper 'Welcome' from component fields")
	}

	if !strings.Contains(markup, "Test Title") {
		t.Error("Expected title 'Test Title' from component fields")
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
