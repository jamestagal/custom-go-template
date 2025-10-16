package builder

import (
	"fmt"
	"strings"
	"github.com/jimafisk/custom_go_template/ast"
)

func TestCSSBugManual() {
	component := ComponentTemplate{
		Name: "TestCSSBug",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "style",
					Children: []ast.Node{
						&ast.TextNode{Content: "body "},
						&ast.ExpressionNode{Expression: "color"},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})
	fmt.Println("Result:", result)
	
	if strings.Contains(result, "${props.") {
		fmt.Println("BUG: ExpressionNode in style tag converted to props")
	} else {
		fmt.Println("OK: No conversion")
	}
}
