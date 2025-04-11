package scoping

import (
	"log"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/jimafisk/custom_go_template/utils"
	// renderer "github.com/jimafisk/custom_go_template/renderer" // Avoid this dependency if possible
)

// ScopedElement holds information about an element and its generated scope class.
type ScopedElement struct {
	Tag         string
	ID          string
	Classes     []string
	ScopedClass string
}

// ScopeHTML adds scoped classes to a full HTML document string and returns the modified markup
// and a list of elements that were scoped. It also handles adding x-data and x-text attributes.
func ScopeHTML(markup string, props map[string]any) (string, []ScopedElement) {
	scopedElements := []ScopedElement{}
	node, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		// Handle parsing error more gracefully?
		log.Printf("Warning: Failed to parse HTML for scoping: %v", err)
		return markup, scopedElements // Return original markup on error
	}

	node, scopedElements = traverse(node, scopedElements, props)

	// Render the modified HTML back to a string
	buf := &strings.Builder{}
	err = html.Render(buf, node)
	if err != nil {
		log.Fatal(err) // Consider returning error instead of fatal
	}
	// The default html.Render escapes entities like quotes in attributes.
	// We might need selective unescaping or different rendering if it causes issues.
	markup = buf.String() // html.UnescapeString might be too broad

	return markup, scopedElements
}

// ScopeHTMLComp adds scoped classes to an HTML fragment (component) string.
// It uses html.ParseFragment to avoid adding <html><body> tags.
func ScopeHTMLComp(comp_markup string, comp_props map[string]any, comp_data map[string]any) (string, []ScopedElement) {
	scopedElements := []ScopedElement{}
	fragments := []string{}
	// Use a dummy context node (like body) for ParseFragment
	contextNode := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}
	nodes, err := html.ParseFragment(strings.NewReader(comp_markup), contextNode)
	if err != nil {
		log.Printf("Warning: Failed to parse HTML fragment for scoping: %v", err)
		return comp_markup, scopedElements // Return original on error
	}

	for _, node := range nodes {
		node, scopedElements = traverse(node, scopedElements, comp_props) // Pass comp_props here

		// Add x-data for component props if needed
		if len(comp_data) > 0 { // Use comp_data which contains expressions/literals for x-data
			attr := html.Attribute{
				Key: "x-data",
				Val: utils.MakeGetter(comp_data), // Use utils.MakeGetter
			}
			// Append x-data carefully, avoid duplicates if possible
			hasXData := false
			for _, existingAttr := range node.Attr {
				if existingAttr.Key == "x-data" {
					hasXData = true
					break
				}
			}
			if !hasXData {
				node.Attr = append(node.Attr, attr)
			}
		}

		buf := &strings.Builder{}
		err := html.Render(buf, node)
		if err != nil {
			log.Fatal(err) // Consider returning error
		}
		fragments = append(fragments, buf.String()) // Don't unescape fragment render
	}
	comp_markup = strings.Join(fragments, "")

	return comp_markup, scopedElements
}

// traverse walks the HTML node tree, applying scoping and transformations.
func traverse(node *html.Node, scopedElements []ScopedElement, props map[string]any) (*html.Node, []ScopedElement) {
	// Use a closure for recursion to easily pass scopedElements
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "html" {
			// Add top-level x-data if props exist (only for full document scopeHTML)
			if len(props) > 0 {
				// Check if x-data already exists
				hasXData := false
				for _, attr := range n.Attr {
					if attr.Key == "x-data" {
						hasXData = true
						break
					}
				}
				if !hasXData {
					attr := html.Attribute{
						Key: "x-data",
						// Use AnyToJSValue for correct JS literal formatting
						Val: strings.ReplaceAll(utils.AnyToJSValue(props), "\"", "'"), // Replace double quotes for HTML attribute
					}
					n.Attr = append(n.Attr, attr)
				}
			}
		}

		if n.Type == html.TextNode {
			// Expressions in text nodes are handled during the part processing in Render.
			// Removing temporary call to evalAllBrackets from here.
			// n.Data = renderer.evalAllBrackets(n.Data, props)
		}

		if n.Type == html.ElementNode && n.DataAtom != 0 { // Check for valid DataAtom
			tag := n.Data
			id := ""
			var classes []string
			scopedClass := "" // Initialize scopedClass

			// Find existing scope class or generate a new one
			hasScopeClass := false
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					classes = strings.Fields(attr.Val) // Use Fields to handle multiple spaces
					for _, c := range classes {
						if strings.HasPrefix(c, "plenti-") {
							scopedClass = c
							hasScopeClass = true
							break
						}
					}
				}
				if attr.Key == "id" {
					id = attr.Val
				}
			}

			if !hasScopeClass {
				randomStr, err := utils.GenerateRandom() // Use utils.GenerateRandom
				if err != nil {
					log.Fatal(err) // Consider returning error
				}
				scopedClass = "plenti-" + randomStr
			}

			// Process attributes: add scope class, handle Alpine bindings, evaluate expressions
			newAttrs := []html.Attribute{}
			hasClassAttr := false
			for _, attr := range n.Attr {
				currentAttr := attr // Work on a copy
				if currentAttr.Key == "class" {
					hasClassAttr = true
					if !hasScopeClass {
						currentAttr.Val = strings.TrimSpace(currentAttr.Val + " " + scopedClass)
					}
				}

				// Expressions in attributes are handled during the part processing in Render.
				// Removing temporary call to evalAllBrackets from here.
				// if strings.Contains(currentAttr.Val, "{") && strings.Contains(currentAttr.Val, "}") {
				// ...
				// 		currentAttr.Val = renderer.evalAllBrackets(currentAttr.Val, props)
				// ...
				// }
				newAttrs = append(newAttrs, currentAttr)
			}

			// Add class attribute if it didn't exist
			if !hasClassAttr && !hasScopeClass {
				newAttrs = append(newAttrs, html.Attribute{Key: "class", Val: scopedClass})
			}
			n.Attr = newAttrs // Update node attributes

			// Add x-text if necessary (check children)
			hasAlpineText := false
			for _, attr := range n.Attr {
				if attr.Key == "x-text" {
					hasAlpineText = true
					break
				}
			}
			if !hasAlpineText {
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					if child.Type == html.TextNode && strings.Contains(child.Data, "{") && strings.Contains(child.Data, "}") {
						// Check if it's just whitespace around an expression handled elsewhere
						trimmedData := strings.TrimSpace(child.Data)
						if strings.HasPrefix(trimmedData, "{") && strings.HasSuffix(trimmedData, "}") {
							attr := html.Attribute{
								Key: "x-text",
								// Create JS template literal
								Val: "`" + strings.ReplaceAll(strings.ReplaceAll(child.Data, "{", "${"), "\"", "'") + "`",
							}
							n.Attr = append(n.Attr, attr)
							break
						}
					}
				}
			}

			// Record the scoped element details
			scopedElements = append(scopedElements, ScopedElement{
				Tag:         tag,
				ID:          id,
				Classes:     classes, // Original classes might be useful
				ScopedClass: scopedClass,
			})
		}

		// Recurse through children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(node)
	return node, scopedElements
}

// GetScopedClass finds the generated scope class for a given target (tag, id, class).
// This function might be better placed in the scoping package as well.
func GetScopedClass(target string, target_type string, scopedElements []ScopedElement) string {
	for _, elem := range scopedElements {
		if target_type == "tag" && elem.Tag == target {
			return elem.ScopedClass
		}
		if target_type == "id" && elem.ID == target {
			return elem.ScopedClass
		}
		if target_type == "class" {
			for _, class := range elem.Classes {
				if class == target {
					return elem.ScopedClass
				}
			}
		}
	}
	return ""
}
