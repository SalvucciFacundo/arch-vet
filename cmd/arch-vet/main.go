package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
	"github.com/SalvucciFacundo/arch-vet/internal/mcp"
	"github.com/SalvucciFacundo/arch-vet/internal/reporter"
	"github.com/SalvucciFacundo/arch-vet/internal/rules"
	"gopkg.in/yaml.v3"
)

var Version = "0.1.0-dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		server := mcp.NewServer(Version)
		if err := server.Serve(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	dirFlag := flag.String("dir", ".", "Root directory of the Go project")
	presetFlag := flag.String("preset", "", "Architecture preset (hexagonal, clean). Defaults to auto-detection")
	configFlag := flag.String("config", ".arch-vet.yaml", "Path to configuration file")
	formatFlag := flag.String("format", "text", "Output format (text, json, agent)")
	noColorFlag := flag.Bool("no-color", false, "Disable ANSI color output")
	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("arch-vet version %s\n", Version)
		os.Exit(0)
	}

	// 1. Resolve configuration
	var cfg *config.Config

	if _, err := os.Stat(*configFlag); err == nil {
		data, err := os.ReadFile(*configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file %s: %v\n", *configFlag, err)
			os.Exit(2)
		}
		var loadedCfg config.Config
		if err := yaml.Unmarshal(data, &loadedCfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing config file %s: %v\n", *configFlag, err)
			os.Exit(2)
		}
		cfg = &loadedCfg
	} else if *presetFlag != "" {
		presetCfg, err := config.GetPreset(*presetFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading preset: %v\n", err)
			os.Exit(2)
		}
		cfg = presetCfg
	} else {
		// Auto-detect architecture preset
		cfg = config.DetectArchitecture(*dirFlag)
	}

	// 2. Analyze project
	project, err := analyzer.LoadProject(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project: %v\n", err)
		os.Exit(2)
	}

	// 3. Evaluate rules
	ctx := &rules.Context{
		Config:  cfg,
		Project: project,
	}

	engine := rules.NewEngine()
	diagnostics := engine.Evaluate(ctx)

	// 4. Report results
	var rep reporter.Reporter
	switch *formatFlag {
	case "json":
		rep = &reporter.JSONReporter{}
	case "agent", "ai":
		rep = &reporter.AgentReporter{}
	default:
		rep = &reporter.TerminalReporter{NoColor: *noColorFlag}
	}

	if err := rep.Report(os.Stdout, diagnostics); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(2)
	}

	// Exit code 1 on violations
	for _, d := range diagnostics {
		if d.Severity == config.SeverityError {
			os.Exit(1)
		}
	}
}
