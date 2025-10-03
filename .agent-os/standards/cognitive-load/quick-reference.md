# Cognitive Load Quick Reference

## Immediate Action Items for All Agents

### 🔴 ALWAYS Block (Score 7-10)
- Direct store assignment in reactive contexts
- Defer inside loops  
- Concurrent map access without sync
- Browser APIs in +page.server.ts files
- Complex DI containers (GoFast violation)
- Over-abstracted patterns (GoFast violation)
- sql.Null* types in API responses (DTO violation)
- Missing ToDTO() conversions

### 🟡 Fix Immediately (Score 4-6)
- Naked error returns without context
- Missing slice preallocation
- Data fetching in onMount instead of load
- Missing initialization guards
- External deps for simple tasks (GoFast)
- Hidden configuration values (GoFast)
- Incomplete pattern implementation

### 🟢 Best Practices (Score 1-3)
- Use fmt.Errorf for error wrapping
- Check len() not nil for slices
- Break store references with spread
- Use load functions for data
- Errors as values (GoFast philosophy)
- Simple constructor DI (GoFast philosophy)
- Proper DTO layer for APIs

## Go Quick Fixes

```go
// Error Context
- if err != nil { return err }
+ if err != nil { return fmt.Errorf("funcName: context: %w", err) }

// Slice Preallocation  
- var results []Type
+ results := make([]Type, 0, knownSize)

// Defer in Loop
- for _, f := range files {
-   file, _ := os.Open(f)
-   defer file.Close()
- }
+ for _, f := range files {
+   processFile(f)
+ }
+ func processFile(f string) error {
+   file, _ := os.Open(f)
+   defer file.Close()
+   // process
+ }

// Concurrent Map
- m := make(map[string]string)
+ type SafeMap struct {
+   mu sync.RWMutex
+   m  map[string]string
+ }

// GoFast: Simple DI
- container.Register("service", serviceFactory)
+ func NewService(repo Repository) *Service {
+   return &Service{repo: repo}
+ }

// GoFast: Error Values
- panic("invalid input")
+ return fmt.Errorf("invalid input: %w", ErrInvalidInput)

// GoFast: Minimal Deps
- import "github.com/uuid"
- id := uuid.New()
+ import "crypto/rand"
+ bytes := make([]byte, 16)
+ rand.Read(bytes)

// GoFast: Explicit Config
- db.SetMaxOpenConns(25) // Magic number
+ db.SetMaxOpenConns(cfg.MaxOpenConns)
```

## Svelte/SvelteKit Quick Fixes

```svelte
<!-- Store Reference -->
- onMount(() => { data = store.value })
+ onMount(() => { data = { ...store.value } })

<!-- Init Guard -->
+ let isInitializing = true
  onMount(() => {
    // setup
+   isInitializing = false
  })

<!-- Load vs Component -->
- <!-- In component -->
- onMount(async () => {
-   data = await fetch('/api').then(r => r.json())
- })
+ // In +page.ts
+ export const load = async ({ fetch }) => {
+   const data = await fetch('/api').then(r => r.json())
+   return { data }
+ }

<!-- Environment Check -->
- // In +page.server.ts
- const theme = localStorage.getItem('theme')
+ // In +page.ts  
+ import { browser } from '$app/environment'
+ const theme = browser ? localStorage.getItem('theme') : 'light'
```

## Validation Commands

Before writing code:
```bash
# Check cognitive load score
grep -E "return err|defer.*for|var.*\[\].*append" file.go

# Count violations
grep -c "return err\s*}" **/*.go
```

## Score Calculation

```
Total Score = Sum of:
- Each pattern violation × its load score
- Nesting depth ^ 2
- Magic numbers × 2
- Missing comments × 1

If score > 30: STOP and refactor
If score > 15: Consider simplifying
If score < 15: Good to proceed
```

## Resources

- Full patterns: `.agent-os/standards/cognitive-load/foundational-patterns.md`
- Config: `.agent-os/standards/cognitive-load/config.yml`
- Validation rules: `.agent-os/standards/cognitive-load/validation-system.md`
