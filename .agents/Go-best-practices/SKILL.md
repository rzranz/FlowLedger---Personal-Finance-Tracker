# AGENTS.md — Go (Golang) Development & Best Practices Guide

This document defines the coding standards, architectural patterns, and operational workflows for AI coding agents working on this Go repository. Every change must adhere to idiomatic Go principles: **simplicity, readability, and explicitness**.

## 1. Executable Commands (Run & Verify First)

Always validate your changes using these exact commands before completing a task:

```bash
# Format code and manage imports
gofmt -s -w .
goimports -w .

# Tidy and verify module dependencies
go mod tidy

# Static analysis and linting
go vet ./...
golangci-lint run ./...

# Run all tests with race detector enabled
go test -race -v ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...
```

## 2. Agent Operational Boundaries

* **ALWAYS:**
  * Format code with `gofmt` / `goimports` after editing `.go` files.
  * Run `go test -race ./...` and `go vet ./...` to verify changes.
  * Wrap errors with context using `fmt.Errorf("...: %w", err)`.
  * Pass `context.Context` as the first argument in functions performing I/O, network calls, or long-running operations.
  * Write table-driven tests for new or modified functions.
  * Apply Clean Code and idiomatic Go design patterns (Dependency Injection, Functional Options, Guard Clauses).

* **ASK FIRST:**
  * Before adding new third-party dependencies (`go get`). Prefer the Go standard library whenever possible.
  * Before modifying database migrations, public API contracts, or CI/CD pipelines.
  * Before changing exported package interfaces used across multiple modules.

* **NEVER:**
  * Ignore errors using `_` unless explicitly justified in a comment (e.g., `_ = resp.Body.Close()`).
  * Use `panic()` in library or business logic code (reserve `panic` only for unrecoverable startup/initialization failures).
  * Edit auto-generated files manually (`*.pb.go`, `*_gen.go`, `mock_*.go`).
  * Commit secrets, `.env` files, or hardcoded credentials.

## 3. Code Formatting & Style

Go enforces strict formatting via official tools, eliminating manual style debates.

* All code **must** be formatted using `gofmt` (or `goimports` to also sort and manage imports).
* Use `gofmt -s` to simplify code constructs automatically where applicable.

## 4. Naming Conventions

Go naming is concise, explicit, and uses capitalization to control visibility (Exported vs. Unexported).

```go
// Package name: lowercase single word, short, clear, no underscores
package httputil

// Exported identifiers: MixedCaps (PascalCase)
func NewClient() *Client {}

// Unexported identifiers: mixedCaps (camelCase)
func parseURL(rawURL string) error {}

// Constants: MixedCaps or mixedCaps (NEVER use SCREAMING_SNAKE_CASE)
const MaxRetryCount = 3
const defaultTimeout = 30 * time.Second

// Interfaces: single-method interfaces use the "-er" suffix
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

* **Acronyms & Initialisms:** Keep consistent casing for initialisms (`URL`, `HTTP`, `ID`, `JSON`). Write `serveHTTP`, `userID`, `xmlParser`, or `HTTPClient` (not `serveHttp`, `userId`, or `HttpClient`).
* **Avoid Stutter:** Do not repeat the package name in the exported type or function name. Use `httputil.NewClient()`, not `httputil.NewHTTPUtilClient()`.
* **Variable Length:** Scope dictates name length. Use short names (`i`, `v`, `ctx`, `err`, `r`, `w`) for small scopes, and descriptive names for wider scopes.

## 5. Package Design & Project Structure

```go
// Package httputil provides HTTP utility functions for common web operations.
package httputil

// Import grouping: 1. Standard library, 2. Third-party, 3. Local packages
// Separated by a single blank line.
import (
    "context"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"

    "myproject/internal/config"
)

// Define interfaces in the CONSUMER package, not the producer/implementation package.
// Principle: "Accept interfaces, return structs."
type UserReader interface {
    GetUser(ctx context.Context, id string) (*User, error)
}
```

* **Meaningful Package Names:** Avoid generic, catch-all package names like `util`, `common`, `misc`, `helpers`, or `base`. Name packages after what they *provide* (e.g., `auth`, `storage`, `httputil`).
* **Standard Layout:**
  * `cmd/<app>/main.go`: Application entry points (keep `main` thin; wire dependencies here).
  * `internal/`: Private application and library code that cannot be imported by external projects.
  * `pkg/`: Library code safe for external consumption (use sparingly).
* **Avoid `init()` and Global State:** Do not use package-level mutable variables or complex `init()` functions. Use explicit constructor functions (`New...`) and dependency injection.

## 6. Clean Code & Idiomatic Design Patterns

Apply Clean Code principles adapted to Go idioms—avoid over-engineering or forcing Java/C# OOP patterns into Go.

### 6.1 Clean Architecture & Dependency Inversion (SOLID)
Separate business logic (use cases/domain) from delivery mechanisms (HTTP, gRPC) and infrastructure (PostgreSQL, Redis). Depend on abstractions (interfaces) owned by the consumer.

```go
// internal/order/service.go (Consumer defines the small, focused interface)
package order

type Repository interface {
    Save(ctx context.Context, order *Order) error
}

type PaymentGateway interface {
    Charge(ctx context.Context, amount int64) error
}

// Explicit Dependency Injection via Constructor
type Service struct {
    repo    Repository
    payment PaymentGateway
}

func NewService(repo Repository, payment PaymentGateway) *Service {
    return &Service{
        repo:    repo,
        payment: payment,
    }
}
```

### 6.2 Functional Options Pattern
Use Functional Options for clean, extensible struct configuration when constructors have optional parameters. Avoid telescoping constructors or bloated config structs with ambiguous zero values.

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func NewServer(host string, opts ...Option) *Server {
    // Sensible defaults
    srv := &Server{
        host:    host,
        port:    8080,
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(srv)
    }
    return srv
}
```

### 6.3 Composition Over Inheritance
Go does not have classes or inheritance. Use struct embedding and small interface composition (`io.ReadWriter`) to share behavior cleanly.

```go
type AuditMetadata struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Invoice struct {
    ID     string
    Amount int64
    AuditMetadata // Embedded composition
}
```

### 6.4 Strategy & Adapter Patterns via Implicit Interfaces
Because Go interfaces are satisfied implicitly, implement Strategy and Adapter patterns using lightweight interfaces or first-class functions without boilerplate hierarchies.

```go
// Strategy pattern using interface
type Notifier interface {
    Send(ctx context.Context, message string) error
}

// Adapter function pattern (like http.HandlerFunc)
type NotifierFunc func(ctx context.Context, message string) error

func (f NotifierFunc) Send(ctx context.Context, message string) error {
    return f(ctx, message)
}
```

### 6.5 Guard Clauses & Line of Sight ("Happy Path" Left-Aligned)
Keep the normal execution path aligned to the left edge. Avoid `else` blocks after `return` or `continue`, and avoid deep nesting (max 2–3 levels).

```go
// GOOD: Early returns keep the happy path flat and readable
func RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
    if req.Email == "" {
        return nil, ErrInvalidEmail
    }
    if req.Age < 18 {
        return nil, ErrUnderage
    }

    user := &User{Email: req.Email, Age: req.Age}
    if err := saveUser(ctx, user); err != nil {
        return nil, fmt.Errorf("saving user: %w", err)
    }

    return user, nil
}
```

### 6.6 Make the Zero Value Useful
Design structs so their default zero-initialized state works safely without requiring initialization whenever possible (like `sync.Mutex` or `bytes.Buffer`).

## 7. Error Handling Patterns

Errors in Go are values. Handle them explicitly, early, and with rich context.

```go
var ErrUserNotFound = errors.New("user not found")

// Standard error wrapping pattern
func ReadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("reading config file %s: %w", filename, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }

    return &cfg, nil
}
```

* **Wrap Errors:** Always use `fmt.Errorf("doing something: %w", err)` to preserve the error chain. Use lowercase messages without trailing punctuation.
* **Inspect Errors:** Always use `errors.Is(err, TargetErr)` and `errors.As(err, &targetType)` instead of direct `==` comparisons or type assertions.
* **Handle Once:** Either log an error **or** return it—never both. Logging and returning creates duplicate noisy logs up the call stack.

## 8. Function and Method Design

```go
type User struct {
    Name string
    Age  int
}

// Value receiver: use when the method does not mutate the receiver and the struct is small
func (u User) String() string {
    return fmt.Sprintf("%s (%d)", u.Name, u.Age)
}

// Pointer receiver: use when mutating state, for large structs, or structs with sync primitives
func (u *User) UpdateAge(age int) {
    u.Age = age
}

// Multiple return values: context is ALWAYS first parameter; error is ALWAYS last return value
func ParseUser(ctx context.Context, data []byte) (User, error) {
    var u User
    err := json.Unmarshal(data, &u)
    return u, err
}
```

* **Single Responsibility Principle (SRP):** Each function should do one thing well and stay small enough to understand at a glance.
* **Receiver Consistency:** Receiver names must be 1–2 letters representing the type (e.g., `u` for `User`, `c` for `Client`) and must be consistent across all methods of that type (never mix `this` or `self`).
* **Concise Signatures:** If a function requires more than 3–4 parameters, group them into a request/options `struct`.

## 9. Concurrency Guidelines

Go's concurrency is built on goroutines, channels, and the `sync` package.

```go
// Explicitly manage goroutine lifecycle with context.Context
func processData(ctx context.Context, data <-chan string) <-chan Result {
    results := make(chan Result)

    go func() {
        defer close(results)
        for {
            select {
            case item, ok := <-data:
                if !ok || item == "" {
                    return
                }
                select {
                case results <- process(item):
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()

    return results
}
```

* **Zero Goroutine Leaks:** Never start a goroutine without knowing how and when it will stop. Always provide a cancellation mechanism (`context.Context` or a `done` channel).
* **Channel Ownership:** Only the sender should close a channel; never close a channel from the receiver side.
* **Choose the Right Primitive:**
  * Use **channels** for passing ownership of data, distributing work, or communicating async results.
  * Use **`sync.Mutex` / `sync.RWMutex`** for protecting simple shared state or caches inside a struct.
  * Use **`golang.org/x/sync/errgroup`** when running concurrent subtasks that need error propagation and context cancellation.

## 10. Comment & Documentation Conventions

Comments are parsed by `godoc` / `pkgsite` to generate official documentation.

```go
// Package math provides basic constants and mathematical functions.
package math

// Pi represents the mathematical constant pi.
const Pi = 3.14159265358979323846

// Sqrt returns the square root of x.
// It panics if x is negative.
func Sqrt(x float64) float64 {
    if x < 0 {
        panic("math: square root of negative number")
    }
    return 0
}
```

* Every exported type, const, var, and function **must** have a doc comment.
* Doc comments must be complete sentences that start with the exact name of the identifier being declared.
* Explain **why** something is done, not *what* the code is literally doing. Clean code should be self-documenting for the *what*.

## 11. Testing Conventions

```go
// user_test.go
func TestUser_UpdateAge(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        user     User
        newAge   int
        expected int
    }{
        {
            name:     "updates to valid positive age",
            user:     User{Name: "Alice", Age: 25},
            newAge:   30,
            expected: 30,
        },
        {
            name:     "updates to zero age",
            user:     User{Name: "Bob", Age: 20},
            newAge:   0,
            expected: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            tt.user.UpdateAge(tt.newAge)
            if tt.user.Age != tt.expected {
                t.Errorf("UpdateAge(%d) = %d; want %d", tt.newAge, tt.user.Age, tt.expected)
            }
        })
    }
}
```

* **Table-Driven Tests:** Use slice-of-structs with `t.Run()` subtests as the standard testing pattern.
* **Clear Failure Messages:** Format test failures as `got X, want Y` (e.g., `t.Errorf("Func() = %v, want %v", got, want)`).
* **Test Helpers:** Mark helper functions with `t.Helper()` so failure line numbers point to the test case, not the helper.

## 12. Performance & Modern Go Idioms

Write correct, clear code first, then optimize hot paths using `go test -bench=.` and `pprof`.

```go
// Preallocate slice capacity when size is known to avoid repeated allocations
func processItems(items []string) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// Use strings.Builder for efficient string concatenation in loops
func buildMessage(parts []string) string {
    var b strings.Builder
    b.Grow(estimateSize(parts))
    for _, part := range parts {
        b.WriteString(part)
    }
    return b.String()
}
```

* **Preallocate Containers:** Always pass a capacity hint to `make([]T, 0, cap)` and `make(map[K]V, cap)` when the size is known or estimable.
* **Structured Logging:** Prefer the standard library `log/slog` over unstructured `fmt.Println` or `log.Printf` for application logging.
* **Modern Standard Library:** Use built-in packages introduced in recent Go versions (`slices`, `maps`, `cmp`, `errors.Join`, `context.WithoutCancel`) rather than writing custom boilerplate loops.

## 13. Canonical References

When in doubt, follow the hierarchy of official and community-accepted Go standards:

1. [Effective Go](https://go.dev/doc/effective_go)
2. [Google Go Style Guide](https://google.github.io/styleguide/go/) & [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
3. [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)