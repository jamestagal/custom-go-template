package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// testAlpineParser specifically tests parsing Alpine.js attributes and directives
func testAlpineParser() {
	// Configure logging
	logFile, err := os.Create("alpine_parser_test.log")
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ltime | log.Lshortfile)

	// Test cases targeting Alpine.js specifically
	testCases := []struct {
		name     string
		input    string
		expected bool // Whether parsing should succeed
	}{
		{
			name:     "Basic Alpine x-data",
			input:    `<div x-data="{ message: 'Hello' }">{message }</div>`,
			expected: true,
		},
		{
			name:     "Alpine with complex data",
			input:    `<div x-data="{ message: 'Hello Alpine!', count: 0, items: ['apple', 'banana'] }">{message }</div>`,
			expected: true,
		},
		{
			name:     "Alpine shortcuts @ and :",
			input:    `<button @click="count++" :disabled="count > 10">Click me</button>`,
			expected: true,
		},
		{
			name: "Alpine full document with nested attributes",
			input: `<!DOCTYPE html>
<html lang="en" x-data="{ message: 'Hello Alpine!', count: 0 }">
<head>
    <meta charset="UTF-8">
    <title>Alpine Test</title>
</head>
<body>
    <h1 x-text="message"></h1>
    <button @click="count++" :disabled="count > 10">Increment</button>
    <p>Count: <span x-text="count"></span></p>
</body>
</html>`,
			expected: true,
		},
		{
			name:     "Alpine using test_basic.html",
			input:    readFileOrDefault("views/test_basic.html", ""),
			expected: true,
		},
		{
			name:     "Home template with components",
			input:    readFileOrDefault("views/home.html", ""),
			expected: true,
		},
		{
			name:     "Age template with dynamic content",
			input:    readFileOrDefault("views/age.html", ""),
			expected: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		fmt.Printf("Testing: %s\n", tc.name)

		if tc.input == "" {
			fmt.Printf("❌ Could not read test file for: %s\n", tc.name)
			continue
		}

		// Attempt to parse the input
		templateAST, err := parser.ParseTemplate(tc.input)

		// Check result against expected
		if tc.expected {
			if err != nil {
				fmt.Printf("❌ Expected success but got error: %v\n", err)
			} else {
				fmt.Printf("✅ Successfully parsed: %s\n", tc.name)
				// Print some info about the parsed AST
				printASTSummary(templateAST)
				// Validate Alpine attributes if present
				validateAlpineAttributes(templateAST)
			}
		} else {
			if err != nil {
				fmt.Printf("✅ Expected failure and got error: %v\n", err)
			} else {
				fmt.Printf("❌ Expected failure but parsing succeeded\n")
			}
		}
		fmt.Println(strings.Repeat("-", 50))
	}
}

// readFileOrDefault reads a file or returns a default value if the file doesn't exist
func readFileOrDefault(filePath, defaultValue string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Warning: Could not read file %s: %v\n", filePath, err)
		return defaultValue
	}
	return string(content)
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

// validateAlpineAttributes specifically looks for and validates Alpine.js attributes in the AST
func validateAlpineAttributes(template *ast.Template) {
	// Track Alpine attribute count for logging
	alpineAttributeCount := 0

	// Function to examine elements recursively
	var examineNode func(node ast.Node)
	examineNode = func(node ast.Node) {
		if node == nil {
			return
		}

		switch n := node.(type) {
		case *ast.Element:
			// Check attributes for Alpine directives
			for _, attr := range n.Attributes {
				if attr.IsAlpine {
					alpineAttributeCount++
					fmt.Printf("  Found Alpine attribute: %s=\"%s\" (Type: %s, Key: %s)\n",
						attr.Name, attr.Value, attr.AlpineType, attr.AlpineKey)
				}
			}

			// Recursively check children
			for _, child := range n.Children {
				examineNode(child)
			}
		case *ast.Conditional:
			// Check children in conditional branches
			for _, child := range n.IfContent {
				examineNode(child)
			}
			for _, elseIfContent := range n.ElseIfContent {
				for _, child := range elseIfContent {
					examineNode(child)
				}
			}
			for _, child := range n.ElseContent {
				examineNode(child)
			}
		case *ast.Loop:
			// Check children in loop content
			for _, child := range n.Content {
				examineNode(child)
			}
		}
	}

	// Start examining from root nodes
	for _, rootNode := range template.RootNodes {
		examineNode(rootNode)
	}

	fmt.Printf("Total Alpine.js attributes found: %d\n", alpineAttributeCount)
}

func main() {
	fmt.Println("Testing Alpine.js Parser")
	testAlpineParser()
}
