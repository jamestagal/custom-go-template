package renderer

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast" // Import AST package
	// Removed duplicate import of ast
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/scoping"
	"github.com/jimafisk/custom_go_template/utils" // Import utils
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

// --- Alpine.js Attribute Generation ---

func escapeAttrValue(value string) string {
	// Basic escaping for HTML attributes. More robust escaping might be needed.
	// Replace " with " to prevent breaking attribute quotes
	// Replace < and > to prevent HTML injection if value comes from untrusted source
	value = strings.ReplaceAll(value, `"`, `"`)
	value = strings.ReplaceAll(value, `<`, `<`)
	value = strings.ReplaceAll(value, `>`, `>`)
	// Consider escaping & to & as well
	// value = strings.ReplaceAll(value, `&`, `&`)
	return value
}

// GenerateAlpineDirectives processes Alpine directives from attributes.
// Note: This assumes evaluation of dynamic values ({...}) happens *before* this step.
// The current structure evaluates during GenerateMarkup, which needs adjustment.
// For now, it will render the raw expression string.
func GenerateAlpineDirectives(attributes []ast.Attribute) string {
	var builder strings.Builder

	// Process x-data first as it sets up the context
	for _, attr := range attributes {
		if attr.IsAlpine && attr.AlpineType == "data" {
			// TODO: Evaluate attr.Value if it's dynamic before escaping
			builder.WriteString(fmt.Sprintf(`x-data="%s" `, escapeAttrValue(attr.Value)))
		}
	}

	// Process other Alpine directives
	for _, attr := range attributes {
		if !attr.IsAlpine || attr.AlpineType == "data" {
			continue // Skip non-Alpine or already processed x-data
		}

		// TODO: Evaluate attr.Value if it's dynamic before escaping
		escapedValue := escapeAttrValue(attr.Value)

		switch attr.AlpineType {
		case "bind":
			if attr.AlpineKey != "" {
				// Shorthand :key or full x-bind:key
				builder.WriteString(fmt.Sprintf(`x-bind:%s="%s" `, attr.AlpineKey, escapedValue))
			} else {
				// x-bind="{...}" object syntax - value itself is the binding object
				builder.WriteString(fmt.Sprintf(`x-bind="%s" `, escapedValue))
			}
		case "on":
			if attr.AlpineKey != "" {
				// Shorthand @event or full x-on:event
				builder.WriteString(fmt.Sprintf(`x-on:%s="%s" `, attr.AlpineKey, escapedValue))
			} // else: Invalid x-on without event? Log warning?
		case "text":
			builder.WriteString(fmt.Sprintf(`x-text="%s" `, escapedValue))
		case "html":
			builder.WriteString(fmt.Sprintf(`x-html="%s" `, escapedValue))
		case "model":
			builder.WriteString(fmt.Sprintf(`x-model="%s" `, escapedValue))
		case "init":
			builder.WriteString(fmt.Sprintf(`x-init="%s" `, escapedValue))
		case "show":
			builder.WriteString(fmt.Sprintf(`x-show="%s" `, escapedValue))
		case "transition":
			// Transitions might have modifiers (x-transition:enter, etc.) handled by AlpineKey
			if attr.AlpineKey != "" {
				builder.WriteString(fmt.Sprintf(`x-transition:%s="%s" `, attr.AlpineKey, escapedValue))
			} else {
				builder.WriteString(fmt.Sprintf(`x-transition="%s" `, escapedValue)) // Basic x-transition
			}
		case "effect":
			builder.WriteString(fmt.Sprintf(`x-effect="%s" `, escapedValue))
		case "ignore":
			builder.WriteString(`x-ignore `) // Boolean directive
		case "ref":
			builder.WriteString(fmt.Sprintf(`x-ref="%s" `, escapedValue))
		case "cloak":
			builder.WriteString(`x-cloak `) // Boolean directive
		case "if": // x-if requires <template>
			builder.WriteString(fmt.Sprintf(`x-if="%s" `, escapedValue))
		case "for": // x-for requires <template>
			builder.WriteString(fmt.Sprintf(`x-for="%s" `, escapedValue))
		// Handle other Alpine directives similarly
		default:
			// Handle generic x-* directives or custom ones
			// This assumes the format x-plugin:key or x-plugin
			if attr.AlpineKey != "" {
				builder.WriteString(fmt.Sprintf(`x-%s:%s="%s" `,
					attr.AlpineType, attr.AlpineKey, escapedValue))
			} else {
				builder.WriteString(fmt.Sprintf(`x-%s="%s" `,
					attr.AlpineType, escapedValue))
			}
		}
	}

	return strings.TrimSpace(builder.String()) // Trim trailing space
}

// Render renders the template with the given data using an AST approach.
func Render(path string, props map[string]any) (string, string, string) {
	c, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", path, err)
	}
	template := string(c)

	// 1. Parse the template into an AST
	templateAST, err := parser.ParseTemplate(template)
	if err != nil {
		log.Fatalf("Failed to parse template %s into AST: %v", path, err)
	}
	log.Printf("Parsed AST: %#v\n", templateAST) // Log the generated AST

	// If the AST has empty root nodes, use the original template
	if len(templateAST.RootNodes) == 0 {
		log.Printf("Warning: AST has empty root nodes. Using original template as fallback.")
		// Parse the template using a simple HTML parser
		// This is a fallback for when the AST parsing fails
		return renderFallback(template, props)
	}

	// Initialize return values
	var markup, script, style string
	var bodyNodes []ast.Node

	// 2. Extract top-level sections and process Fence
	var fenceNode *ast.FenceSection
	var scriptNode *ast.ScriptSection
	var styleNode *ast.StyleSection
	var components []Component // Define components here, needed by RenderComponents

	for _, node := range templateAST.RootNodes {
		switch n := node.(type) {
		case *ast.FenceSection:
			if fenceNode != nil {
				log.Println("Warning: Multiple fence sections found, using the first one.")
			} else {
				fenceNode = n
			}
		case *ast.ScriptSection:
			if scriptNode != nil {
				log.Println("Warning: Multiple script sections found, using the first one.")
			} else {
				scriptNode = n
			}
		case *ast.StyleSection:
			if styleNode != nil {
				log.Println("Warning: Multiple style sections found, using the first one.")
			} else {
				styleNode = n
			}
		default:
			// Collect all other nodes as body nodes
			bodyNodes = append(bodyNodes, n)
		}
	}

	// Process Fence Logic (if fence exists)
	var propsDecl string // Store JS declarations for context
	if fenceNode != nil {
		fenceContent := fenceNode.RawContent // Use RawContent for now
		var fenceComponents []Component
		fenceContent, fenceComponents = GetComponents(fenceContent)
		fenceContent = SetProps(fenceContent, props)
		allVars := GetAllVars(fenceContent)
		props = EvaluateProps(fenceContent, allVars, props)
		components = fenceComponents // Assign components extracted from fence
		propsDecl = utils.DeclProps(props)
	} else {
		propsDecl = utils.DeclProps(props)
		components = []Component{} // Initialize if no fence
	}

	// 3. Transform AST Body (Placeholder)
	transformedBodyNodes := TransformAST(bodyNodes, props, propsDecl) // Pass propsDecl

	// 4. Generate Output from AST (Placeholders / Temporary Logic)
	markup = GenerateMarkup(transformedBodyNodes, props, propsDecl) // Pass propsDecl
	if scriptNode != nil {
		script = scriptNode.Content
	}
	if styleNode != nil {
		style = styleNode.Content
	}

	// --- Render Components & Scope (Still needs refactoring for AST) ---
	// Since RenderComponents is still needed temporarily, keep it uncommented for now.
	markup, script, style = RenderComponents(markup, script, style, props, components, Render)

	// Scoping should ideally happen on the AST or during generation.
	markup, scopedElements := scoping.ScopeHTML(markup, props)
	style, _ = scoping.ScopeCSS(style, scopedElements)
	script = scoping.ScopeJS(script, scopedElements)

	// --- Final JS Processing ---
	if strings.TrimSpace(script) != "" {
		jsAst, err := js.Parse(parse.NewInputString(script), js.Options{})
		if err != nil {
			log.Printf("Warning: Failed to parse final script for %s: %v", path, err)
		} else {
			script = jsAst.JSString()
		}
	}

	// Return the processed strings
	return markup, script, style
}

// renderFallback is a simple fallback renderer that uses the original template
// when the AST parsing fails or returns empty root nodes.
func renderFallback(template string, props map[string]any) (string, string, string) {
	log.Printf("Using fallback renderer for template")

	// Extract script and style tags
	var script, style string

	// Extract script content
	scriptRegex := regexp.MustCompile(`<script[^>]*>([\s\S]*?)</script>`)
	scriptMatches := scriptRegex.FindAllStringSubmatch(template, -1)
	for _, match := range scriptMatches {
		if len(match) > 1 {
			script += match[1] + "\n"
		}
	}

	// Extract style content
	styleRegex := regexp.MustCompile(`<style[^>]*>([\s\S]*?)</style>`)
	styleMatches := styleRegex.FindAllStringSubmatch(template, -1)
	for _, match := range styleMatches {
		if len(match) > 1 {
			style += match[1] + "\n"
		}
	}

	// Process expressions in the template
	propsDecl := utils.DeclProps(props)
	exprRegex := regexp.MustCompile(`{([^}]*)}`)
	markup := exprRegex.ReplaceAllStringFunc(template, func(match string) string {
		expr := match[1 : len(match)-1] // Remove { and }
		evaluated := fmt.Sprintf("%v", EvalJS(expr, propsDecl))
		return evaluated
	})

	// Return the processed strings
	return markup, script, style
}

// --- AST Processing Functions ---

// TransformAST walks the AST, evaluates expressions, resolves conditionals/loops etc.
// Returns the transformed AST (or potentially modifies in place).
// NOTE: This is a simplified starting point. Full implementation is complex.
func TransformAST(nodes []ast.Node, props map[string]any, propsDecl string) []ast.Node {
	transformedNodes := []ast.Node{}
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.ExpressionNode:
			// Evaluation now happens in GenerateMarkup
			transformedNodes = append(transformedNodes, n)
		case *ast.TextNode:
			transformedNodes = append(transformedNodes, n)
		case *ast.Element:
			// Recursively transform children
			n.Children = TransformAST(n.Children, props, propsDecl) // Pass propsDecl down
			// TODO: Evaluate expressions within attributes
			transformedNodes = append(transformedNodes, n)
		// TODO: Handle *ast.Conditional, *ast.Loop, *ast.ComponentNode
		default:
			transformedNodes = append(transformedNodes, n)
		}
	}
	return transformedNodes
}

// GenerateMarkup converts a slice of AST nodes into an HTML string.
func GenerateMarkup(nodes []ast.Node, props map[string]any, propsDecl string) string {
	var builder strings.Builder
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.TextNode:
			log.Printf("GenerateMarkup: Adding TextNode: '%s'", n.Content) // Added log
			builder.WriteString(n.Content)                                 // Directly write text content
		case *ast.ExpressionNode:
			// Evaluate expression during generation using propsDecl
			evaluated := fmt.Sprintf("%v", EvalJS(n.Expression, propsDecl))                                 // Pass propsDecl
			log.Printf("GenerateMarkup: Evaluating Expression '%s', Result: '%s'", n.Expression, evaluated) // Added log
			builder.WriteString(evaluated)
		case *ast.Element:
			log.Printf("GenerateMarkup: Adding Element: <%s>", n.TagName) // Added log
			log.Printf("GenerateMarkup: Adding Element: <%s>", n.TagName) // Added log
			builder.WriteString("<" + n.TagName)

			// --- Attribute Generation ---
			hasAlpine := false
			for _, attr := range n.Attributes {
				if attr.IsAlpine {
					hasAlpine = true
					break
				}
			}

			if hasAlpine {
				// Process Alpine directives first
				alpineAttrsStr := GenerateAlpineDirectives(n.Attributes)
				if alpineAttrsStr != "" {
					builder.WriteString(" " + alpineAttrsStr)
				}
				// Add non-Alpine attributes
				for _, attr := range n.Attributes {
					if !attr.IsAlpine {
						// TODO: Evaluate attr.Value if dynamic ({...}) before escaping
						// For now, renders raw value or placeholder
						builder.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name, escapeAttrValue(attr.Value)))
					}
				}
			} else {
				// Process attributes normally (no Alpine)
				for _, attr := range n.Attributes {
					// TODO: Evaluate attr.Value if dynamic ({...}) before escaping
					builder.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name, escapeAttrValue(attr.Value)))
				}
			}
			// --- End Attribute Generation ---

			if n.SelfClosing {
				builder.WriteString("/>")
			} else {
				builder.WriteString(">")
				builder.WriteString(GenerateMarkup(n.Children, props, propsDecl)) // Pass propsDecl down
				builder.WriteString("</" + n.TagName + ">")
			}
		case *ast.ComponentNode:
			builder.WriteString(fmt.Sprintf("<!-- Component %s placeholder -->", n.Name))
		case *ast.IfEndNode, *ast.ForEndNode, *ast.ElseNode:
			// Do nothing - these should be handled by TransformAST
		default:
			// log.Printf("Warning: Ignoring unhandled node type %T during markup generation", n)
		}
	}
	return builder.String()
}

// --- Placeholder Functions (To be implemented or removed) ---

// func GenerateScript(scriptNode *ast.ScriptSection) string { ... }
// func GenerateStyle(styleNode *ast.StyleSection) string { ... }
