package parser

import (
	"fmt"
	"log"

	"github.com/jimafisk/custom_go_template/ast"
)

// Sequence runs a series of parsers in order and returns their results as a slice
func Sequence(parsers ...Parser) Parser {
	return func(input string) Result {
		values := make([]interface{}, 0, len(parsers))
		current := input
		for _, p := range parsers {
			result := p(current)
			if !result.Successful {
				return Result{nil, input, false, result.Error, false} // Added Dynamic field
			}
			if result.Value != nil {
				values = append(values, result.Value)
			}
			current = result.Remaining
		}
		return Result{values, current, true, "", false} // Added Dynamic field
	}
}

// Choice tries each parser in order until one succeeds
func Choice(parsers ...Parser) Parser {
	return func(input string) Result {
		for _, p := range parsers {
			result := p(input)
			if result.Successful {
				return result
			}
		}
		return Result{nil, input, false, "no choice matched", false} // Added Dynamic field
	}
}

// Between parses content between start and end parsers
func Between(start, end Parser, content Parser) Parser {
	return func(input string) Result {
		startResult := start(input)
		if !startResult.Successful {
			return startResult
		}
		contentResult := content(startResult.Remaining)
		if !contentResult.Successful {
			return Result{nil, input, false, contentResult.Error, false} // Added Dynamic field
		}
		endResult := end(contentResult.Remaining)
		if !endResult.Successful {
			return Result{nil, input, false, endResult.Error, false} // Added Dynamic field
		}
		return Result{contentResult.Value, endResult.Remaining, true, "", false} // Added Dynamic field
	}
}

// Many applies parser p zero or more times, collecting AST nodes
func Many(p Parser) Parser {
	return func(input string) Result {
		var values []ast.Node
		current := input
		for {
			if len(current) == 0 {
				break
			}
			result := p(current)
			if !result.Successful {
				break
			}
			// Handle both single nodes and slices of nodes returned by parsers
			if node, ok := result.Value.(ast.Node); ok && node != nil { // Check for nil node
				values = append(values, node)
			} else if nodes, ok := result.Value.([]ast.Node); ok {
				for _, n := range nodes { // Append non-nil nodes from slice
					if n != nil {
						values = append(values, n)
					}
				}
			} else if result.Value != nil {
				// Ignore non-node, non-nil results if necessary, or log warning
				// log.Printf("[Many] Warning: Parser %T returned non-node, non-nil value: %T", p, result.Value)
			}

			// Safety check: if parser succeeded but didn't consume input, break to avoid infinite loop
			if result.Remaining == current {
				// Only break if the parser didn't return a value either, otherwise allow zero-consumption parsers
				if result.Value == nil {
					log.Printf("[Many] Warning: Parser %T succeeded without consuming input or returning value. Breaking loop.", p)
					break
				}
			}
			current = result.Remaining
		}
		return Result{values, current, true, "", false} // Added Dynamic field
	}
}

// Map transforms the result of a parser using a function
func Map(p Parser, fn func(interface{}) (interface{}, error)) Parser {
	return func(input string) Result {
		res := p(input)
		if !res.Successful {
			return res
		}
		newValue, err := fn(res.Value)
		if err != nil {
			return Result{nil, input, false, fmt.Sprintf("map function failed: %v", err), false} // Added Dynamic field
		}
		return Result{newValue, res.Remaining, true, "", false} // Added Dynamic field
	}
}
