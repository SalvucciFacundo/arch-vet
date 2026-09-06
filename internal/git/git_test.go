package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetModifiedFiles(t *testing.T) {
	// Test against current git repo root
	files, err := GetModifiedFiles(".")
	if err != nil {
		t.Fatalf("GetModifiedFiles failed: %v", err)
	}

	// Create a temporary dummy Go file to test detection
	dummyFile := filepath.Join(".", "test_dummy_change.go")
	if err := os.WriteFile(dummyFile, []byte("package git\n"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}
	defer os.Remove(dummyFile)

	filesAfter, err := GetModifiedFiles(".")
	if err != nil {
		t.Fatalf("GetModifiedFiles after dummy write failed: %v", err)
	}

	found := false
	for _, f := range filesAfter {
		if f == "test_dummy_change.go" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected test_dummy_change.go to be detected in modified files, got: %v (initial: %v)", filesAfter, files)
	}
}
