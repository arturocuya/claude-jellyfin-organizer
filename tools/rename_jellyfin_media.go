package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type RenameJellyfinMediaInput struct {
	SourcePath string `json:"source_path" jsonschema_description:"The source file or folder path to move/rename. Must be within Jellyfin media directories."`
	TargetPath string `json:"target_path" jsonschema_description:"The target file or folder path. Must be within Jellyfin media directories."`
}

var RenameJellyfinMediaInputSchema = GenerateSchema[RenameJellyfinMediaInput]()

var RenameJellyfinMediaDefinition = ToolDefinition{
	Name:        "rename_jellyfin_media",
	Description: "Move or rename files and folders within Jellyfin media directories. Both source and target paths must be within JELLYFIN_SHOWS_FOLDER or JELLYFIN_MOVIES_FOLDER. Works like 'mv' command but restricted to Jellyfin media folders.",
	InputSchema: RenameJellyfinMediaInputSchema,
	Function:    RenameJellyfinMedia,
}

func RenameJellyfinMedia(input json.RawMessage) (string, error) {
	renameInput := RenameJellyfinMediaInput{}
	err := json.Unmarshal(input, &renameInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %v", err)
	}

	// Validate both source and target paths are within Jellyfin directories
	sourcePath, err := validateJellyfinPath(renameInput.SourcePath)
	if err != nil {
		return "", fmt.Errorf("invalid source path: %v", err)
	}

	targetPath, err := validateJellyfinPath(renameInput.TargetPath)
	if err != nil {
		return "", fmt.Errorf("invalid target path: %v", err)
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Create target directory if it doesn't exist (for the parent directory)
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %v", err)
	}

	// Check if target already exists
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("target path already exists: %s", targetPath)
	}

	// Perform the move/rename operation
	err = os.Rename(sourcePath, targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to move/rename: %v", err)
	}

	return fmt.Sprintf("Successfully moved/renamed %s to %s", sourcePath, targetPath), nil
}

func validateJellyfinPath(inputPath string) (string, error) {
	showsFolder := os.Getenv("JELLYFIN_SHOWS_FOLDER")
	moviesFolder := os.Getenv("JELLYFIN_MOVIES_FOLDER")

	if showsFolder == "" || moviesFolder == "" {
		return "", fmt.Errorf("JELLYFIN_SHOWS_FOLDER or JELLYFIN_MOVIES_FOLDER environment variable not set")
	}

	// Check if it's an absolute path within one of the base folders
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %v", err)
	}

	absShowsFolder, err := filepath.Abs(showsFolder)
	if err != nil {
		return "", fmt.Errorf("failed to resolve shows folder absolute path: %v", err)
	}

	absMoviesFolder, err := filepath.Abs(moviesFolder)
	if err != nil {
		return "", fmt.Errorf("failed to resolve movies folder absolute path: %v", err)
	}

	// Check if path is within shows folder
	if relPath, err := filepath.Rel(absShowsFolder, absPath); err == nil && relPath != ".." && !(len(relPath) > 2 && relPath[:3] == "../") {
		return absPath, nil
	}

	// Check if path is within movies folder
	if relPath, err := filepath.Rel(absMoviesFolder, absPath); err == nil && relPath != ".." && !(len(relPath) > 2 && relPath[:3] == "../") {
		return absPath, nil
	}

	// If not absolute path within base folders, try as relative path within shows folder
	showsPath := filepath.Join(showsFolder, inputPath)
	if absShowsPath, err := filepath.Abs(showsPath); err == nil {
		if relPath, err := filepath.Rel(absShowsFolder, absShowsPath); err == nil && relPath != ".." && !(len(relPath) > 2 && relPath[:3] == "../") {
			return absShowsPath, nil
		}
	}

	// Try as relative path within movies folder
	moviesPath := filepath.Join(moviesFolder, inputPath)
	if absMoviesPath, err := filepath.Abs(moviesPath); err == nil {
		if relPath, err := filepath.Rel(absMoviesFolder, absMoviesPath); err == nil && relPath != ".." && !(len(relPath) > 2 && relPath[:3] == "../") {
			return absMoviesPath, nil
		}
	}

	return "", fmt.Errorf("path must be within Jellyfin media directories (%s or %s)", showsFolder, moviesFolder)
}
