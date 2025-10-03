# Foundational Cognitive Load Patterns

## Overview

This document defines cognitive load validation patterns that all agents must follow to ensure code maintainability and reduce mental overhead. Each pattern is scored on a cognitive load scale of 1-10, where 10 represents maximum complexity.

## Pattern Categories

### 1. Go Backend Patterns

#### GO-ERROR-CONTEXT (Load: 3)
**Mistake:** Returning naked errors without context
```go
// ❌ BAD - No context
if err != nil {
    return err
}
```

**Correct:** Always wrap errors with function context
```go
// ✅ GOOD - Clear error context
if err != nil {
    return fmt.Errorf("functionName: failed to process %s: %w", id, err)
}
```

#### GO-DEFER-LOOP (Load: 7)
**Mistake:** Using defer inside loops
```go
// ❌ BAD - Defers accumulate until function returns
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close() // Accumulates!
}
```

**Correct:** Extract to separate function
```go
// ✅ GOOD - Each file closed immediately
for _, file := range files {
    if err := processFile(file); err != nil {
        return err
    }
}

func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    // process file
    return nil
}
```

#### GO-SLICE-PREALLOC (Load: 4)
**Mistake:** Appending to slice without preallocation
```go
// ❌ BAD - Multiple allocations
var results []Item
for _, item := range items {
    if item.IsValid() {
        results = append(results, item)
    }
}
```

**Correct:** Preallocate when size is known
```go
// ✅ GOOD - Single allocation
results := make([]Item, 0, len(items))
for _, item := range items {
    if item.IsValid() {
        results = append(results, item)
    }
}
```

#### GO-MAP-CONCURRENT (Load: 9)
**Mistake:** Concurrent map access without synchronization
```go
// ❌ BAD - Race condition
cache := make(map[string]string)
go func() { cache["key"] = "value" }()
go func() { val := cache["key"] }()
```

**Correct:** Use mutex or sync.Map
```go
// ✅ GOOD - Thread-safe
type Cache struct {
    mu sync.RWMutex
    m  map[string]string
}

func (c *Cache) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.m[key]
}
```

#### GO-NIL-CHECK (Load: 3)
**Mistake:** Checking slice for nil instead of length
```go
// ❌ BAD - Doesn't handle empty slice
if users == nil {
    return errors.New("no users")
}
```

**Correct:** Check length for empty check
```go
// ✅ GOOD - Handles both nil and empty
if len(users) == 0 {
    return errors.New("no users")
}
```

### 2. GoFast-Specific Patterns

#### GOFAST-SIMPLE-DI (Load: 5)
**Mistake:** Using complex DI containers instead of constructor injection
```go
// ❌ BAD - Complex DI container
container.Register("service", func() interface{} {
    return NewService(container.Get("repo").(Repository))
})
service := container.Get("service").(*Service)
```

**Correct:** Simple constructor injection following GoFast philosophy
```go
// ✅ GOOD - Simple, explicit dependency injection
func NewService(repo Repository, logger Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}

// Wire up in main.go or setup function
repo := repository.NewPostgresRepo(db)
logger := logging.NewLogger()
service := NewService(repo, logger)
```

#### GOFAST-STRATEGY-PATTERN (Load: 8)
**Mistake:** Complex inheritance hierarchies or excessive abstractions
```go
// ❌ BAD - Over-abstracted
type PaymentAbstractFactory interface {
    CreateProcessor() ProcessorBuilder
}

type ProcessorBuilder interface {
    WithConfig(cfg Config) ProcessorBuilder
    WithLogger(log Logger) ProcessorBuilder
    Build() PaymentProcessor
}
```

**Correct:** Simple interface with concrete implementations
```go
// ✅ GOOD - Simple strategy pattern
type PaymentProcessor interface {
    Process(ctx context.Context, amount int64) error
    Refund(ctx context.Context, transactionID string) error
}

// Concrete implementations
type StripeProcessor struct {
    apiKey string
}

func (s *StripeProcessor) Process(ctx context.Context, amount int64) error {
    // Direct implementation, no unnecessary abstraction
    return fmt.Errorf("StripeProcessor.Process: %w", err)
}

// Selection based on config
func NewPaymentProcessor(provider string) PaymentProcessor {
    switch provider {
    case "stripe":
        return &StripeProcessor{apiKey: config.StripeKey}
    case "lemon":
        return &LemonProcessor{apiKey: config.LemonKey}
    default:
        return &StripeProcessor{apiKey: config.StripeKey}
    }
}
```

#### GOFAST-ERROR-VALUES (Load: 3)
**Mistake:** Using panic/recover for error handling
```go
// ❌ BAD - Using panic for errors
func ProcessPayment(amount int64) {
    if amount <= 0 {
        panic("invalid amount")
    }
    // process...
}

// ❌ BAD - Ignoring errors
result, _ := db.Query("SELECT * FROM users")
```

**Correct:** Errors as values with proper handling
```go
// ✅ GOOD - Errors as values
func ProcessPayment(amount int64) error {
    if amount <= 0 {
        return fmt.Errorf("ProcessPayment: invalid amount %d", amount)
    }
    
    result, err := db.Query("SELECT * FROM users")
    if err != nil {
        return fmt.Errorf("ProcessPayment: query failed: %w", err)
    }
    defer result.Close()
    
    return nil
}

// Custom errors when needed
var (
    ErrInvalidAmount = errors.New("amount must be positive")
    ErrUserNotFound  = errors.New("user not found")
)
```

#### GOFAST-MINIMAL-DEPS (Load: 6)
**Mistake:** Adding external dependencies for simple tasks
```go
// ❌ BAD - External dependency for simple task
import "github.com/some/uuid-library"

func generateID() string {
    return uuidlib.New().String()
}
```

**Correct:** Use standard library when possible
```go
// ✅ GOOD - Standard library solution
import (
    "crypto/rand"
    "encoding/hex"
)

func generateID() string {
    bytes := make([]byte, 16)
    if _, err := rand.Read(bytes); err != nil {
        return ""
    }
    return hex.EncodeToString(bytes)
}

// Or use simple solution
func generateID() string {
    return fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Int63())
}
```

#### GOFAST-EXPLICIT-CONFIG (Load: 4)
**Mistake:** Hidden configuration or magic values
```go
// ❌ BAD - Hidden configuration
func SetupDB() *sql.DB {
    db, _ := sql.Open("postgres", "localhost:5432/mydb")
    db.SetMaxOpenConns(25) // Magic number!
    return db
}
```

**Correct:** Explicit configuration structure
```go
// ✅ GOOD - Explicit config following GoFast pattern
type DatabaseConfig struct {
    Host            string
    Port            int
    Database        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

func SetupDB(cfg DatabaseConfig) (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d dbname=%s", 
        cfg.Host, cfg.Port, cfg.Database)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("SetupDB: failed to open: %w", err)
    }
    
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    return db, nil
}
```

#### GOFAST-DTO-PATTERN (Load: 6)
**Mistake:** Exposing database types (sql.Null*) directly in API responses
```go
// ❌ BAD - Domain model exposed to API with sql.Null types
type Consultation struct {
    ID                   string         `json:"id"`
    CompletionPercentage sql.NullInt32  `json:"completion_percentage"`
    CreatedAt           sql.NullTime   `json:"created_at"`
}

func (h *Handler) GetConsultation(w http.ResponseWriter, r *http.Request) {
    consultation := h.service.GetConsultation(ctx, id)
    json.NewEncoder(w).Encode(consultation) // sql.Null* types in JSON!
}

// ❌ BAD - MarshalJSON workaround (technical debt)
func (c Consultation) MarshalJSON() ([]byte, error) {
    // Custom marshaling for sql.Null types
    // This is a workaround, not proper separation!
}
```

**Correct:** Separate DTO layer for API responses
```go
// ✅ GOOD - Domain model (no JSON tags)
type Consultation struct {
    ID                   string
    CompletionPercentage sql.NullInt32
    CreatedAt           sql.NullTime
}

// ✅ GOOD - DTO for API (no sql.Null*)
type ConsultationDTO struct {
    ID                   string    `json:"id"`
    CompletionPercentage *int32    `json:"completion_percentage,omitempty"`
    CreatedAt           time.Time `json:"created_at"`
}

// ✅ GOOD - Conversion method
func (c *Consultation) ToDTO() ConsultationDTO {
    dto := ConsultationDTO{
        ID:        c.ID,
        CreatedAt: c.CreatedAt.Time,
    }
    if c.CompletionPercentage.Valid {
        dto.CompletionPercentage = &c.CompletionPercentage.Int32
    }
    return dto
}

// ✅ GOOD - Handler uses DTO
func (h *Handler) GetConsultation(w http.ResponseWriter, r *http.Request) {
    consultation, err := h.service.GetConsultation(ctx, id)
    if err != nil {
        h.respondError(w, http.StatusNotFound, "not found")
        return
    }
    dto := consultation.ToDTO() // Convert to DTO
    json.NewEncoder(w).Encode(dto) // Clean JSON
}
```

### 3. Svelte/SvelteKit Patterns

#### SVELTE-STORE-LOOP (Load: 8)
**Mistake:** Direct store assignment in reactive context
```svelte
<!-- ❌ BAD - Creates infinite loop -->
<script>
  import { store } from './stores';
  let data = $state();
  
  onMount(() => {
    data = store.value; // Direct reference!
  });
  
  $effect(() => {
    store.update(data); // Circular dependency!
  });
</script>
```

**Correct:** Break reference with spread/clone
```svelte
<!-- ✅ GOOD - No circular dependency -->
<script>
  import { store } from './stores';
  let data = $state();
  let isInitializing = true;
  
  onMount(() => {
    data = { ...store.value }; // Break reference
    isInitializing = false;
  });
  
  $effect(() => {
    if (!isInitializing) {
      store.update(structuredClone(data));
    }
  });
</script>
```

#### SVELTE-INIT-GUARD (Load: 7)
**Mistake:** Missing initialization guard
```svelte
<!-- ❌ BAD - Functions run during init -->
<script>
  onMount(async () => {
    await loadData();
    updateUI(); // May run before ready
  });
</script>
```

**Correct:** Use initialization flag
```svelte
<!-- ✅ GOOD - Controlled initialization -->
<script>
  let isInitializing = true;
  
  onMount(async () => {
    await loadData();
    isInitializing = false;
    updateUI();
  });
  
  function updateUI() {
    if (isInitializing) return;
    // safe to update
  }
</script>
```

#### SVELTEKIT-SERVER-CLIENT (Load: 8)
**Mistake:** Browser API in server file
```typescript
// ❌ BAD - In +page.server.ts
export async function load() {
  const theme = localStorage.getItem('theme'); // ERROR!
  return { theme };
}
```

**Correct:** Use appropriate file or check environment
```typescript
// ✅ GOOD - In +page.ts
import { browser } from '$app/environment';

export async function load() {
  const theme = browser ? localStorage.getItem('theme') : 'light';
  return { theme };
}
```

#### SVELTEKIT-WATERFALL (Load: 6)
**Mistake:** Loading data in component instead of load function
```svelte
<!-- ❌ BAD - Causes waterfall loading -->
<script>
  let data = [];
  
  onMount(async () => {
    data = await fetch('/api/data').then(r => r.json());
  });
</script>
```

**Correct:** Use load function
```typescript
// ✅ GOOD - In +page.ts
export async function load({ fetch }) {
  const data = await fetch('/api/data').then(r => r.json());
  return { data };
}
```

## Cognitive Load Scoring

### Severity Levels
- **1-3**: Low complexity - Info level
- **4-6**: Medium complexity - Warning level  
- **7-10**: High complexity - Error level

### Total Score Thresholds
- **0-15**: Acceptable complexity
- **16-30**: Should simplify (warning)
- **31+**: Must refactor (block)

## Validation Rules

When an agent generates code:
1. Check against all applicable patterns
2. Calculate total cognitive load score
3. Block if score > 30
4. Warn if score > 15
5. Apply auto-fixes where available
6. Educate on violations

## Pattern Detection

Patterns are detected using:
- Regular expressions for simple patterns
- AST analysis for complex patterns
- Context-aware checking for framework-specific patterns

## Auto-fix Strategy

When possible, provide automatic fixes:
1. Simple replacements (e.g., error wrapping)
2. Structural changes (e.g., extracting functions)
3. Reference breaking (e.g., object spread)
4. Guard additions (e.g., initialization flags)
