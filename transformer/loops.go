package transformer

import (
	"fmt"
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)


// transformLoop transforms a Loop node into an Alpine.js compatible structure
// BUILD-TIME EXPANSION: This function now expands loops at build time instead of
// generating x-for templates, allowing component.name resolution in loops.
//
// **HYBRID APPROACH**: Build-time when possible, runtime fallback when needed
//
// Pattern: Build-Time Loop Expansion with Runtime Fallback [Cognitive Load: 22]
// Cognitive Load: 22 (collection resolution: 6, iteration: 8, fallback: 4, edge cases: 4)
//
// Behavior:
//   1. Attempt to resolve collection from dataScope
//   2. Check if loop body needs runtime evaluation (Alpine directives, reactive expressions)
//   3. IF resolvable at build-time AND body is static:
//      → Expand loop in Go (no x-for template)
//   4. ELSE (runtime-only collection OR reactive content):
//      → Fallback to Alpine.js x-for template (runtime)
//
// Example BUILD-TIME:
//   Input:  {for component in components}
//             <div>{component.name}</div>
//           {/for}
//
//   With dataScope: {"components": [{"name": "Hero"}, {"name": "Footer"}]}
//
//   Output: <div><span x-text="component.name"></span></div>
//           <div><span x-text="component.name"></span></div>
//           (2 expanded divs, NO x-for template)
//
// Example RUNTIME (fallback):
//   Input:  {for item in $store.cart.items}
//             <div>{item.name}</div>
//           {/for}
//
//   Output: <template x-for="item in $store.cart.items">
//             <div><span x-text="item.name"></span></div>
//           </template>
//           (Alpine x-for template for runtime evaluation)
func transformLoop(node *ast.Loop, dataScope map[string]any) []ast.Node {
	log.Printf("========== transformLoop START ==========")
	log.Printf("transformLoop: iterator=%s, value=%s, collection=%s, isOf=%v",
		node.Iterator, node.Value, node.Collection, node.IsOf)

	// Extract variables from the collection expression (add collection to parent scope)
	// This makes the collection accessible in the loop expression
	extractVariablesFromExpr(node.Collection, dataScope)

	log.Printf("transformLoop: parent scope keys: %v", getMapKeys(dataScope))

	// IMPORTANT: Parser uses different field assignments for {#each} vs {for} syntax
	// {#each items as item} → Iterator="", Value="item", IsOf=true
	// {for item in items} → Iterator="item", Value="", IsOf=false
	// We need to normalize this

	var itemVar, indexVar string

	if node.IsOf {
		// This is {#each} syntax: Value=item, Iterator=index (or empty)
		itemVar = node.Value
		indexVar = node.Iterator
	} else {
		// This is {for} syntax: Value=item, Iterator=index (or empty)
		// FIXED: The parser sets Value=item and Iterator=index for BOTH syntaxes
		itemVar = node.Value
		indexVar = node.Iterator
	}

	log.Printf("transformLoop: normalized - itemVar=%s, indexVar=%s", itemVar, indexVar)

	// Clean up the collection expression
	cleanedCollection := cleanLoopCollection(node.Collection)
	log.Printf("transformLoop: cleaned collection: %s", cleanedCollection)

	// Transform store expressions in collection (Task 2.3)
	// If collection is a store expression like "$cart.items", transform to "$store.cart.items"
	cleanedCollection = transformStoreExpressionInCollection(cleanedCollection)

	log.Printf("transformLoop: after store transformation: %s", cleanedCollection)

	// CRITICAL FIX: Check if loop body needs runtime evaluation
	// If the body contains Alpine directives or reactive expressions that reference the loop variable,
	// we MUST use runtime x-for template, not build-time expansion
	log.Printf("transformLoop: calling loopBodyNeedsRuntime() to check if runtime evaluation needed")
	log.Printf("transformLoop: loop has %d content nodes to check", len(node.Content))
	needsRuntime := loopBodyNeedsRuntime(node.Content, itemVar, indexVar)

	if needsRuntime {
		log.Printf("transformLoop: ✓ RUNTIME DETECTION: loop body contains runtime-reactive content, forcing runtime x-for template")
		return generateRuntimeLoopTemplate(node, itemVar, indexVar, cleanedCollection, dataScope)
	}

	// HYBRID APPROACH: Try build-time expansion first
	// Attempt to resolve collection from dataScope
	collectionData, collectionResolved := resolveCollectionFromScope(cleanedCollection, dataScope)

	if !collectionResolved {
		log.Printf("transformLoop: collection '%s' not resolvable in dataScope, falling back to runtime x-for", cleanedCollection)
		return generateRuntimeLoopTemplate(node, itemVar, indexVar, cleanedCollection, dataScope)
	}

	log.Printf("transformLoop: ✓ BUILD-TIME EXPANSION: collection '%s' resolved with %d items, expanding loop at build time",
		cleanedCollection, len(collectionData))

	// BUILD-TIME EXPANSION: Iterate over collection and transform each iteration
	var expandedNodes []ast.Node

	for iterationIndex, itemData := range collectionData {
		log.Printf("transformLoop: [iteration %d/%d] expanding with item data: %v", iterationIndex+1, len(collectionData), itemData)

		// Create child scope for this iteration (clones parent scope)
		iterationScope := CreateChildScope(dataScope)

		// Add loop variables to iteration scope
		iterationScope[itemVar] = itemData
		if indexVar != "" {
			iterationScope[indexVar] = iterationIndex
		}

		log.Printf("transformLoop: [iteration %d/%d] iteration scope keys: %v", iterationIndex+1, len(collectionData), getMapKeys(iterationScope))

		// Transform loop body with iteration scope
		// applyAlpineWrapper=false, inLiteralContext=false for loop body
		transformedBody := transformNodes(node.Content, iterationScope, false, false)

		// Append to result
		expandedNodes = append(expandedNodes, transformedBody...)
	}

	log.Printf("transformLoop: build-time expansion complete, produced %d nodes", len(expandedNodes))
	log.Printf("========== transformLoop END ==========")
	return expandedNodes
}

// loopBodyNeedsRuntime determines if the loop body contains reactive content
// that requires runtime Alpine.js evaluation via x-for template.
// Pattern: Content Analysis for Runtime Detection [Cognitive Load: 12]
//
// CRITICAL FIX: This function now ONLY checks for Alpine directives and event handlers,
// NOT simple expressions. Simple expressions like {component.fields.title} can be
// build-time expanded because the loop variable values are known at build time.
//
// Returns true if the loop body contains:
// - Alpine directives (@click, :style, x-text, x-model, etc.)
// - HTML event handlers (onclick, onchange, onsubmit, etc.)
// - DYNAMIC attributes that reference loop variables (for reactive binding)
//
// Returns false for:
// - Simple expressions in content: {component.name}, {component.fields.title}
// - Static attributes: class="item" (even if "item" matches loop var name)
// - These can be build-time expanded when collection is resolvable
//
// This prevents build-time expansion ONLY when the loop variable must be
// available in Alpine's reactive scope at runtime (for directives/handlers).
func loopBodyNeedsRuntime(content []ast.Node, itemVar, indexVar string) bool {
	log.Printf("  >> loopBodyNeedsRuntime START: checking %d nodes, itemVar=%s, indexVar=%s", len(content), itemVar, indexVar)

	for i, node := range content {
		log.Printf("  >> loopBodyNeedsRuntime: node[%d] type=%T", i, node)

		switch n := node.(type) {
		case *ast.Element:
			log.Printf("  >> loopBodyNeedsRuntime: checking Element <%s> with %d attributes", n.TagName, len(n.Attributes))

			// Check if element has Alpine directives
			for j, attr := range n.Attributes {
				attrName := attr.Name
				log.Printf("  >> loopBodyNeedsRuntime:   attr[%d]: name=%s, value=%s, dynamic=%v", j, attrName, attr.Value, attr.Dynamic)

				// Alpine directives that need runtime evaluation
				if strings.HasPrefix(attrName, "@") ||
				   strings.HasPrefix(attrName, ":") ||
				   strings.HasPrefix(attrName, "x-") {
					log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND Alpine directive %s, needs runtime", attrName)
					return true
				}

				// Check for HTML event handlers (onclick, onchange, onsubmit, etc.)
				// These need runtime evaluation because they reference loop variables
				if strings.HasPrefix(attrName, "on") && len(attrName) > 2 {
					log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND HTML event handler '%s', needs runtime", attrName)
					return true
				}

				// CRITICAL FIX: Only check DYNAMIC attributes for loop variable references
				// Static attributes like class="item" should NOT trigger runtime evaluation
				// even if "item" matches the loop variable name
				if attr.Dynamic {
					if itemVar != "" && strings.Contains(attr.Value, itemVar) {
						log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND dynamic attribute value references loop var %s, needs runtime", itemVar)
						return true
					}
					if indexVar != "" && strings.Contains(attr.Value, indexVar) {
						log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND dynamic attribute value references index var %s, needs runtime", indexVar)
						return true
					}
				} else {
					log.Printf("  >> loopBodyNeedsRuntime: skipping static attribute (not dynamic)")
				}
			}

			// Recursively check children
			log.Printf("  >> loopBodyNeedsRuntime: recursively checking %d children of <%s>", len(n.Children), n.TagName)
			if loopBodyNeedsRuntime(n.Children, itemVar, indexVar) {
				log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND runtime content in children of <%s>", n.TagName)
				return true
			}

		case *ast.ExpressionNode:
			// CRITICAL FIX: DO NOT check expressions for loop variable references
			// Simple expressions like {component.fields.title} CAN be build-time expanded
			// because we have the actual values in the iteration scope.
			//
			// Only Alpine directives and attributes need runtime evaluation.
			log.Printf("  >> loopBodyNeedsRuntime: checking ExpressionNode: %s (allowing build-time expansion)", n.Expression)
			// DO NOT return true here - expressions are fine for build-time expansion

		case *ast.Conditional:
			log.Printf("  >> loopBodyNeedsRuntime: checking Conditional with condition: %s", n.IfCondition)

			// CRITICAL FIX: Conditionals with loop variable references CAN be build-time expanded
			// IF the collection is resolvable. The condition will be evaluated during iteration.
			// Only flag as runtime if used in Alpine directive context (handled by Element case above)
			log.Printf("  >> loopBodyNeedsRuntime: allowing conditional (will be evaluated during build-time expansion)")

			// Check content recursively (use IfContent field)
			log.Printf("  >> loopBodyNeedsRuntime: checking %d nodes in IfContent", len(n.IfContent))
			if loopBodyNeedsRuntime(n.IfContent, itemVar, indexVar) {
				log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND runtime content in IfContent")
				return true
			}

			// Check else-if branches (use ElseIfConditions and ElseIfContent)
			for i, elseIfCond := range n.ElseIfConditions {
				log.Printf("  >> loopBodyNeedsRuntime: checking else-if[%d] condition: %s", i, elseIfCond)
				if i < len(n.ElseIfContent) {
					log.Printf("  >> loopBodyNeedsRuntime: checking %d nodes in ElseIfContent[%d]", len(n.ElseIfContent[i]), i)
					if loopBodyNeedsRuntime(n.ElseIfContent[i], itemVar, indexVar) {
						log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND runtime content in ElseIfContent[%d]", i)
						return true
					}
				}
			}

			// Check else branch (use ElseContent field, check len instead of nil)
			if len(n.ElseContent) > 0 {
				log.Printf("  >> loopBodyNeedsRuntime: checking %d nodes in ElseContent", len(n.ElseContent))
				if loopBodyNeedsRuntime(n.ElseContent, itemVar, indexVar) {
					log.Printf("  >> loopBodyNeedsRuntime: ✓ FOUND runtime content in ElseContent")
					return true
				}
			}

		default:
			log.Printf("  >> loopBodyNeedsRuntime: skipping node type %T", node)
		}
	}

	log.Printf("  >> loopBodyNeedsRuntime END: ✗ NO runtime content detected, returning false")
	return false
}

// generateRuntimeLoopTemplate creates an Alpine.js x-for template for runtime-only collections
// Pattern: Runtime Loop Template Generation [Cognitive Load: 12]
//
// This function handles cases where collection cannot be resolved at build-time:
// - Store collections: $store.cart.items
// - Complex expressions: Array(count), filteredItems, etc.
// - Dynamic collections that depend on runtime state
// - Loop bodies with Alpine directives or reactive expressions
//
// The generated template will be evaluated by Alpine.js at runtime.
func generateRuntimeLoopTemplate(node *ast.Loop, itemVar, indexVar, cleanedCollection string, dataScope map[string]any) []ast.Node {
	log.Printf("generateRuntimeLoopTemplate: creating x-for template for runtime evaluation")

	// Create child scope for loop body (loop variables are runtime-only, so nil markers)
	loopBodyScope := CreateChildScope(dataScope)

	// Add variables to loop body scope as nil (runtime markers)
	if itemVar != "" {
		loopBodyScope[itemVar] = nil
	}
	if indexVar != "" {
		loopBodyScope[indexVar] = nil
	}

	// Build the x-for expression - always use Alpine.js "in" syntax for arrays
	var loopExpr string

	if indexVar != "" {
		// CRITICAL FIX: Alpine.js x-for syntax is (item, index), not (index, item)
		// The ITEM comes FIRST, then the INDEX
		// Both index and item: "(item, index) in collection"
		loopExpr = fmt.Sprintf("(%s, %s) in %s", itemVar, indexVar, cleanedCollection)
	} else {
		// Only item: "item in items"
		loopExpr = fmt.Sprintf("%s in %s", itemVar, cleanedCollection)
	}

	log.Printf("generateRuntimeLoopTemplate: generated x-for expression: %s", loopExpr)

	// Create the template element with the loop expression
	// Use loop body scope for transforming content
	return createLoopTemplate(loopExpr, node.Content, loopBodyScope)
}

// needsWrapper determines if the loop content needs a wrapper div
// Alpine.js x-for requires exactly ONE child element
func needsWrapper(content []ast.Node) bool {
	// Filter out whitespace-only text nodes
	var nonWhitespaceNodes []ast.Node
	for _, node := range content {
		if textNode, ok := node.(*ast.TextNode); ok {
			// Skip whitespace-only text nodes
			trimmed := strings.TrimSpace(textNode.Content)
			if trimmed == "" {
				continue
			}
		}
		nonWhitespaceNodes = append(nonWhitespaceNodes, node)
	}

	// If there's more than one non-whitespace node, we need a wrapper
	if len(nonWhitespaceNodes) > 1 {
		return true
	}

	// If there's exactly one non-whitespace node, check if it's a template element
	if len(nonWhitespaceNodes) == 1 {
		if elem, ok := nonWhitespaceNodes[0].(*ast.Element); ok {
			// Template elements need a wrapper
			if elem.TagName == "template" {
				return true
			}
		}
	}

	// Single element that's not a template doesn't need wrapper
	return false
}

// createLoopTemplate creates a template element with x-for attribute
// Pattern: Template Element Creation [Cognitive Load: 8]
func createLoopTemplate(loopExpr string, content []ast.Node, dataScope map[string]any) []ast.Node {
	log.Printf("createLoopTemplate: received %d content nodes: %v", len(content), content)

	// Transform the content nodes with the loop body scope
	// applyAlpineWrapper=false, inLiteralContext=false for loop template content
	transformedContent := transformNodes(content, dataScope, false, false)

	log.Printf("createLoopTemplate: needsWrapper returned %v for %d nodes", needsWrapper(transformedContent), len(transformedContent))

	// Determine wrapper need
	if needsWrapper(transformedContent) {
		log.Printf("createLoopTemplate: wrapping %d nodes in div", len(transformedContent))
		// Multiple nodes or template element - wrap in div
		wrapper := &ast.Element{
			TagName:  "div",
			Children: transformedContent,
		}
		transformedContent = []ast.Node{wrapper}
	} else {
		log.Printf("createLoopTemplate: NOT wrapping, using nodes directly")
	}

	// Create the template element
	templateElement := &ast.Element{
		TagName: "template",
		Attributes: []ast.Attribute{
			{Name: "x-for", Value: loopExpr},
		},
		Children: transformedContent,
	}

	return []ast.Node{templateElement}
}

// cleanLoopCollection removes whitespace from collection expression
// Pattern: String Cleaning [Cognitive Load: 3]
func cleanLoopCollection(collection string) string {
	return strings.TrimSpace(collection)
}
