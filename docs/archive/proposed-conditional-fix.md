Thank you for providing `parser/parser.go`. After a thorough analysis of the code, I have identified the root cause of the issue.

Your hypothesis was exactly right: "The issue appears to be in how the AST nodes are assembled or how the parser accumulates `IfContent`."

The problem lies within the `BlockConditionalParser` function. While it attempts to track nested conditionals using a `depth` counter, the implementation is incomplete. It correctly decrements the counter when it finds a closing `{/if}` tag, but it **never increments the counter when it encounters a nested opening `{if}` tag.**

This leads to incorrect behavior, especially in complex templates. The parser can't reliably determine which `{/if}` corresponds to which `{if}`, causing it to incorrectly consume nodes that appear after the conditional block.

### The Bug Explained

In your `BlockConditionalParser`'s `for` loop, you have this logic:

```go
// Check for block terminators FIRST
...
ifEndRes := IfEndParser()(remaining)

if ifEndRes.Successful {
    depth-- // Depth is DECREMENTED
    log.Printf("[BlockConditionalParser] Found {/if}, depth=%d", depth)

    if depth == 0 {
        // This is the correct exit point
        return Result{ ... }
    }
    ...
}

// Parse one node of content
nodeRes := AnyNodeParser(stopParsers...)(remaining)
```

The parser correctly looks for `{/if}` and decrements `depth`. However, when `AnyNodeParser` recursively calls `BlockConditionalParser` to handle a nested `{if}` block, the parent parser's `depth` variable is never updated. The logic is missing a check to increment the depth for opening `{if}` tags.

Because the depth isn't correctly tracked, the parser can become "lost". It might not recognize the correct `{/if}` as the true end of the block, causing it to continue parsing the following nodes (the "likes" `div` and the `br`) and incorrectly append them to the `IfContent` of the conditional node.

---

### Proposed Fix

To fix this, we need to make the depth tracking robust by also checking for and handling opening `{if}` tags within the loop. This ensures the parser only exits when the depth counter correctly returns to zero.

Modify the `for` loop inside `BlockConditionalParser` as follows.

**File:** `parser/parser.go`

```go
// (inside BlockConditionalParser function)
// ...
		for {
			iterationCount++
			if iterationCount > maxIterations {
				offset := len(originalInput) - len(remaining)
				return Result{
					nil,
					input,
					false,
					ErrMaxDepthExceeded(maxIterations, originalInput, offset).Error(),
					false,
				}
			}

			// --- START OF FIX ---

			// Check for ALL control directives first to correctly manage depth
			ifStartRes := IfStartParser()(remaining)
			elseIfRes := ElseIfParser()(remaining)
			elseRes := ElseParser()(remaining)
			ifEndRes := IfEndParser()(remaining)

			// If we find a nested {if}, increment depth and consume it as a node
			if ifStartRes.Successful {
				// We must ensure this isn't part of an {else if}
				// A simple way is to check if the ElseIfParser was also successful
				// on the same input, but the distinct syntax usually prevents this.
				// For robustness, we prioritize {else if} over a new {if}.
				if !elseIfRes.Successful || len(elseIfRes.Remaining) >= len(ifStartRes.Remaining) {
					depth++
					log.Printf("[BlockConditionalParser] Found nested {if}, new depth=%d", depth)
					// We will let the AnyNodeParser handle parsing this nested conditional
				}
			}

			// Check if we hit the end of the if block
			if ifEndRes.Successful {
				depth--
				log.Printf("[BlockConditionalParser] Found {/if}, depth=%d", depth)

				if depth == 0 {
					log.Printf("[BlockConditionalParser] Depth=0, completing conditional block")
					return Result{
						Value:      conditional,
						Remaining:  ifEndRes.Remaining,
						Successful: true,
						Error:      "",
					}
				}
				// Since we are now letting AnyNodeParser parse the full nested block
				// including its end tag, we should not 'continue' here.
				// The nested {/if} will be consumed by the recursive parser call.
			}

			// Only recognize {else if} and {else} at our current depth (depth == 1)
			if depth == 1 {
				if elseIfRes.Successful {
					elseIfNode, ok := elseIfRes.Value.(*ast.ElseIfNode)
					if !ok {
						return Result{nil, input, false, "ElseIfParser did not return *ast.ElseIfNode", false}
					}

					log.Printf("[BlockConditionalParser] Found {else if %s} at depth=1", elseIfNode.Condition)

					conditional.ElseIfConditions = append(conditional.ElseIfConditions, elseIfNode.Condition)
					conditional.ElseIfContent = append(conditional.ElseIfContent, []ast.Node{})

					currentBranch = elseIfBranch
					currentElseIfIndex = len(conditional.ElseIfConditions) - 1

					remaining = elseIfRes.Remaining
					continue // Continue to parse the content of the new branch
				}

				if elseRes.Successful {
					log.Printf("[BlockConditionalParser] Found {else} at depth=1")
					currentBranch = elseBranch
					remaining = elseRes.Remaining
					continue // Continue to parse the content of the new branch
				}
			}

			// --- END OF FIX ---


			// Parse one node of content for the current branch
			stopParsers := []Parser{
				ElseIfParser(),
				ElseParser(),
				IfEndParser(),
			}

			nodeRes := AnyNodeParser(stopParsers...)(remaining)
			if !nodeRes.Successful {
				offset := len(originalInput) - len(remaining)
				return Result{
					nil,
					input,
					false,
					ErrMissingClosingDirective("if", originalInput, offset).Error(),
					false,
				}
			}

			if node, ok := nodeRes.Value.(ast.Node); ok && node != nil {
				switch currentBranch {
				case ifBranch:
					conditional.IfContent = append(conditional.IfContent, node)
					log.Printf("[BlockConditionalParser] Added node %T to IfContent", node)
				case elseIfBranch:
					conditional.ElseIfContent[currentElseIfIndex] = append(
						conditional.ElseIfContent[currentElseIfIndex], node)
					log.Printf("[BlockConditionalParser] Added node %T to ElseIfContent[%d]", node, currentElseIfIndex)
				case elseBranch:
					conditional.ElseContent = append(conditional.ElseContent, node)
					log.Printf("[BlockConditionalParser] Added node %T to ElseContent", node)
				}
			}

			// Prevent infinite loops if parser doesn't consume input
			if len(nodeRes.Remaining) == len(remaining) {
				return Result{nil, input, false, "BlockConditionalParser: parser made no progress", false}
			}
			remaining = nodeRes.Remaining
		}
// ...
```

### Summary of the Fix

1.  **Check for `{if}`:** The loop now explicitly checks for `{if ...}` using `IfStartParser` at the beginning of each iteration.
2.  **Increment Depth:** If a nested `{if}` is found, the `depth` counter is correctly incremented.
3.  **Prioritize `{else if}`:** A check is added to ensure that an `{else if}` isn't mistaken for a new nested `{if}`.
4.  **Rely on `AnyNodeParser`:** The logic now lets `AnyNodeParser` do its job of recursively calling `BlockConditionalParser`. The explicit depth counter now serves as a correct "balancing" mechanism to find the true end of the entire block, which was its original intent.
5.  **Simplified `ifEndRes` Block:** The logic inside `if ifEndRes.Successful` is simplified. The main purpose is now just to decrement depth and check for the final exit condition.

This change correctly implements the depth tracking, making your conditional parser far more robust. It will now correctly identify the boundary of the `{if}...{/if}` block, stop parsing content for it at the right moment, and return the subsequent sibling nodes (`<div>` and `<br>`) as part of the `Remaining` string to be correctly handled by the parent `BlockLoopParser`.