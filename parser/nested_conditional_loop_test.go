package parser

import (
	"log"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestNestedConditionalWithLoop tests the exact structure from home.html
// where a loop with a nested conditional is inside an outer conditional
func TestNestedConditionalWithLoop(t *testing.T) {
	log.Println("\n=== TestNestedConditionalWithLoop ===")

	// This mimics the structure in home.html lines 42-79
	template := `{if age > 0}
  <div>Component Examples</div>

  <div class="animals">
    <h2>Animals Loop</h2>
    {for animal of animals}
      {if animal == "cat"}
        <div>Hi</div>
      {else}
        <div>Bye</div>
      {/if}
      <div class="type">SIBLING</div>
      <br>
    {/for}
  </div>

  <div>More examples</div>
{/if}`

	log.Printf("\n[TEST] Parsing nested structure...")

	result := BlockConditionalParser()(template)
	if !result.Successful {
		t.Fatalf("Failed to parse outer conditional: %s", result.Error)
	}

	outerConditional, ok := result.Value.(*ast.Conditional)
	if !ok {
		t.Fatalf("Expected *ast.Conditional, got %T", result.Value)
	}

	log.Printf("\n[TEST] Outer conditional parsed successfully")
	log.Printf("[TEST] Outer conditional IfContent has %d nodes", len(outerConditional.IfContent))

	// Find the loop inside the outer conditional
	var loop *ast.Loop
	_ = -1 // loopIndex
	for i, node := range outerConditional.IfContent {
		log.Printf("  Outer IfContent[%d]: %T", i, node)

		// The loop might be inside an Element (the <div class="animals">)
		if elem, ok := node.(*ast.Element); ok {
			log.Printf("    Element: <%s>", elem.TagName)
			if elem.TagName == "div" {
				// Look inside this div for the loop
				for j, child := range elem.Children {
					log.Printf("      Child[%d]: %T", j, child)
					if l, ok := child.(*ast.Loop); ok {
						loop = l
						_ = j
						log.Printf("        Found loop at child[%d]!", j)
						break
					}
				}
			}
		}

		// Or the loop might be a direct child
		if l, ok := node.(*ast.Loop); ok {
			loop = l
			_ = i
			log.Printf("    Found loop directly at index %d!", i)
			break
		}
	}

	if loop == nil {
		t.Fatalf("No loop found in outer conditional's IfContent")
	}

	log.Printf("\n[TEST] Loop found. Analyzing loop content...")
	log.Printf("[TEST] Loop.Content has %d nodes:", len(loop.Content))

	for i, node := range loop.Content {
		log.Printf("  Loop.Content[%d]: %T", i, node)
		if elem, ok := node.(*ast.Element); ok {
			log.Printf("    Element: <%s>", elem.TagName)
		}
	}

	// Count node types in loop content
	var innerConditional *ast.Conditional
	elementCount := 0
	textNodeCount := 0

	for i, node := range loop.Content {
		switch n := node.(type) {
		case *ast.Conditional:
			innerConditional = n
			log.Printf("  [%d] Found inner Conditional", i)
		case *ast.Element:
			elementCount++
			log.Printf("  [%d] Found Element: <%s>", i, n.TagName)
		case *ast.TextNode:
			trimmed := strings.TrimSpace(n.Content)
			if len(trimmed) > 0 {
				log.Printf("  [%d] Found non-whitespace TextNode: %q", i, trimmed)
			} else {
				textNodeCount++
			}
		}
	}

	// Verify structure
	if innerConditional == nil {
		t.Errorf("BUG: No inner conditional found in loop content")
	}

	// We should have at least 2 elements: the SIBLING div and the br
	if elementCount < 2 {
		t.Errorf("BUG: Expected at least 2 elements in loop content (SIBLING div + br), got %d", elementCount)
	}

	// Check the inner conditional structure
	if innerConditional != nil {
		log.Printf("\n[TEST] Analyzing inner conditional...")
		log.Printf("[TEST] Inner IfContent has %d nodes:", len(innerConditional.IfContent))
		for i, node := range innerConditional.IfContent {
			log.Printf("  [%d] %T", i, node)
			if elem, ok := node.(*ast.Element); ok {
				log.Printf("      Element: <%s>", elem.TagName)
			}
		}

		log.Printf("\n[TEST] Inner ElseContent has %d nodes:", len(innerConditional.ElseContent))
		for i, node := range innerConditional.ElseContent {
			log.Printf("  [%d] %T", i, node)
			if elem, ok := node.(*ast.Element); ok {
				log.Printf("      Element: <%s>", elem.TagName)
			}
		}

		// Check that SIBLING is NOT in the inner conditional's ElseContent
		for _, node := range innerConditional.ElseContent {
			if elem, ok := node.(*ast.Element); ok {
				if elem.TagName == "div" {
					for _, child := range elem.Children {
						if txt, ok := child.(*ast.TextNode); ok {
							if strings.Contains(txt.Content, "SIBLING") {
								t.Errorf("BUG: SIBLING found in inner conditional's ElseContent!")
							}
						}
					}
				}
			}
		}

		// Check that SIBLING is NOT in the inner conditional's IfContent
		for _, node := range innerConditional.IfContent {
			if elem, ok := node.(*ast.Element); ok {
				if elem.TagName == "div" {
					for _, child := range elem.Children {
						if txt, ok := child.(*ast.TextNode); ok {
							if strings.Contains(txt.Content, "SIBLING") {
								t.Errorf("BUG: SIBLING found in inner conditional's IfContent!")
							}
						}
					}
				}
			}
		}
	}

	log.Printf("\n[TEST] ✅ Test complete")
}
