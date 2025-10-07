# Technology Stack

**Project:** Custom Go Template Engine for Plenti
**Last Updated:** October 7, 2025

---

## Overview

The Custom Go Template Engine is built with a carefully selected technology stack optimized for:
- **Build-time performance** (native Go speed)
- **Runtime efficiency** (minimal JavaScript)
- **Developer experience** (familiar syntax, fast feedback)
- **Production readiness** (type safety, testing, observability)

---

## Core Technologies

### Programming Language

#### Go 1.22.2
**Role:** Primary implementation language for all engine components

**Why Go:**
- ✅ **Performance** - Native compilation, fast execution
- ✅ **Concurrency** - Excellent for parallel template processing
- ✅ **Type Safety** - Catch errors at compile time
- ✅ **Deployment** - Single binary, no runtime dependencies
- ✅ **Ecosystem** - Rich standard library, great tooling

**Features Used:**
- Generics (Go 1.18+) for type-safe collections
- Context package for cancellation and timeouts
- Testing package for comprehensive test coverage
- `fmt.Errorf` with `%w` for error wrapping (Go 1.13+)

**Project Structure:**
```
go.mod         # Go 1.22.2 minimum version
go.sum         # Dependency checksums
```

---

## Dependencies

### Production Dependencies

#### goja (JavaScript Runtime)
**Package:** `github.com/dop251/goja v0.0.0-20240516125602`

**Purpose:** Execute JavaScript expressions and functions from fence sections

**Usage:**
```go
import "github.com/dop251/goja"

vm := goja.New()
result, err := vm.RunString(fenceSection.Functions)
```

**Why This Library:**
- Pure Go implementation (no CGo)
- ES5.1+ compatibility
- Good performance for template expressions
- Active maintenance

**Considerations:**
- May be optional in future (if fence functions moved to Go)
- Current usage: Evaluating fence section JavaScript
- Performance: <1ms for typical expressions

---

#### tdewolff/parse (HTML/CSS Parsing)
**Package:** `github.com/tdewolff/parse/v2 v2.7.15`

**Purpose:** Low-level HTML and CSS parsing utilities

**Usage:**
```go
import "github.com/tdewolff/parse/v2/html"

tokenizer := html.NewTokenizer(reader)
for {
    tt, _ := tokenizer.Next()
    // Process tokens...
}
```

**Why This Library:**
- High performance (streaming parser)
- Low memory allocation
- HTML5 compliant
- Battle-tested in minifiers

**Current Usage:**
- HTML structure parsing
- CSS style extraction
- Future: CSS transformation for scoping

---

#### golang.org/x/net (HTML Processing)
**Package:** `golang.org/x/net v0.26.0`

**Purpose:** HTML parsing and manipulation

**Usage:**
```go
import "golang.org/x/net/html"

doc, err := html.Parse(reader)
// Traverse and modify DOM
```

**Why This Library:**
- Official Go supplementary library
- Robust HTML parsing
- DOM-like API
- Good error handling

**Current Usage:**
- HTML structure validation
- DOM traversal for component discovery
- Attribute extraction

---

### Development Dependencies

#### Go Standard Library Packages

**Testing:**
- `testing` - Test framework
- `testing/quick` - Property-based testing
- `net/http/httptest` - HTTP testing utilities

**Utilities:**
- `strings` - String manipulation
- `regexp` - Regular expressions
- `path/filepath` - File path handling
- `encoding/json` - JSON marshaling
- `crypto/sha256` - Hash generation
- `sync` - Concurrency primitives

**No External Test Dependencies:**
- We use only Go standard library for tests
- Table-driven test pattern
- No heavy test frameworks needed

---

## Frontend Technologies

### Alpine.js (Target Framework)
**Version:** 3.x
**Role:** Client-side reactive framework for generated HTML

**Why Alpine.js:**
- ✅ **Tiny** - ~15KB minified (vs 40KB+ for Svelte)
- ✅ **Declarative** - HTML-based directives
- ✅ **Progressive** - Works without build step
- ✅ **Familiar** - Vue-like syntax

**Generated Directives:**
```html
<!-- Our templates generate Alpine.js directives -->
<div x-data="{ count: 0 }">
  <span x-text="count"></span>
  <button @click="count++">Increment</button>
</div>
```

**Integration:**
- Custom Go Template → Alpine.js directives
- x-data for component state
- x-if/x-for for logic
- x-text/x-html for content

**No Build Step:**
- Alpine.js included via CDN or static file
- No bundling required
- Works directly in browser

---

## Build & Development Tools

### Agent OS Workflow
**Version:** Custom integration
**Role:** Development methodology and workflow

**Components:**
- **Specs** - Feature specifications in `.agent-os/specs/`
- **Standards** - Code quality rules in `.agent-os/standards/`
- **Agents** - Specialized AI agents (go-backend.md)

**Benefits:**
- Spec-driven development
- Cognitive load validation (<30 per file)
- Pattern enforcement
- Consistent code quality

---

### Go Tooling

#### Build System
```bash
# Native Go build (no Make/Bash required)
go build ./...
go build -o bin/server cmd/server/main.go
```

#### Testing
```bash
# Unit tests
go test ./... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# With race detector
go test ./... -race

# Benchmarks
go test ./... -bench=. -benchmem
```

#### Linting
```bash
# gofmt (built-in)
gofmt -s -w .

# go vet (built-in)
go vet ./...

# Optional: golangci-lint (not required)
golangci-lint run
```

---

### Version Control

#### Git
**Version:** 2.x+
**Configuration:**
- Conventional commits (informal)
- Co-author attribution (Claude Code)
- Branch: `main` (no feature branches currently)

**Workflow:**
```bash
# Commit with context
git commit -m "feat: add style aggregation with caching

Implements complete component style aggregation system...

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Plenti Integration Stack

### Plenti (Target Platform)
**Version:** 0.6.x+
**Role:** Static site generator being integrated with

**Architecture:**
- **CLI:** Go-based build system
- **Content:** JSON files in `content/`
- **Templates:** Currently Svelte (to be replaced)
- **Output:** Static HTML + hydrated SPA

**Integration Points:**
```go
// Our engine replaces these Plenti components:
cmd/build/compile.go    → cmd/build/render_templates.go
bundle.go               → Simplified (no JS bundling)
createHTML()           → Uses our renderer
```

**Magic Variables:**
```go
// Provided by Plenti to our templates
type MagicVars struct {
    Content    Content              // Current page
    AllContent []Content            // All pages
    AllLayouts map[string]Component // Components
    Env        Environment          // Config
    Params     url.Values          // Query params
}
```

---

## Architecture Patterns

### Pipeline Pattern
```
Template Source → Parser → AST → Transformer → Renderer → Output
```

**Each stage:**
- **Independent** - Can be tested in isolation
- **Composable** - Can be recombined
- **Typed** - Strong Go interfaces
- **Testable** - Minimal dependencies

### Package Structure
```
custom_go_template/
├── ast/              # AST node definitions
│   └── ast.go
├── parser/           # Template → AST
│   ├── parser.go
│   ├── expressions.go
│   └── *_test.go
├── transformer/      # AST → Alpine.js AST
│   ├── transformer.go
│   ├── components.go
│   └── *_test.go
├── renderer/         # AST → HTML/CSS/JS
│   ├── render.go
│   ├── styles.go     # Style aggregation
│   └── *_test.go
├── scoping/          # CSS/JS scoping utils
├── cmd/
│   └── server/       # Dev server
└── tests/            # Integration tests
```

---

## Testing Stack

### Test Framework
**Built-in Go testing** - No external frameworks

**Pattern:** Table-driven tests
```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {name: "basic", input: "{x}", want: `<span x-text="x">`, wantErr: false},
    {name: "error", input: "{", want: "", wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Parse(tt.input)
        // Assertions...
    })
}
```

### Test Coverage
```
Package         Coverage
---------       --------
parser          95%+
transformer     90%+
renderer        90%+
ast             100%
styles          100% (new)
```

### Test Categories

**Unit Tests:**
- Each package has `*_test.go` files
- Test individual functions
- Mock external dependencies
- Fast (<100ms total)

**Integration Tests:**
- `tests/alpine/` - Alpine.js transformation
- `tests/components/` - Component system
- Test full pipeline
- Moderate speed (~1s total)

**Manual Tests:**
- Dev server (`go run cmd/server/main.go`)
- Example pages in `examples/pages/`
- Visual verification in browser

---

## Performance Stack

### Benchmarking
```bash
go test ./renderer -bench=BenchmarkGetAggregatedStyles -benchmem
```

**Results:**
```
BenchmarkGetAggregatedStyles_CacheHit-8      643,537 ns/op
  1.86 μs per operation
  16 B/op
  1 allocs/op
```

### Profiling
```bash
# CPU profiling
go test -cpuprofile cpu.prof -bench .
go tool pprof cpu.prof

# Memory profiling
go test -memprofile mem.prof -bench .
go tool pprof mem.prof

# Trace
go test -trace trace.out
go tool trace trace.out
```

### Optimization Techniques
- **Preallocation** - `make([]T, 0, n)` when size known
- **Pooling** - `sync.Pool` for reusable buffers
- **Caching** - `sync.RWMutex` for thread-safe caches
- **Streaming** - Process templates without loading fully into memory

---

## Production Stack (Future)

### Deployment (Planned)

**Build Artifact:**
- Single Go binary (~10-20MB)
- No runtime dependencies
- Cross-platform (Linux, macOS, Windows)

**Docker (Optional):**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o /server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
```

### Observability (Planned)

**Logging:**
- Structured logging with `slog` (Go 1.21+)
- JSON output for production
- Contextual fields (request ID, component name)

**Metrics:**
- Build time metrics
- Cache hit rates
- Template processing times
- Memory usage

**Tracing:**
- OpenTelemetry integration (planned)
- Distributed tracing for complex builds

---

## Development Environment

### Recommended Setup

**Go Installation:**
```bash
# macOS
brew install go

# Verify
go version  # Should be 1.22+
```

**Editor:**
- VS Code with Go extension (recommended)
- GoLand (JetBrains IDE)
- Vim with vim-go
- Any editor with LSP support

**Go Tools:**
```bash
# Install useful tools
go install golang.org/x/tools/gopls@latest     # Language server
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Project Setup
```bash
# Clone
git clone <repo>
cd custom_go_template

# Install dependencies
go mod download

# Verify setup
go build ./...
go test ./...

# Run dev server
go run cmd/server/main.go
```

---

## Dependency Management

### Go Modules
**File:** `go.mod`

**Philosophy:**
- Minimal dependencies
- Prefer standard library
- Well-maintained packages only
- Regular updates for security

**Commands:**
```bash
# Add dependency
go get github.com/pkg/errors

# Update all
go get -u ./...

# Tidy (remove unused)
go mod tidy

# Vendor (optional)
go mod vendor
```

### Security
```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Or using govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Future Stack Additions

### Planned Additions

**Phase 1: Plenti Integration**
- No new dependencies needed
- Uses existing Plenti Go modules

**Phase 2: Component Prop Scoping**
- Possibly: AST manipulation library
- Maybe: Type checker for prop validation

**Phase 3: Production Hardening**
- OpenTelemetry SDK
- Prometheus client
- Error tracking SDK (Sentry, Rollbar)

**Phase 4: Ecosystem**
- GraphQL (for component discovery API)
- WebSocket (for hot reload)
- WASM (for browser-based preview)

### Will NOT Add

**Heavy Frameworks:**
- ❌ Web frameworks (gin, echo) - We're a library, not a server
- ❌ ORM - No database access needed
- ❌ DI frameworks - Simple constructor injection

**JavaScript Build Tools:**
- ❌ Webpack - No bundling needed
- ❌ Babel - No transpilation needed
- ❌ npm - Go modules sufficient

---

## Technology Decision Log

### Why Go Over...

**Rust:**
- ✅ Faster development (simpler syntax)
- ✅ Better for text processing
- ✅ Gentler learning curve
- ❌ Slightly slower (acceptable tradeoff)

**Node.js:**
- ✅ Better performance (compiled)
- ✅ Single binary deployment
- ✅ Type safety
- ✅ Better concurrency

**Python:**
- ✅ Much faster execution
- ✅ Type safety at compile time
- ✅ Better for tooling/CLI
- ✅ Easier deployment

### Why Alpine.js Over...

**Svelte:**
- ✅ Smaller runtime (~15KB vs ~40KB)
- ✅ No build step required
- ✅ Progressive enhancement friendly
- ❌ Less features (acceptable for our use case)

**React:**
- ✅ Much smaller (~15KB vs ~100KB+)
- ✅ No JSX needed
- ✅ Simpler mental model
- ✅ Better for server-rendered content

**Vue:**
- ✅ Smaller runtime
- ✅ Similar syntax (easier migration)
- ✅ Less opinionated
- ❌ Fewer ecosystem packages (acceptable)

### Why No Framework for Testing...

**Standard Go testing:**
- ✅ No dependencies
- ✅ IDE integration
- ✅ Fast
- ✅ Sufficient features
- ❌ Less DSL magic (actually a pro)

---

## Tech Stack Summary

| Layer | Technology | Version | Why |
|-------|-----------|---------|-----|
| **Language** | Go | 1.22.2 | Performance, type safety |
| **Runtime (template)** | goja | latest | JS expression evaluation |
| **Parsing** | tdewolff/parse | 2.7.15 | Fast HTML/CSS parsing |
| **HTML** | golang.org/x/net | 0.26.0 | DOM manipulation |
| **Output** | Alpine.js | 3.x | Tiny reactive framework |
| **Testing** | Go stdlib | - | No dependencies needed |
| **Workflow** | Agent OS | Custom | Spec-driven development |
| **Integration** | Plenti | 0.6.x | Static site generator |

---

## Upgrade Path

### Current → Future

**Go Version:**
- Current: 1.22.2
- Next: 1.23+ (when stable)
- Benefits: Better generics, improved perf

**Dependencies:**
- Review quarterly
- Security patches monthly
- Breaking changes evaluated carefully

**Alpine.js:**
- Current: 3.x
- Next: Stay on 3.x (stable)
- No plans to upgrade to 4.x yet

---

**Maintained By:** Benjamin Waller
**Review Cycle:** Quarterly
**Last Technology Review:** October 2025
