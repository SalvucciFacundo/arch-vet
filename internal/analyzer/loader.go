package analyzer

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// LoadProject analyzes the Go project rooted at rootDir.
func LoadProject(rootDir string) (*Project, error) {
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root dir: %w", err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedModule,
		Dir: absRootDir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("packages.Load failed: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go packages found in %s", absRootDir)
	}

	project := &Project{
		RootDir:  absRootDir,
		Packages: make(map[string]*PackageInfo),
	}

	for _, pkg := range pkgs {
		if pkg.Module != nil && project.ModulePath == "" {
			project.ModulePath = pkg.Module.Path
		}

		pkgInfo := &PackageInfo{
			ID:        pkg.PkgPath,
			Name:      pkg.Name,
			Fset:      pkg.Fset,
			Syntax:    pkg.Syntax,
			Types:     pkg.Types,
			TypesInfo: pkg.TypesInfo,
		}

		// Calculate relative path from module root
		if len(pkg.GoFiles) > 0 {
			dir := filepath.Dir(pkg.GoFiles[0])
			rel, err := filepath.Rel(absRootDir, dir)
			if err == nil && !strings.HasPrefix(rel, "..") {
				pkgInfo.RelPath = filepath.ToSlash(rel)
			} else if project.ModulePath != "" && strings.HasPrefix(pkg.PkgPath, project.ModulePath) {
				rel = strings.TrimPrefix(pkg.PkgPath, project.ModulePath)
				pkgInfo.RelPath = filepath.ToSlash(strings.TrimPrefix(rel, "/"))
			}
		}

		// Extract imports with positions from AST
		for _, file := range pkg.Syntax {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				pos := pkg.Fset.Position(imp.Pos())
				pkgInfo.Imports = append(pkgInfo.Imports, ImportInfo{
					Path:     path,
					Position: pos,
					File:     pos.Filename,
				})
			}
		}

		project.Packages[pkg.PkgPath] = pkgInfo
	}

	return project, nil
}
