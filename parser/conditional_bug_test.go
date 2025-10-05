package parser

import (
	"log"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// getElementContent recursively extracts text content from an element
func getElementContent(node ast.Node) string {
	switch n := node.(type) {
	case *ast.TextNode:
		return n.Content
	case *ast.Element:
		var content string
		for _, child := range n.Children {
			content += getElementContent(child)
		}
		return content
	default:
		return ""
	}
}

// TestConditionalInLoopBug reproduces the bug where BlockConditionalParser
// incorrectly includes content after {/if} in the Conditional's ElseContent.
func TestConditionalInLoopBug(t *testing.T) {
	log.Println("\n=== TestConditionalInLoopBug ===")

	template := `{for animal of animals}
  {if animal == "cat"}
    <div>Hi</div>
  {else}
    <div>Bye</div>
  {/if}
  <div>SIBLING</div>
{/for}`

	log.Printf("\n[TEST] Parsing template:\n%s\n", template)

	// Parse the loop
	result := BlockLoopParser()(template)
	if !result.Successful {
		t.Fatalf("Failed to parse loop: %s", result.Error)
	}

	loop, ok := result.Value.(*ast.Loop)
	if !ok {
		t.Fatalf("Expected *ast.Loop, got %T", result.Value)
	}

	log.Printf("\n[TEST] Loop parsed successfully. Content has %d nodes:", len(loop.Content))
	for i, node := range loop.Content {
		log.Printf("  [%d] %T", i, node)
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			log.Printf("      Element <%s> content: %q", elem.TagName, strings.TrimSpace(content))
		}
	}

	// Find the conditional node
	var conditional *ast.Conditional
	conditionalIndex := -1
	for i, node := range loop.Content {
		if c, ok := node.(*ast.Conditional); ok {
			conditional = c
			conditionalIndex = i
			break
		}
	}

	if conditional == nil {
		t.Fatalf("No Conditional node found in loop content")
	}

	log.Printf("\n[TEST] Conditional found at index %d", conditionalIndex)
	log.Printf("[TEST] Conditional.IfContent has %d nodes:", len(conditional.IfContent))
	for i, node := range conditional.IfContent {
		log.Printf("  [%d] %T", i, node)
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			log.Printf("      Element <%s> content: %q", elem.TagName, strings.TrimSpace(content))
		}
	}

	log.Printf("\n[TEST] Conditional.ElseContent has %d nodes:", len(conditional.ElseContent))
	for i, node := range conditional.ElseContent {
		log.Printf("  [%d] %T", i, node)
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			log.Printf("      Element <%s> content: %q", elem.TagName, strings.TrimSpace(content))
		}
	}

	// BUG CHECK: The SIBLING div should NOT be in ElseContent
	// It should be a sibling in Loop.Content after the Conditional
	foundSiblingInElseContent := false
	for _, node := range conditional.ElseContent {
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			if strings.Contains(content, "SIBLING") {
				foundSiblingInElseContent = true
				log.Printf("\n[TEST] ❌ BUG DETECTED: SIBLING div found in ElseContent!")
			}
		}
	}

	if foundSiblingInElseContent {
		t.Errorf("BUG: SIBLING div should be a sibling to Conditional in Loop.Content, not in ElseContent")
	}

	// CHECK: Loop should have the SIBLING div as a sibling to the Conditional
	siblingDivFound := false
	for i, node := range loop.Content {
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			if strings.Contains(content, "SIBLING") {
				siblingDivFound = true
				log.Printf("\n[TEST] ✅ SIBLING div found at loop.Content[%d]", i)
			}
		}
	}

	if !siblingDivFound {
		t.Errorf("BUG: SIBLING div not found in Loop.Content")
		// Print all nodes for debugging
		log.Printf("\n[DEBUG] All loop content nodes:")
		for i, node := range loop.Content {
			log.Printf("  [%d] %T", i, node)
			if elem, ok := node.(*ast.Element); ok {
				content := getElementContent(elem)
				log.Printf("      Element <%s> content: %q", elem.TagName, content)
			}
		}
	} else {
		log.Printf("\n[TEST] ✅ CORRECT: SIBLING div is in Loop.Content (expected)")
	}
}

// TestMinimalConditionalBug tests just the conditional parsing in isolation
func TestMinimalConditionalBug(t *testing.T) {
	log.Println("\n=== TestMinimalConditionalBug ===")

	template := `{if animal == "cat"}
    <div>Hi</div>
  {else}
    <div>Bye</div>
  {/if}
  <div>SIBLING</div>`

	log.Printf("\n[TEST] Parsing template:\n%s\n", template)

	// Parse the conditional
	result := BlockConditionalParser()(template)
	if !result.Successful {
		t.Fatalf("Failed to parse conditional: %s", result.Error)
	}

	conditional, ok := result.Value.(*ast.Conditional)
	if !ok {
		t.Fatalf("Expected *ast.Conditional, got %T", result.Value)
	}

	log.Printf("\n[TEST] Conditional parsed successfully")
	log.Printf("[TEST] Remaining content length: %d", len(result.Remaining))
	if len(result.Remaining) > 0 {
		log.Printf("[TEST] Remaining starts with: %.50s", result.Remaining)
	}

	log.Printf("\n[TEST] ElseContent has %d nodes:", len(conditional.ElseContent))
	for i, node := range conditional.ElseContent {
		log.Printf("  [%d] %T", i, node)
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			log.Printf("      Element <%s> content: %q", elem.TagName, strings.TrimSpace(content))
		}
	}

	// Check if SIBLING is in ElseContent (it shouldn't be)
	foundSiblingInElseContent := false
	for _, node := range conditional.ElseContent {
		if elem, ok := node.(*ast.Element); ok {
			content := getElementContent(elem)
			if strings.Contains(content, "SIBLING") {
				foundSiblingInElseContent = true
				log.Printf("\n[TEST] ❌ BUG DETECTED: SIBLING div found in ElseContent!")
			}
		}
	}

	if foundSiblingInElseContent {
		t.Errorf("BUG: SIBLING div should be in Remaining, not in ElseContent")
	} else {
		log.Printf("\n[TEST] ✅ CORRECT: SIBLING div not in ElseContent")
	}

	// Check that remaining content contains SIBLING
	if !strings.Contains(result.Remaining, "SIBLING") {
		t.Errorf("BUG: Remaining content should contain SIBLING div, but it doesn't")
		log.Printf("[TEST] Remaining: %q", result.Remaining)
	} else {
		log.Printf("\n[TEST] ✅ CORRECT: SIBLING div is in remaining content")
	}
}
