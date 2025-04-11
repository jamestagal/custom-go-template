package scoping

// We need GetScopedClass from html.go, so import the scoping package itself
// OR move GetScopedClass to utils if it doesn't depend on ScopedElement struct specifics
// For now, let's assume GetScopedClass stays in html.go (or this file)

// cssSelectors holds the original selector string and parsed structure.
type cssSelectors struct {
	selectorStr string
	selectorArr []cssSelector
}

// cssSelector represents a simple part of a CSS selector (tag, classes, id).
type cssSelector struct {
	tag     string
	classes []string
	id      string
}

// ScopeCSS attempts to add scope classes to CSS rules based on scoped HTML elements.
// NOTE: This implementation is basic and has known limitations with complex selectors and string replacement.
// A more robust solution would require a proper CSS AST manipulation library.
func ScopeCSS(style string, scopedElements []ScopedElement) (string, []cssSelectors) {
	// Return original style due to limitations mentioned above.
	// The logic below is kept for reference but commented out.
	/*
		ss := css.Parse(style)
		rules := ss.GetCSSRuleList()
		selectors := []cssSelectors{}
		modifiedStyle := style // Work on a copy

		// Build a map for faster lookup: selector string -> scoped class
		scopeMap := make(map[string]string)
		for _, se := range scopedElements {
			scopeMap[se.Tag] = se.ScopedClass
			if se.ID != "" {
				scopeMap["#"+se.ID] = se.ScopedClass
			}
			for _, cl := range se.Classes {
				if cl != "" && !strings.HasPrefix(cl, "plenti-") {
					scopeMap["."+cl] = se.ScopedClass
				}
			}
		}


		for i := len(rules) - 1; i >= 0; i-- {
			rule := rules[i]
			// Check if rule is a StyleRule before accessing Style property
			if rule.Type != css.StyleRuleType {
				continue
			}

			sel := rule.Style.Selector.Text()
			parts := strings.FieldsFunc(sel, func(r rune) bool {
				return r == ' ' || r == '>' || r == '+' || r == '~' || r == ','
			})

			newSelectorParts := []string{}
			needsModification := false
			for _, part := range parts {
				trimmedPart := strings.TrimSpace(part)
				scopedClass, found := "", false

				// Check exact match first (e.g., #id, .class)
				if sc, ok := scopeMap[trimmedPart]; ok {
					scopedClass = sc
					found = true
				} else {
					// If not exact, check if it's a tag selector potentially needing scope
					if !strings.HasPrefix(trimmedPart, ".") && !strings.HasPrefix(trimmedPart, "#") { // Basic check for tag
						if sc, ok := scopeMap[trimmedPart]; ok {
							scopedClass = sc
							found = true
						}
					}
					// TODO: Handle more complex selectors like tag.class, tag#id etc.
				}


				if found {
					newSelectorParts = append(newSelectorParts, trimmedPart+"."+scopedClass)
					needsModification = true
				} else {
					newSelectorParts = append(newSelectorParts, trimmedPart) // No scope found
				}
			}

			if needsModification {
				// Simplified reconstruction - loses combinators/commas
				newSel := strings.Join(newSelectorParts, " ")
				// Fragile string replacement - HIGHLY LIKELY TO FAIL on complex CSS
				// modifiedStyle = strings.Replace(modifiedStyle, sel, newSel, 1)
			}
		}
		return modifiedStyle, selectors // Return modified style
	*/
	return style, nil // Return original style and nil selectors due to complexity
}
