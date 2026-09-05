# arch-vet

> **AI-Native Architecture Linter & Guardrail for Go**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`arch-vet` protects your Go codebase from architectural rot and boundary violations during rapid development and autonomous AI agent coding sessions (Claude Code, Cursor, Copilot, Aider).

Unlike legacy linters, `arch-vet` features **zero-config preset auto-detection** and emits **actionable self-healing remediation guidance** designed for LLMs to autonomously correct architectural drift before code is committed.

---

## Key Features

* 🚀 **Zero-Config Presets:** Automatically detects your architectural layout (`hexagonal`, `clean`, `layered`). No mandatory 100-line YAML files.
* 🛡️ **Boundary Enforcement (`ARCH001`):** Prevents domain layers from importing outer infrastructure drivers, web frameworks, or databases.
* 🤖 **Agent Self-Healing Format:** Emits rich diagnostic markdown (`--format agent`) with root-cause explanations and step-by-step refactoring instructions for AI coding agents.
* ⚡ **Ultra-Fast & Pure Go:** Built directly on `go/packages` and `go/ast`. Runs in milliseconds without long-running daemons.
* 🔌 **CI/CD Quality Gate:** Exits with code `1` on architectural violations, colorized terminal reporting, and machine-readable JSON.

---

## Installation

```bash
go install github.com/SalvucciFacundo/arch-vet/cmd/arch-vet@latest
```

---

## Quick Start

### 1. Run with Auto-Detection

Run inside any Go project. `arch-vet` inspects your package structure and applies the best matching preset:

```bash
arch-vet -dir .
```

### 2. Specify an Architecture Preset

```bash
# Hexagonal (Ports & Adapters)
arch-vet -preset hexagonal

# Clean Architecture
arch-vet -preset clean
```

### 3. Agent Mode (For AI Coding Agents)

Produce self-healing instructions tailored for LLM context windows:

```bash
arch-vet -format agent
```

Example agent output:
```markdown
## Architecture Violations Detected (1)

Please fix the following architectural regressions before finalizing your changes:

### 1. [ARCH001] forbidden-import
* **Location:** `internal/domain/user.go:14:2`
* **Problem:** layer "domain" cannot import external package "database/sql"
* **Architectural Rationale:** Layer "domain" is an inner layer that must remain pure and free from concrete infrastructure drivers.
* **Self-Healing Action:** Define an abstraction or interface in layer "domain" or an adjacent ports layer, and inject the concrete implementation from an outer adapter layer.
```

### 5. MCP Server (For Claude Code, Cursor & Antigravity)

`arch-vet` provides a native Model Context Protocol (MCP) server over `stdio` so that AI coding agents can verify architecture and read boundaries dynamically.

#### Start the Server:
```bash
arch-vet mcp
```

#### Add to Claude Code or Cursor (`~/.config/claude/claude_desktop_config.json` or agent settings):
```json
{
  "mcpServers": {
    "arch-vet": {
      "command": "arch-vet",
      "args": ["mcp"]
    }
  }
}
```

Exposed MCP Tools:
* `verify_architecture`: Audits the workspace and returns structured violations with actionable self-healing steps.
* `get_architectural_map`: Gives the AI the blueprint of layers and allowed dependency flows before it writes code.
* `explain_rule`: Explains the architectural principle and shows idiomatic Go refactoring patterns.

---

## Rule Catalog

| Rule ID | Name | Description | Severity |
| :--- | :--- | :--- | :--- |
| `ARCH001` | `forbidden-import` | Package imports a forbidden layer or external infrastructure driver | Error |
| `ARCH002` | `domain-leaky-type` | Domain structs or signatures expose infrastructure types | Error |
| `ARCH003` | `producer-interface` | Interfaces declared inside adapter/infra instead of consumer package | Warning |

---

## Roadmap

- [x] **Phase 1:** Core AST loader, zero-config presets, `ARCH001` rule, terminal/json/agent reporting.
- [x] **Phase 2:** Semantic deep rules (`ARCH002` leaky types, `ARCH003` producer interfaces).
- [x] **Phase 3:** Built-in Model Context Protocol (`arch-vet mcp`) stdio server for native Claude Code and Cursor integration.
- [ ] **Phase 4:** GitHub Action and SARIF report generation.

---

## License

MIT
