package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
)

// testHTMLParser tests our parser with different HTML snippets
func testHTMLParser() {
	// Setup logging for debugging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.Lshortfile)

	// Test cases
	testCases := []struct {
		name     string
		input    string
		expected bool // Whether parsing should succeed
	}{
		{
			name:     "Basic Element",
			input:    "<div>Hello</div>",
			expected: true,
		},
		{
			name:     "Self-closing Element",
			input:    "<input type=\"text\" />",
			expected: true,
		},
		{
			name:     "Nested Elements",
			input:    "<div><p>Text</p></div>",
			expected: true,
		},
		{
			name:     "Alpine.js Attribute",
			input:    "<div x-data=\"{ message: 'Hello' }\">{{message}}</div>",
			expected: true,
		},
		{
			name:     "HTML with Expression",
			input:    "<p>Hello, {name}!</p>",
			expected: true,
		},
		{
			name:     "HTML with Comment",
			input:    "<div><!-- This is a comment --></div>",
			expected: true,
		},
		{
			name: "Complete Simple Document",
			input: `
<html>
<head>
  <title>Test</title>
</head>
<body>
  <h1>Hello, {name}!</h1>
  <p>Welcome to our site.</p>
</body>
</html>`,
			expected: true,
		},
		{
			name: "Document with Alpine.js",
			input: `
<html lang="en" x-data="{ message: 'Hello Alpine!', count: 0 }">
<head>
  <meta charset="UTF-8">
  <title>Alpine Test</title>
</head>
<body>
  <h1 x-text="message"></h1>
  <button @click="count++">Increment</button>
  <p>Count: <span x-text="count"></span></p>
</body>
</html>`,
			expected: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		fmt.Printf("Testing: %s\n", tc.name)

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

// printASTSummary prints a summary of the AST structure
func printASTSummary(template *ast.Template) {
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
		nodeType := node.NodeType()
		counts[nodeType]++

		// Recursively count child nodes
		switch n := node.(type) {
		case *ast.Element:
			countNodeTypes(n.Children, counts)
		case *ast.Conditional:
			countNodeTypes(n.IfContent, counts)
			for _, content := range n.ElseIfContent {
				countNodeTypes(content, counts)
			}
			countNodeTypes(n.ElseContent, counts)
		case *ast.Loop:
			countNodeTypes(n.Content, counts)
		}
	}
}

func main() {
	fmt.Println("Testing HTML Parser")
	testHTMLParser()
}
