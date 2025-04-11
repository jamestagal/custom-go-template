package parser

import (
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// ComponentParser parses component tags (<Component /> or <={expr} />)
func ComponentParser() Parser {
	return func(input string) Result {
		isDynamic := strings.HasPrefix(input, "<=")
		isStatic := strings.HasPrefix(input, "<") && len(input) > 1 && (input[1] >= 'A' && input[1] <= 'Z')
		if !isDynamic && !isStatic {
			return Result{nil, input, false, "not a component tag", false} // Added Dynamic
		}
		endTag := "/>"
		endTagPos := strings.Index(input, endTag)
		if endTagPos == -1 {
			return Result{nil, input, false, "no closing /> for component tag", false} // Added Dynamic
		}
		contentStart := 1
		if isDynamic {
			contentStart = 2
		}
		fullContent := input[contentStart:endTagPos]
		nameOrPath := "PlaceholderComp"
		props := []ast.ComponentProp{}
		parts := strings.Fields(fullContent)
		if len(parts) > 0 {
			nameOrPath = parts[0]
		} // TODO: Parse props
		compNode := &ast.ComponentNode{Name: nameOrPath, Dynamic: isDynamic, Props: props}
		return Result{compNode, input[endTagPos+len(endTag):], true, "", false} // Added Dynamic
	}
}

// TODO: Enhance component parsing with prop handling

// ComponentPropParser parses component properties
// func ComponentPropParser() Parser {
//     // Implementation for parsing component props
// }
