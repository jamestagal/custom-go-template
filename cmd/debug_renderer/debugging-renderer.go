package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
)

// debugRenderer tests the entire pipeline from parsing to rendering
func debugRenderer() {
	// Configure logging
	logFile, err := os.Create("renderer_debug.log")
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ltime | log.Lshortfile)

	// Create a directory for output
	outputDir := "./debug_output"
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	// Define test props
	props := map[string]any{
		"name":    "John",
		"age":     2,
		"animals": []string{"cat", "dog", "pig"},
	}

	// Files to test
	testFiles := []string{
		"views/test_basic.html",
		"views/home.html",
		"views/age.html",
		"views/double.html",
		"views/head.html",
		"views/todos.html",
		"views/mycomp.html",
	}

	for _, file := range testFiles {
		fmt.Printf("Testing: %s\n", file)

		// 1. Verify the file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", file)
			continue
		}

		// 2. Read the file content
		fileContent, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			continue
		}

		// 3. Parse the template
		fmt.Printf("Parsing template %s...\n", file)
		templateAST, err := parser.ParseTemplate(string(fileContent))
		if err != nil {
			fmt.Printf("❌ Failed to parse template: %v\n", err)
			continue
		}
		fmt.Printf("✅ Successfully parsed template %s\n", file)

		// 4. Summary of the parsed AST
		printASTSummary(templateAST)

		// 5. Render the template using the renderer package
		fmt.Printf("Rendering template %s...\n", file)
		markup, script, style := renderer.Render(file, props)

		// 6. Check render output
		if markup == "" {
			fmt.Printf("⚠️ Empty markup generated for %s\n", file)
		} else {
			fmt.Printf("✅ Generated %d bytes of markup\n", len(markup))
		}

		if script == "" {
			fmt.Printf("⚠️ No script generated for %s\n", file)
		} else {
			fmt.Printf("✅ Generated %d bytes of script\n", len(script))
		}

		if style == "" {
			fmt.Printf("⚠️ No style generated for %s\n", file)
		} else {
			fmt.Printf("✅ Generated %d bytes of style\n", len(style))
		}

		// 7. Write output files for inspection
		fileBase := strings.TrimSuffix(strings.TrimPrefix(file, "views/"), ".html")
		outputFileBase := fmt.Sprintf("%s/%s", outputDir, fileBase)

		// Write markup
		err = os.WriteFile(outputFileBase+".html", []byte(markup), 0644)
		if err != nil {
			fmt.Printf("❌ Failed to write markup file: %v\n", err)
		}

		// Write script
		if script != "" {
			err = os.WriteFile(outputFileBase+".js", []byte(script), 0644)
			if err != nil {
				fmt.Printf("❌ Failed to write script file: %v\n", err)
			}
		}

		// Write style
		if style != "" {
			err = os.WriteFile(outputFileBase+".css", []byte(style), 0644)
			if err != nil {
				fmt.Printf("❌ Failed to write style file: %v\n", err)
			}
		}

		fmt.Println(strings.Repeat("-", 50))
	}

	fmt.Printf("Debug output written to %s directory\n", outputDir)
}

// printASTSummary prints a summary of the AST structure
func printASTSummary(template *ast.Template) {
	if template == nil {
		fmt.Println("Template is nil")
		return
	}

	fmt.Printf("Template has %d root nodes\n", len(template.RootNodes))

	// Count node types
	nodeTypes := make(map[string]int)
	countNodeTypes(template.RootNodes, nodeTypes)

	fmt.Println("Node types:")
	for nodeType, count := range nodeTypes {
		fmt.Printf("  - %s: %d\n", nodeType, count)
	}
}

// countNodeTypes counts the different types of nodes in the AST
func countNodeTypes(nodes []ast.Node, counts map[string]int) {
	for _, node := range nodes {
		if node == nil {
			continue
		}

		nodeType := node.NodeType()
		counts[nodeType]++

		// Recursively count child nodes
		switch n := node.(type) {
		case *ast.Element:
			countNodeTypes(n.Children, counts)
		case *ast.Conditional:
			if n.IfContent != nil {
				countNodeTypes(n.IfContent, counts)
			}
			for _, content := range n.ElseIfContent {
				if content != nil {
					countNodeTypes(content, counts)
				}
			}
			if n.ElseContent != nil {
				countNodeTypes(n.ElseContent, counts)
			}
		case *ast.Loop:
			if n.Content != nil {
				countNodeTypes(n.Content, counts)
			}
		}
	}
}

func main() {
	fmt.Println("Debugging Renderer Pipeline")
	debugRenderer()
}
