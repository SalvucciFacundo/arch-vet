# Contributing to `arch-vet`

Thank you for your interest in contributing to `arch-vet`!

## Development Prerequisites

* Go 1.22 or higher
* Git

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/SalvucciFacundo/arch-vet.git
   cd arch-vet
   ```

2. Download dependencies:
   ```bash
   go mod download
   ```

3. Run tests with race detection:
   ```bash
   go test -v -race ./...
   ```

4. Build the binary locally:
   ```bash
   go build -o bin/arch-vet ./cmd/arch-vet
   ```

---

## Adding a New Architecture Rule

1. Create a new file under `internal/rules/your_rule.go`.
2. Implement the `Rule` interface:
   ```go
   type YourRule struct{}

   func (r *YourRule) ID() string { return "ARCH004" }
   func (r *YourRule) Name() string { return "your-rule-name" }
   func (r *YourRule) Evaluate(ctx *Context) []Diagnostic { ... }
   ```
3. Register the rule in `internal/rules/engine.go`.
4. Add unit tests in `internal/rules/your_rule_test.go`.
5. Update `SPEC.md` and `README.md` rule catalogs.

---

## Commit Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):
* `feat:` A new feature or rule.
* `fix:` A bug fix.
* `docs:` Documentation changes.
* `ci:` CI/CD pipeline changes.
* `test:` Adding or updating tests.
