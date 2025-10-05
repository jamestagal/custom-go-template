package transformer

import (
	"log"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// TestConditionalInLoopTransformation tests that conditionals inside loops
// are transformed correctly without losing sibling content
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
				// This should be the x-for template
				for _, attr := range elem.Attributes {
					if attr.Name == "x-for" {
						log.Printf("      x-for expression: %s", attr.Value)
					}
				}

				log.Printf("      Template has %d children:", len(elem.Children))
				for j, child := range elem.Children {
					log.Printf("        [%d] %T", j, child)
					if childElem, ok := child.(*ast.Element); ok {
						log.Printf("            Element: <%s>", childElem.TagName)

						// Check if this is the wrapper div
						if childElem.TagName == "div" && len(childElem.Children) > 0 {
							log.Printf("            Wrapper div has %d children:", len(childElem.Children))
							for k, grandchild := range childElem.Children {
								log.Printf("              [%d] %T", k, grandchild)
								if gcElem, ok := grandchild.(*ast.Element); ok {
									log.Printf("                  Element: <%s>", gcElem.TagName)
								}
							}
						}
					}
				}
			}
		}
	}

	// Verify the transformation produced correct structure
	// We should have a <template x-for> element
	var xForTemplate *ast.Element
	for _, node := range transformedTemplate.RootNodes {
		if elem, ok := node.(*ast.Element); ok {
			if elem.TagName == "template" {
				for _, attr := range elem.Attributes {
					if attr.Name == "x-for" {
						xForTemplate = elem
						break
					}
				}
			}
		}
	}

	if xForTemplate == nil {
		t.Fatalf("No <template x-for> element found in transformed output")
	}

	// The x-for template should have children
	if len(xForTemplate.Children) == 0 {
		t.Fatalf("x-for template has no children")
	}

	// Since there are multiple nodes (conditional + div), they should be wrapped in a div
	// Check the first child - it should be a wrapper div
	wrapperDiv, ok := xForTemplate.Children[0].(*ast.Element)
	if !ok {
		t.Fatalf("Expected wrapper div as first child of x-for template, got %T", xForTemplate.Children[0])
	}

	if wrapperDiv.TagName != "div" {
		t.Fatalf("Expected wrapper to be a div, got <%s>", wrapperDiv.TagName)
	}

	log.Printf("\n[TEST] Checking wrapper div children...")

	// The wrapper div should contain:
	// 1. The transformed conditional (a <template x-if>)
	// 2. The SIBLING div

	hasConditionalTemplate := false
	hasSiblingDiv := false

	for i, child := range wrapperDiv.Children {
		log.Printf("  [%d] %T", i, child)
		if elem, ok := child.(*ast.Element); ok {
			log.Printf("      Element: <%s>", elem.TagName)

			if elem.TagName == "template" {
				// Check if this is an x-if template
				for _, attr := range elem.Attributes {
					if attr.Name == "x-if" || attr.Name == "x-else" {
						hasConditionalTemplate = true
						log.Printf("      Found conditional template with %s", attr.Name)
					}
				}
			}

			if elem.TagName == "div" {
				// Check if this could be the SIBLING div
				// (we can't check content easily after transformation,
				// but we can check that there's a div)
				log.Printf("      Found div element (possibly SIBLING)")
				hasSiblingDiv = true
			}
		}
	}

	if !hasConditionalTemplate {
		t.Errorf("BUG: No conditional template found in wrapper div")
	}

	if !hasSiblingDiv {
		t.Errorf("BUG: No SIBLING div found in wrapper div")
	}

	// Additional check: count non-whitespace nodes
	nonWhitespaceCount := 0
	for _, child := range wrapperDiv.Children {
		if _, ok := child.(*ast.TextNode); ok {
			// Skip text nodes for this count
			continue
		}
		nonWhitespaceCount++
	}

	// We should have at least 2 non-text nodes:
	// 1. The conditional template
	// 2. The SIBLING div
	if nonWhitespaceCount < 2 {
		t.Errorf("BUG: Expected at least 2 non-text nodes in wrapper (conditional + SIBLING), got %d", nonWhitespaceCount)
		log.Printf("\n[DEBUG] Wrapper div children details:")
		for i, child := range wrapperDiv.Children {
			log.Printf("  [%d] %T", i, child)
			if elem, ok := child.(*ast.Element); ok {
				log.Printf("      <%s>", elem.TagName)
			} else if txt, ok := child.(*ast.TextNode); ok {
				log.Printf("      TextNode: %q", strings.TrimSpace(txt.Content))
			}
		}
	}

	log.Printf("\n[TEST] ✅ Transformation test complete")
}
