package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
	"github.com/SalvucciFacundo/arch-vet/internal/git"
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
	diffFlag := flag.Bool("diff", false, "Only report violations in modified or untracked Git files")
	formatFlag := flag.String("format", "text", "Output format (text, json, agent, sarif)")
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

	// 4. If --diff specified, filter diagnostics to only modified files
	if *diffFlag {
		modifiedFiles, err := git.GetModifiedFiles(*dirFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting modified files from git: %v\n", err)
			os.Exit(2)
		}
		if len(modifiedFiles) == 0 {
			if *formatFlag == "agent" || *formatFlag == "ai" {
				fmt.Printf("## Architecture Verification Passed\n\nNo modified or untracked Go files detected in Git working tree.\n")
			} else if *formatFlag == "json" {
				fmt.Println(`{"total_violations": 0, "violations": []}`)
			} else if *formatFlag == "sarif" {
				rep := &reporter.SARIFReporter{ToolVersion: Version}
				_ = rep.Report(os.Stdout, nil)
			} else {
				fmt.Println("✓ Architecture clean: no modified files to check.")
			}
			os.Exit(0)
		}

		modMap := make(map[string]bool)
		for _, f := range modifiedFiles {
			modMap[filepath.ToSlash(f)] = true
		}

		var filtered []rules.Diagnostic
		for _, d := range diagnostics {
			relFile, err := filepath.Rel(*dirFlag, d.Position.Filename)
			if err == nil && modMap[filepath.ToSlash(relFile)] {
				filtered = append(filtered, d)
			} else if modMap[filepath.ToSlash(d.Position.Filename)] {
				filtered = append(filtered, d)
			}
		}
		diagnostics = filtered
	}

	// 5. Report results
	var rep reporter.Reporter
	switch *formatFlag {
	case "json":
		rep = &reporter.JSONReporter{}
	case "agent", "ai":
		rep = &reporter.AgentReporter{}
	case "sarif":
		rep = &reporter.SARIFReporter{ToolVersion: Version}
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
