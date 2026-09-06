# Specification: `arch-vet`

**AI-Native Architecture Linter & Guardrail for Go**

* **Repository:** `github.com/SalvucciFacundo/arch-vet`
* **Status:** Draft / Specification
* **Target Language:** Go 1.22+
* **Primary Interfaces:** CLI, CI/CD Gate, MCP Server (Model Context Protocol)

---

## 1. Executive Summary & Vision

AI coding assistants (Claude Code, Cursor, Copilot, Aider) write code quickly but frequently degrade application architecture over time. They introduce circular dependencies, breach layer boundaries, leak database queries into business logic, and import infrastructure drivers directly into core domain packages.

Existing architecture linters in Go (such as `go-arch-lint` and `arch-go`) were built for the pre-AI era:
1. They require verbose, manual YAML files mapping every single folder and package.
2. They rely strictly on lexical package import checks, ignoring deeper Go semantic contracts (e.g., consumer-side interfaces, abstraction leaks).
3. They output terse failure messages designed for humans in CI terminals, lacking the structured context and actionable remediation instructions that LLMs need for self-correction.

`arch-vet` is designed to be the idiomatic, zero-config architecture guardrail for Go in the AI era. It operates as both a blazing-fast CLI for developers/CI and an MCP server that provides AI agents with instant feedback and self-healing recommendations before changes are committed.

---

## 2. Core Principles

1. **Zero-Config by Default:** Automatically infer architecture based on common repository layout conventions (`internal/core`, `internal/domain`, `internal/adapters`, `pkg/`, etc.) with built-in presets (Hexagonal, Clean, Layered).
2. **Beyond Import Checking:** Validate Go-specific architectural conventions:
   * Interfaces defined on the consumer side, not producer side.
   * Domain packages must not expose or return infrastructure types (e.g., `*sql.DB`, `gorm.DB`, HTTP response writers).
3. **Agent Self-Healing:** Diagnostic outputs provide not only the file and line of the violation, but also the architectural rationale and exact remediation code examples so that AI agents can fix their own mistakes autonomously.
4. **Zero Heavy Dependencies:** Written purely in Go using `go/packages`, `go/ast`, and standard tooling. Runs in milliseconds without long-running background daemons or heavy external runtimes.

---

## 3. Functional Requirements

### 3.1 Architecture Presets & Configuration
* **Auto-Discovery:** When run without a configuration file, `arch-vet` inspects the directory structure and selects the most appropriate preset (e.g., Hexagonal / Ports & Adapters, Clean Architecture, or Standard Layered).
* **Presets:**
  * `hexagonal`: Strict separation between `domain` (entities, values), `ports` (interfaces), `application` (use cases), and `adapters` (inbound/outbound infrastructure).
  * `clean`: Entities $\rightarrow$ Use Cases $\rightarrow$ Interface Adapters $\rightarrow$ Frameworks & Drivers.
  * `custom`: Defined in a lightweight `.arch-vet.yaml` for bespoke boundaries.
* **Exceptions / Waivers:** Ability to ignore specific legacy violations or annotate files with `// arch-vet:ignore <rule>` comments during gradual migrations.

### 3.2 Semantic Rules Engine
`arch-vet` implements modular rules evaluated over the AST and type graphs:

| Rule ID | Name | Description | Severity |
| :--- | :--- | :--- | :--- |
| `ARCH001` | `forbidden-import` | A layer imports a forbidden outer layer or external infrastructure driver. | Error |
| `ARCH002` | `domain-leaky-type` | Domain structs or function signatures contain infrastructure types (SQL, HTTP, vendor SDKs). | Error |
| `ARCH003` | `producer-interface` | Interfaces defined alongside their concrete implementations rather than where they are consumed. | Warning |
| `ARCH004` | `circular-layer-dep` | Dependency cycles detected between defined architectural components. | Error |
| `ARCH005` | `bypassed-boundary` | Direct cross-layer access bypassing application ports or service layers. | Error |

### 3.3 Reporting & Output Formats
* **Human CLI (`text`):** High-contrast terminal output with color-coded severity, file locations, and diff-style context.
* **JSON (`--json`):** Structured machine-readable schema for automated pipelines.
* **AI-Agent Format (`--ai`):** Markdown-rich diagnostic payload optimized for LLM consumption, outlining:
  1. Root cause summary.
  2. The broken architectural constraint.
  3. Step-by-step refactoring guidance.
* **SARIF (`--sarif`):** Standard format for GitHub Code Scanning and CI integration.

### 3.4 MCP Server Interface
`arch-vet mcp` starts a stdio-based Model Context Protocol server exposing:
* `verify_architecture`: Runs full architectural verification on the workspace and returns structured violations with self-healing advice.
* `check_diff`: Inspects only files modified in the active Git working tree or branch to provide instant feedback during editing cycles.
* `get_architectural_map`: Returns the detected layers, boundaries, and rules governing the repository to ground the AI before it begins writing code.

---

## 4. System Architecture & Package Layout

```
arch-vet/
├── cmd/
│   └── arch-vet/
│       └── main.go              # Entry point: CLI routing & flags
├── internal/
│   ├── analyzer/
│   │   ├── loader.go            # Package loading via golang.org/x/tools/go/packages
│   │   ├── ast.go               # AST traversal & symbol inspection
│   │   └── graph.go             # Dependency graph builder
│   ├── config/
│   │   ├── config.go            # Configuration model (.arch-vet.yaml)
│   │   ├── detector.go          # Layout auto-detection
│   │   └── presets.go           # Built-in presets (Hexagonal, Clean, Layered)
│   ├── rules/
│   │   ├── engine.go            # Rule registry and evaluation coordinator
│   │   ├── rule.go              # Core Rule interface & Diagnostic struct
│   │   ├── forbidden_import.go  # ARCH001 implementation
│   │   ├── domain_leak.go       # ARCH002 implementation
│   │   └── producer_iface.go    # ARCH003 implementation
│   ├── reporter/
│   │   ├── reporter.go          # Output format dispatch
│   │   ├── terminal.go          # Human-readable color terminal reporter
│   │   ├── json.go              # Machine-readable JSON
│   │   └── agent.go             # LLM prompt/remediation formatter
│   └── mcp/
│       ├── server.go            # MCP stdio protocol handler
│       └── tools.go             # Tool implementations (verify, check_diff, get_map)
├── go.mod
├── go.sum
└── SPEC.md
```

---

## 5. Implementation Milestones

### Phase 1: Core Foundation & Import Boundary Engine
* [x] Initialize Go module (`github.com/SalvucciFacundo/arch-vet`).
* [x] Implement `analyzer/loader.go` using `go/packages` to construct package dependency graphs.
* [x] Implement configuration model and Hexagonal preset in `config/`.
* [x] Implement `ARCH001` (`forbidden-import`).
* [x] Implement terminal, JSON, and agent reporting.

### Phase 2: Semantic Checks & Agent Output
* [x] Implement `ARCH002` (`domain-leaky-type`) using AST struct and function signature inspection.
* [x] Implement `ARCH003` (`producer-interface`) for consumer-side interface enforcement.
* [x] Implement `reporter/agent.go` to produce structured self-healing guidance.

### Phase 3: Model Context Protocol (MCP) Server
* [x] Implement JSON-RPC 2.0 / MCP stdio server in `internal/mcp`.
* [x] Expose `verify_architecture`, `get_architectural_map`, and `explain_rule` tools.
* [x] Add automated end-to-end testing with mock Go projects.

### Phase 4: CI/CD & Community Release
* [x] Add GitHub Actions CI workflow (test matrix across Go versions with `-race`).
* [x] Add multi-platform release automation via GoReleaser (`.goreleaser.yaml` and tag-triggered workflow).
* [x] Publish documentation and installation guide (`go install`).

---

## 6. Future Roadmap: v0.2.0 Planned Enhancements

* [x] **Fast Git Diff Mode (`--diff` & `check_diff` MCP tool):** Inspect only files modified in the active Git index/working tree for sub-50ms feedback loops.
* [x] **Inline Suppression Comments (`// arch-vet:ignore <rule>`):** Allow progressive refactoring and debt management in legacy codebases.
* [ ] **SARIF Report Generator (`--format sarif`):** Native integration with GitHub Code Scanning to render line-level annotations on Pull Requests.
