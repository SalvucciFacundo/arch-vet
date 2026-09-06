package rules

import (
	"path/filepath"
	"strings"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
)

// SuppressionMap maps filename -> line number -> set of suppressed rule IDs ("*" means all).
type SuppressionMap map[string]map[int]map[string]bool

// BuildSuppressionMap inspects all AST files in the project to extract inline // arch-vet:ignore directives.
func BuildSuppressionMap(project *analyzer.Project) SuppressionMap {
	suppMap := make(SuppressionMap)

	for _, pkg := range project.Packages {
		if pkg.Fset == nil {
			continue
		}

		for _, file := range pkg.Syntax {
			for _, commentGroup := range file.Comments {
				for _, comment := range commentGroup.List {
					text := comment.Text
					if !strings.Contains(text, "arch-vet:ignore") {
						continue
					}

					pos := pkg.Fset.Position(comment.Pos())
					filename := filepath.Clean(pos.Filename)

					if suppMap[filename] == nil {
						suppMap[filename] = make(map[int]map[string]bool)
					}

					rulesToIgnore := parseIgnoredRules(text)

					// Suppress the line of the comment itself and the line immediately following
					for _, line := range []int{pos.Line, pos.Line + 1} {
						if suppMap[filename][line] == nil {
							suppMap[filename][line] = make(map[string]bool)
						}
						for _, r := range rulesToIgnore {
							suppMap[filename][line][r] = true
						}
					}
				}
			}
		}
	}

	return suppMap
}

func parseIgnoredRules(text string) []string {
	idx := strings.Index(text, "arch-vet:ignore")
	if idx == -1 {
		return []string{"*"}
	}

	rest := strings.TrimSpace(text[idx+len("arch-vet:ignore"):])
	// Remove trailing comment characters if block comment
	rest = strings.TrimSuffix(rest, "*/")
	rest = strings.TrimSpace(rest)

	if rest == "" {
		return []string{"*"}
	}

	// Split by commas and spaces
	clean := strings.ReplaceAll(rest, ",", " ")
	parts := strings.Fields(clean)
	if len(parts) == 0 {
		return []string{"*"}
	}

	var rules []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			rules = append(rules, strings.ToUpper(p))
		}
	}

	return rules
}

// IsSuppressed checks if a diagnostic is ignored by an inline directive.
func (sm SuppressionMap) IsSuppressed(d Diagnostic) bool {
	filename := filepath.Clean(d.Position.Filename)
	fileMap, exists := sm[filename]
	if !exists {
		return false
	}

	ruleSet, exists := fileMap[d.Position.Line]
	if !exists {
		return false
	}

	if ruleSet["*"] || ruleSet[strings.ToUpper(d.RuleID)] {
		return true
	}

	return false
}
