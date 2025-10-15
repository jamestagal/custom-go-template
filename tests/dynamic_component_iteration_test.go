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
//   <Component:dynamic name={component.name} {...component.fields} />
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
	transformer.RegisterComponentTemplate("Hero", &transformer.ComponentTemplate{
		Name: "Hero",
		AST:  heroAST,
	})

	cardAST, err := parser.ParseTemplate(cardTemplate)
	if err != nil {
		t.Fatalf("Failed to parse card template: %v", err)
	}
	transformer.RegisterComponentTemplate("Card", &transformer.ComponentTemplate{
		Name: "Card",
		AST:  cardAST,
	})

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
	ast, err := parser.ParseTemplate(template)
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

	// Transform
	transformed := transformer.Transform(ast, props)

	// Render
	html, _, _ := renderer.Render(transformed, props)

	// Verify output contains both components
	if !strings.Contains(html, `<div class="hero">`) {
		t.Errorf("Expected hero component in output, got: %s", html)
	}

	if !strings.Contains(html, "<h1>Welcome</h1>") {
		t.Errorf("Expected hero title in output, got: %s", html)
	}

	if !strings.Contains(html, `<div class="card">`) {
		t.Errorf("Expected card component in output, got: %s", html)
	}

	if !strings.Contains(html, "<h2>Card Title</h2>") {
		t.Errorf("Expected card heading in output, got: %s", html)
	}

	t.Logf("Rendered HTML:\n%s", html)
}
