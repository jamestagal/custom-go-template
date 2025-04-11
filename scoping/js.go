package scoping

import (
	"log"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	// Need ScopedElement definition
)

// visitor applies scoping modifications to JS AST nodes.
type visitor struct {
	scopedElements []ScopedElement // Use the exported ScopedElement type
}

func (*visitor) Exit(js.INode) {}

func (v *visitor) Enter(node js.INode) js.IVisitor {
	switch n := node.(type) {
	// Scope querySelector/querySelectorAll calls
	case *js.CallExpr:
		if memberExpr, ok := n.X.(*js.DotExpr); ok {
			objName := string(memberExpr.X.String())
			propName := string(memberExpr.Y.Data)
			if objName == "document" && (propName == "querySelector" || propName == "querySelectorAll") {
				if len(n.Args.List) > 0 {
					arg := n.Args.List[0] // Assume selector is the first argument
					if lit, ok := arg.Value.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken {
						selector, err := strconv.Unquote(string(lit.Data))
						if err == nil {
							// Basic handling: if selector starts simply (tag, .class, #id)
							parts := strings.Fields(selector) // Simplistic split
							if len(parts) > 0 {
								firstPart := parts[0]
								target := firstPart
								targetType := "tag"
								if strings.HasPrefix(firstPart, ".") {
									target = strings.TrimPrefix(firstPart, ".")
									targetType = "class"
								} else if strings.HasPrefix(firstPart, "#") {
									target = strings.TrimPrefix(firstPart, "#")
									targetType = "id"
								}

								scopedClass := GetScopedClass(target, targetType, v.scopedElements) // Use exported GetScopedClass
								if scopedClass != "" && !strings.Contains(selector, "."+scopedClass) {
									// Append scope class - This is very basic and might break complex selectors
									newSelector := firstPart + "." + scopedClass
									if len(parts) > 1 {
										newSelector += " " + strings.Join(parts[1:], " ")
									}
									// Update the AST node with the new quoted selector
									n.Args.List[0].Value = &js.LiteralExpr{
										Data: []byte(strconv.Quote(newSelector)),
									}
								}
							}
						}
					}
				}
			}
		}
	// Potentially scope other things like getElementById, getElementsByClassName if needed
	// case *js.Var: // Example: Renaming variables (if needed, though likely complex)
	// 	if n.Decl.String() == "LexicalDecl" && !strings.Contains(n.String(), "_plenti_") {
	// 		randomStr, _ := utils.GenerateRandom() // Assuming utils is imported
	// 		n.Data = append(n.Data, []byte("_plenti_"+randomStr)...)
	// 	}
	default:
	}
	return v
}

// ScopeJS applies scoping rules to JavaScript code using AST manipulation.
func ScopeJS(script string, scopedElements []ScopedElement) string {
	if strings.TrimSpace(script) == "" {
		return ""
	}
	ast, err := js.Parse(parse.NewInputString(script), js.Options{})
	if err != nil {
		log.Printf("Warning: Failed to parse script for JS scoping: %v", err)
		return script // Return original script on parse error
	}
	v := visitor{scopedElements: scopedElements}
	js.Walk(&v, ast)
	return ast.JSString()
}

// GetScopedClass is defined in html.go within the same package.
