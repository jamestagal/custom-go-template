# Debug Tools

This directory contains standalone debugging utilities for testing the template parser. Each tool is a separate executable with its own `main()` function.

## Tools

### detailed_parse
**Location**: `debug/detailed_parse/main.go`

Tests parsing of complex templates with nested loops and conditionals.

**Run**: `go run debug/detailed_parse/main.go`

### nested_conditional
**Location**: `debug/nested_conditional/main.go`

Tests parsing of nested conditional blocks.

**Run**: `go run debug/nested_conditional/main.go`

### parser_bug
**Location**: `debug/parser_bug/main.go`

Debug tool for investigating parser issues with loops and conditionals.

**Run**: `go run debug/parser_bug/main.go`

## Organization

These files were moved from the project root to avoid "main redeclared" build errors. Each tool is now in its own subdirectory with its own `main.go` file, allowing them to be built and run independently without conflicts.
