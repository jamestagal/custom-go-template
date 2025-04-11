package parser

import (
	"github.com/jimafisk/custom_go_template/ast"
)

// IfStartParser parses {if condition} directives
func IfStartParser() Parser {
	return Map(Between(String("{if "), String("}"), TakeUntil(String("}"))),
		func(v interface{}) (interface{}, error) { return v, nil })
}

// ElseIfParser parses {else if condition} directives
func ElseIfParser() Parser {
	return Map(Between(String("{else if "), String("}"), TakeUntil(String("}"))),
		func(v interface{}) (interface{}, error) { return v, nil })
}

// ElseParser parses {else} directives
func ElseParser() Parser {
	return Map(String("{else}"),
		func(v interface{}) (interface{}, error) { return &ast.ElseNode{}, nil })
}

// IfEndParser parses {/if} directives
func IfEndParser() Parser {
	return Map(String("{/if}"),
		func(v interface{}) (interface{}, error) { return &ast.IfEndNode{}, nil })
}

// ForStartParser parses {for ...} directives
func ForStartParser() Parser {
	return Map(Between(String("{for "), String("}"), TakeUntil(String("}"))),
		func(v interface{}) (interface{}, error) { return v, nil })
}

// ForEndParser parses {/for} directives
func ForEndParser() Parser {
	return Map(String("{/for}"),
		func(v interface{}) (interface{}, error) { return &ast.ForEndNode{}, nil })
}

// TODO: Add full directive parsers that handle nested content

// ConditionalParser parses complete if/else if/else structures (future implementation)
// func ConditionalParser() Parser {
//     // Implementation for parsing complete conditional blocks
// }

// LoopParser parses complete for loops with their content (future implementation)
// func LoopParser() Parser {
//     // Implementation for parsing complete loop blocks
// }
