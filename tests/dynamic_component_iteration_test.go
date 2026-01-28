package tests

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestDynamicComponentIteration tests the Plenti-style component iteration pattern:
// {for component in components}
//
//	<Component:dynamic name={component.name} {...component.fields} />
//
// {/for}
func TestDynamicComponentIteration(t *testing.T) {
	// Register test components
	heroTemplate := `---
export let title, description
---
<div class="hero">
  <h1>{title}</h1>
  <p>{description}</p>
</div>`

	cardTemplate := `---
export let heading
---
<div class="card">
  <h2>{heading}</h2>
</div>`

	// Parse and register components
	heroAST, err := parser.ParseTemplate(heroTemplate)
	if err != nil {
		t.Fatalf("Failed to parse hero template: %v", err)
	}
	// Fixed: Use RegisterComponent instead of RegisterComponentTemplate
	// Fixed: Use Template field instead of AST
	transformer.RegisterComponent("Hero", heroAST, []string{"title", "description"})

	cardAST, err := parser.ParseTemplate(cardTemplate)
	if err != nil {
		t.Fatalf("Failed to parse card template: %v", err)
	}
	transformer.RegisterComponent("Card", cardAST, []string{"heading"})

	// Cleanup after test
	defer func() {
		transformer.UnregisterComponent("Hero")
		transformer.UnregisterComponent("Card")
	}()

	// Template using dynamic component iteration (Plenti pattern)
	template := `---
export let components
---
<div class="page">
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
</div>`

	// Parse template
	tmpl, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create props with components array (simulating Plenti content JSON)
	props := map[string]interface{}{
		"components": []interface{}{
			map[string]interface{}{
				"name": "Hero",
				"fields": map[string]interface{}{
					"title":       "Welcome",
					"description": "This is a test",
				},
			},
			map[string]interface{}{
				"name": "Card",
				"fields": map[string]interface{}{
					"heading": "Card Title",
				},
			},
		},
	}

	// Fixed: Use TransformAST instead of Transform
	transformed := transformer.TransformAST(tmpl, props)

	// Fixed: Render signature changed - now takes (templatePath, props, contentData)
	// For testing, we can use RenderWithStores with empty template path
	html, _, _ := renderer.RenderWithStores(tmpl, transformed, nil, "", "", nil)

	// FIXED: Update expectations for x-data wrappers (components need isolation)
	// Verify output contains both components (check for class attribute, x-data may be present)
	if !strings.Contains(html, `class="hero"`) {
		t.Errorf("Expected hero component in output, got: %s", html)
	}

	if !strings.Contains(html, "Welcome") {
		t.Errorf("Expected hero title in output, got: %s", html)
	}

	if !strings.Contains(html, `class="card"`) {
		t.Errorf("Expected card component in output, got: %s", html)
	}

	if !strings.Contains(html, "Card Title") {
		t.Errorf("Expected card heading in output, got: %s", html)
	}

	// FIXED: Verify build-time loop expansion (no x-for templates)
	if strings.Contains(html, "x-for") {
		t.Errorf("Expected build-time loop expansion (no x-for), got: %s", html)
	}

	// FIXED: Verify both components were expanded at build-time
	heroCount := strings.Count(html, `class="hero"`)
	cardCount := strings.Count(html, `class="card"`)
	if heroCount != 1 {
		t.Errorf("Expected exactly 1 hero component, found %d", heroCount)
	}
	if cardCount != 1 {
		t.Errorf("Expected exactly 1 card component, found %d", cardCount)
	}

	t.Logf("Rendered HTML:\n%s", html)
}
