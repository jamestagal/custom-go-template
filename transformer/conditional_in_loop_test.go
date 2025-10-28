package transformer

import (
	"log"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// TestConditionalInLoopTransformation tests that conditionals inside loops
// are expanded at build-time when the collection is resolvable
func TestConditionalInLoopTransformation(t *testing.T) {
	log.Println("\n=== TestConditionalInLoopTransformation ===")

	template := `{for animal of animals}
  {if animal == "cat"}
    <div>Hi</div>
  {else}
    <div>Bye</div>
  {/if}
  <div>SIBLING</div>
{/for}`

	log.Printf("\n[TEST] Parsing template:\n%s\n", template)

	// Parse the template
	astTemplate, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	log.Printf("\n[TEST] Template parsed. Root nodes: %d", len(astTemplate.RootNodes))
	for i, node := range astTemplate.RootNodes {
		log.Printf("  [%d] %T", i, node)
		if loop, ok := node.(*ast.Loop); ok {
			log.Printf("      Loop.Content has %d nodes", len(loop.Content))
			for j, cnode := range loop.Content {
				log.Printf("        [%d] %T", j, cnode)
			}
		}
	}

	// Find the loop node
	var loop *ast.Loop
	for _, node := range astTemplate.RootNodes {
		if l, ok := node.(*ast.Loop); ok {
			loop = l
			break
		}
	}

	if loop == nil {
		t.Fatalf("No loop node found in parsed template")
	}

	// Verify the loop has the expected content BEFORE transformation
	log.Printf("\n[TEST] BEFORE transformation, loop.Content has %d nodes:", len(loop.Content))

	conditionalCount := 0
	elementCount := 0
	for i, node := range loop.Content {
		log.Printf("  [%d] %T", i, node)
		if _, ok := node.(*ast.Conditional); ok {
			conditionalCount++
		}
		if elem, ok := node.(*ast.Element); ok {
			elementCount++
			log.Printf("      Element: <%s>", elem.TagName)
		}
	}

	// Before transformation, we expect: Conditional + TextNode + Element (SIBLING div)
	// The exact number depends on whitespace handling, but we should have at least:
	// - 1 Conditional
	// - 1 Element (the SIBLING div)
	if conditionalCount != 1 {
		t.Errorf("Expected 1 Conditional in loop content, got %d", conditionalCount)
	}
	if elementCount != 1 {
		t.Errorf("Expected 1 Element (SIBLING div) in loop content, got %d", elementCount)
	}

	// Now transform the template
	log.Printf("\n[TEST] Transforming template...")
	props := map[string]any{
		"animals": []string{"cat", "dog"},
		"name":    "Test",
	}

	transformedTemplate := TransformAST(astTemplate, props)

	log.Printf("\n[TEST] AFTER transformation, got %d top-level nodes:", len(transformedTemplate.RootNodes))
	for i, node := range transformedTemplate.RootNodes {
		log.Printf("  [%d] %T", i, node)
		if elem, ok := node.(*ast.Element); ok {
			log.Printf("      Element: <%s>", elem.TagName)
			if elem.TagName == "template" {
				// This is a conditional template from build-time expansion
				for _, attr := range elem.Attributes {
					if attr.Name == "x-if" {
						log.Printf("      x-if expression: %s", attr.Value)
					}
				}

				log.Printf("      Template has %d children:", len(elem.Children))
				for j, child := range elem.Children {
					log.Printf("        [%d] %T", j, child)
					if childElem, ok := child.(*ast.Element); ok {
						log.Printf("            Element: <%s>", childElem.TagName)
					}
				}
			}
		}
	}

	// UPDATED: With build-time loop expansion, we expect the loop to be fully expanded
	// With animals=["cat", "dog"], we should get 2 iterations worth of content:
	// Iteration 1 (animal="cat"): conditional + SIBLING div
	// Iteration 2 (animal="dog"): conditional + SIBLING div
	//
	// Each conditional will generate templates (x-if/x-else), so we expect:
	// - Multiple template elements (from conditionals)
	// - Multiple div elements (SIBLING divs from each iteration)

	templateCount := 0
	siblingDivCount := 0

	for _, node := range transformedTemplate.RootNodes {
		if elem, ok := node.(*ast.Element); ok {
			if elem.TagName == "template" {
				// These are conditional templates from build-time expanded iterations
				templateCount++
			}
			if elem.TagName == "div" {
				// Check if this might be a SIBLING div (directly in root, not wrapped)
				siblingDivCount++
			}
		}
	}

	// With 2 iterations, each having a conditional (which produces template elements)
	// and a SIBLING div, we expect at least some templates and divs
	if templateCount == 0 {
		t.Errorf("Expected some template elements from build-time expanded conditionals, got %d", templateCount)
	}

	// We should have SIBLING divs from the loop iterations
	// With 2 iterations, we expect 2 SIBLING divs (unless they're wrapped)
	log.Printf("\n[TEST] Found %d template elements and %d divs in expanded output", templateCount, siblingDivCount)

	// Additional validation: check for "SIBLING" text in the output
	// This validates that the SIBLING div was preserved during expansion
	hasSiblingText := false
	for _, node := range transformedTemplate.RootNodes {
		if elem, ok := node.(*ast.Element); ok {
			if containsSiblingText(elem) {
				hasSiblingText = true
				break
			}
		}
	}

	if !hasSiblingText {
		t.Errorf("BUG: 'SIBLING' text not found in expanded output - content may have been lost")
	}

	log.Printf("\n[TEST] ✅ Build-time loop expansion test complete")
}

// containsSiblingText recursively checks if an element or its children contain "SIBLING" text
func containsSiblingText(elem *ast.Element) bool {
	for _, child := range elem.Children {
		if txt, ok := child.(*ast.TextNode); ok {
			if strings.Contains(txt.Content, "SIBLING") {
				return true
			}
		}
		if childElem, ok := child.(*ast.Element); ok {
			if containsSiblingText(childElem) {
				return true
			}
		}
	}
	return false
}
