package config

import (
	"os"
	"path/filepath"
)

// DetectArchitecture scans directory structure and returns the best matching preset.
func DetectArchitecture(rootDir string) *Config {
	hexScore := 0
	cleanScore := 0

	// Check common hexagonal patterns
	checkDirs := func(dirs []string) int {
		score := 0
		for _, d := range dirs {
			info, err := os.Stat(filepath.Join(rootDir, d))
			if err == nil && info.IsDir() {
				score++
			}
		}
		return score
	}

	hexDirs := []string{
		"internal/core/domain",
		"internal/core/ports",
		"internal/core/services",
		"internal/adapters",
		"internal/domain",
		"internal/ports",
		"internal/infra",
		"domain",
		"ports",
		"adapters",
	}
	hexScore = checkDirs(hexDirs)

	cleanDirs := []string{
		"internal/entity",
		"internal/entities",
		"internal/usecase",
		"internal/usecases",
		"internal/controller",
		"internal/framework",
		"usecase",
		"usecases",
	}
	cleanScore = checkDirs(cleanDirs)

	if cleanScore > hexScore {
		return CleanPreset()
	}

	// Default to Hexagonal preset as it is the most common in modern Go services
	return HexagonalPreset()
}
