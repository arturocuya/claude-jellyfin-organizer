package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath validates that a path is within allowed Jellyfin directories
func ValidatePath(inputPath string) error {
	showsFolder := os.Getenv("JELLYFIN_SHOWS_FOLDER")
	moviesFolder := os.Getenv("JELLYFIN_MOVIES_FOLDER")
	sourceFolder := os.Getenv("SOURCE_FOLDER")

	// Check for path traversal attempts
	if strings.Contains(inputPath, "..") {
		return fmt.Errorf("path contains invalid directory traversal: %s", inputPath)
	}

	// Get absolute path
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path is within permitted folders
	permittedFolders := []string{showsFolder, moviesFolder, sourceFolder}
	for _, folder := range permittedFolders {
		if folder != "" {
			absFolderPath, err := filepath.Abs(folder)
			if err != nil {
				continue
			}
			if strings.HasPrefix(absPath, absFolderPath) {
				return nil
			}
		}
	}

	return fmt.Errorf("path is not within permitted folders: %s", inputPath)
}
