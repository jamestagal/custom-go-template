package main

import (
	"fmt"
	"log"
	
	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

func main() {
	template := `{for animal of animals}
  {if animal == "cat"}
    <div>Hi {animal}!</div>
  {else}
    <div>Bye {animal}.</div>
  {/if}
  <div class="type-{animal}">{name} likes: {animal}s</div>
  <br>
{/for}`

	log.SetFlags(0)
	tree, err := parser.ParseTemplate(template)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}
	
	loop := tree.RootNodes[0].(*ast.Loop)
	
	fmt.Printf("=== LOOP CONTENT ===\n")
	fmt.Printf("Total nodes in loop: %d\n\n", len(loop.Content))
	
	for i, node := range loop.Content {
		switch n := node.(type) {
		case *ast.Conditional:
			fmt.Printf("[%d] Conditional\n", i)
			fmt.Printf("    IfContent: %d nodes\n", len(n.IfContent))
			fmt.Printf("    ElseContent: %d nodes\n", len(n.ElseContent))
			
			// Check if div or br are mistakenly in IfContent or ElseContent
			for j, ifNode := range n.IfContent {
				if elem, ok := ifNode.(*ast.Element); ok {
					if elem.TagName == "div" {
						// Check if this is the "Hi" div or the "type-" div
						fmt.Printf("      IfContent[%d]: <div> - SHOULD BE HERE (Hi/Bye message)\n", j)
					}
				}
			}
			for j, elseNode := range n.ElseContent {
				if elem, ok := elseNode.(*ast.Element); ok {
					if elem.TagName == "div" {
						fmt.Printf("      ElseContent[%d]: <div> - SHOULD BE HERE (Bye message)\n", j)
					}
				}
			}
			
		case *ast.Element:
			fmt.Printf("[%d] Element: <%s>\n", i, n.TagName)
			if n.TagName == "div" || n.TagName == "br" {
				fmt.Printf("    ✓ THIS SHOULD BE A SIBLING TO THE CONDITIONAL\n")
			}
			
		case *ast.TextNode:
			content := n.Content
			if len(content) > 20 {
				content = content[:20] + "..."
			}
			fmt.Printf("[%d] TextNode: %q\n", i, content)
		}
	}
	
	// Final verification
	fmt.Printf("\n=== VERIFICATION ===\n")
	
	// Count non-whitespace nodes
	nonWhitespace := 0
	for _, node := range loop.Content {
		if text, ok := node.(*ast.TextNode); ok {
			if len(text.Content) == 0 || text.Content == "\n" || text.Content == "\n  " {
				continue
			}
		}
		nonWhitespace++
	}
	
	fmt.Printf("Non-whitespace nodes: %d\n", nonWhitespace)
	fmt.Printf("Expected: 3 (Conditional, div, br)\n")
	
	if nonWhitespace == 3 {
		fmt.Printf("\n✓ PARSER IS CORRECT - No bug detected!\n")
	} else {
		fmt.Printf("\n✗ PARSER HAS BUG - Wrong number of nodes\n")
	}
}
