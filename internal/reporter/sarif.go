package reporter

import (
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/SalvucciFacundo/arch-vet/internal/rules"
)

// SARIFReporter formats diagnostics according to the OASIS SARIF 2.1.0 standard for GitHub Code Scanning.
type SARIFReporter struct {
	ToolVersion string
}

type sarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID                   string                   `json:"id"`
	Name                 string                   `json:"name"`
	ShortDescription     sarifMessage             `json:"shortDescription"`
	Help                 sarifMessage             `json:"help"`
	DefaultConfiguration sarifRuleDefaultConfig   `json:"defaultConfiguration"`
}

type sarifRuleDefaultConfig struct {
	Level string `json:"level"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
}

func (r *SARIFReporter) Report(w io.Writer, diagnostics []rules.Diagnostic) error {
	version := r.ToolVersion
	if version == "" {
		version = "0.1.0-dev"
	}

	ruleMap := make(map[string]sarifRule)
	var results []sarifResult

	for _, d := range diagnostics {
		level := "error"
		if d.Severity == "warning" {
			level = "warning"
		} else if d.Severity == "info" {
			level = "note"
		}

		if _, exists := ruleMap[d.RuleID]; !exists {
			ruleMap[d.RuleID] = sarifRule{
				ID:               d.RuleID,
				Name:             d.RuleName,
				ShortDescription: sarifMessage{Text: d.RuleName},
				Help:             sarifMessage{Text: d.Explanation + "\n\nRemediation: " + d.Remediation},
				DefaultConfiguration: sarifRuleDefaultConfig{
					Level: level,
				},
			}
		}

		uri := filepath.ToSlash(d.Position.Filename)
		line := d.Position.Line
		if line < 1 {
			line = 1
		}
		col := d.Position.Column
		if col < 1 {
			col = 1
		}

		results = append(results, sarifResult{
			RuleID:  d.RuleID,
			Level:   level,
			Message: sarifMessage{Text: d.Message},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI:       uri,
							URIBaseID: "%SRCROOT%",
						},
						Region: sarifRegion{
							StartLine:   line,
							StartColumn: col,
						},
					},
				},
			},
		})
	}

	rulesList := []sarifRule{}
	for _, rule := range ruleMap {
		rulesList = append(rulesList, rule)
	}

	if results == nil {
		results = []sarifResult{}
	}

	report := sarifReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "arch-vet",
						Version:        version,
						InformationURI: "https://github.com/SalvucciFacundo/arch-vet",
						Rules:          rulesList,
					},
				},
				Results: results,
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
