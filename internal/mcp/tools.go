package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
	"github.com/SalvucciFacundo/arch-vet/internal/reporter"
	"github.com/SalvucciFacundo/arch-vet/internal/rules"
	"gopkg.in/yaml.v3"
)

// ListTools returns the catalog of tools exposed by arch-vet.
func ListTools() []Tool {
	return []Tool{
		{
			Name:        "verify_architecture",
			Description: "Verify architectural layer boundaries and detect leaks in a Go project with actionable AI self-healing guidance.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Root directory of the Go project (defaults to current working directory).",
					},
					"preset": map[string]interface{}{
						"type":        "string",
						"description": "Architecture preset to enforce (hexagonal, clean). Defaults to auto-detection.",
					},
				},
			},
		},
		{
			Name:        "get_architectural_map",
			Description: "Get the active architectural layer map, conventions, and dependency rules of the Go project to ground the AI before generating code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Root directory of the Go project.",
					},
				},
			},
		},
		{
			Name:        "explain_rule",
			Description: "Explain an architectural rule with rationale and concrete Go refactoring examples.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"rule_id": map[string]interface{}{
						"type":        "string",
						"description": "Rule ID to inspect (e.g. ARCH001, ARCH002, ARCH003).",
					},
				},
				"required": []string{"rule_id"},
			},
		},
	}
}

// HandleVerifyArchitecture executes full architecture verification on a directory.
func HandleVerifyArchitecture(args json.RawMessage) (*CallToolResult, error) {
	var params struct {
		Dir    string `json:"dir"`
		Preset string `json:"preset"`
	}
	_ = json.Unmarshal(args, &params)
	if params.Dir == "" {
		params.Dir = "."
	}

	// 1. Resolve configuration
	var cfg *config.Config
	configFile := ".arch-vet.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var loaded config.Config
			if err := yaml.Unmarshal(data, &loaded); err == nil {
				cfg = &loaded
			}
		}
	}

	if cfg == nil {
		if params.Preset != "" {
			presetCfg, err := config.GetPreset(params.Preset)
			if err != nil {
				return &CallToolResult{
					IsError: true,
					Content: []ToolResultContent{{Type: "text", Text: err.Error()}},
				}, nil
			}
			cfg = presetCfg
		} else {
			cfg = config.DetectArchitecture(params.Dir)
		}
	}

	// 2. Load project packages
	project, err := analyzer.LoadProject(params.Dir)
	if err != nil {
		return &CallToolResult{
			IsError: true,
			Content: []ToolResultContent{{Type: "text", Text: fmt.Sprintf("Failed to load Go packages in %s: %v", params.Dir, err)}},
		}, nil
	}

	// 3. Evaluate rules
	ctx := &rules.Context{Config: cfg, Project: project}
	engine := rules.NewEngine()
	diagnostics := engine.Evaluate(ctx)

	// 4. Format report using AgentReporter
	var buf bytes.Buffer
	agentRep := &reporter.AgentReporter{}
	_ = agentRep.Report(&buf, diagnostics)

	return &CallToolResult{
		IsError: len(diagnostics) > 0,
		Content: []ToolResultContent{{Type: "text", Text: buf.String()}},
	}, nil
}

// HandleGetArchitecturalMap returns a grounded summary of layers and allowed flows.
func HandleGetArchitecturalMap(args json.RawMessage) (*CallToolResult, error) {
	var params struct {
		Dir string `json:"dir"`
	}
	_ = json.Unmarshal(args, &params)
	if params.Dir == "" {
		params.Dir = "."
	}

	cfg := config.DetectArchitecture(params.Dir)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Architectural Map (%s preset)\n\n", cfg.Architecture)
	fmt.Fprintln(&buf, "Ground your code generation in these defined boundaries:")
	fmt.Fprintln(&buf)

	for _, layer := range cfg.Layers {
		fmt.Fprintf(&buf, "### Layer: `%s`\n", layer.Name)
		fmt.Fprintf(&buf, "* **Role:** %s\n", layer.Description)
		fmt.Fprintf(&buf, "* **Patterns:** `%v`\n", layer.Patterns)
		if len(layer.MayDependOn) == 0 {
			fmt.Fprintln(&buf, "* **May Depend On:** None (Strictly pure layer, inward boundary)")
		} else {
			fmt.Fprintf(&buf, "* **May Depend On:** `%v`\n", layer.MayDependOn)
		}
		if len(layer.DenyExternal) > 0 {
			fmt.Fprintf(&buf, "* **Deny External:** `%v`\n", layer.DenyExternal)
		}
		fmt.Fprintln(&buf)
	}

	return &CallToolResult{
		Content: []ToolResultContent{{Type: "text", Text: buf.String()}},
	}, nil
}

// HandleExplainRule returns rich rationale for a specific architectural rule.
func HandleExplainRule(args json.RawMessage) (*CallToolResult, error) {
	var params struct {
		RuleID string `json:"rule_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil || params.RuleID == "" {
		return &CallToolResult{
			IsError: true,
			Content: []ToolResultContent{{Type: "text", Text: "rule_id is required"}},
		}, nil
	}

	switch params.RuleID {
	case "ARCH001":
		return &CallToolResult{
			Content: []ToolResultContent{{
				Type: "text",
				Text: `## ARCH001: forbidden-import
### Rationale
In Hexagonal and Clean architectures, dependencies must point inward toward business logic. Core business rules (domain) must never import outer infrastructure (database drivers, HTTP frameworks, external message queues).

### Remediation Pattern
Apply Dependency Inversion Principle (DIP):
1. Define an interface in the consuming domain or ports layer:
   ` + "```go\ntype UserRepository interface {\n    Find(ctx context.Context, id string) (*User, error)\n}\n```" + `
2. Implement the interface inside the adapter layer (` + "`internal/adapters/db`" + `):
   ` + "```go\ntype PostgresUserRepo struct { ... }\n```" + `
3. Inject the interface into your service via constructor injection.`,
			}},
		}, nil

	case "ARCH002":
		return &CallToolResult{
			Content: []ToolResultContent{{
				Type: "text",
				Text: `## ARCH002: domain-leaky-type
### Rationale
Even without importing forbidden packages, structs or function parameters in core layers must not leak infrastructure types (e.g. *sql.DB, gin.Context, *http.Request). Doing so couples business logic to transport or persistence layers.

### Remediation Pattern
1. In use cases, accept domain command/query DTOs instead of HTTP requests:
   ` + "```go\n// Bad\nfunc (s *Service) Create(c *gin.Context) error\n\n// Good\nfunc (s *Service) Create(ctx context.Context, cmd CreateUserCommand) (*User, error)\n```",
			}},
		}, nil

	case "ARCH003":
		return &CallToolResult{
			Content: []ToolResultContent{{
				Type: "text",
				Text: `## ARCH003: producer-interface
### Rationale
In idiomatic Go, interfaces belong in the package that uses values of the interface type, not the package that implements those values. Declaring interfaces in adapter packages creates unnecessary coupling.

### Remediation Pattern
Define interfaces next to their callers (consumer package) and export concrete structs from adapter packages.`,
			}},
		}, nil

	default:
		return &CallToolResult{
			IsError: true,
			Content: []ToolResultContent{{Type: "text", Text: fmt.Sprintf("Unknown rule ID %q", params.RuleID)}},
		}, nil
	}
}
