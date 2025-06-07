package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type CopyFileInput struct {
	InitialPath string `json:"initial_path" jsonschema_description:"The source file path to copy from. Must be within Jellyfin media directories."`
	EndingPath  string `json:"ending_path" jsonschema_description:"The destination file path to copy to. Must be within Jellyfin media directories."`
}

var CopyFileInputSchema = GenerateSchema[CopyFileInput]()

var CopyFileDefinition = ToolDefinition{
	Name:        "copy_file",
	Description: "Copy a file from one path to another within Jellyfin media directories. Both source and destination paths must be within JELLYFIN_SHOWS_FOLDER or JELLYFIN_MOVIES_FOLDER.",
	InputSchema: CopyFileInputSchema,
	Function:    CopyFile,
}

func CopyFile(input json.RawMessage) (string, error) {
	copyFileInput := CopyFileInput{}
	err := json.Unmarshal(input, &copyFileInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %v", err)
	}

	// Validate and resolve source path
	srcPath, err := validateAndResolvePath(copyFileInput.InitialPath)
	if err != nil {
		return "", fmt.Errorf("invalid source path: %v", err)
	}

	// Validate and resolve destination path
	dstPath, err := validateAndResolvePath(copyFileInput.EndingPath)
	if err != nil {
		return "", fmt.Errorf("invalid destination path: %v", err)
	}

	// Check if source file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", fmt.Errorf("source file does not exist: %s", srcPath)
	}

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %v", err)
	}

	return fmt.Sprintf("Successfully copied file from %s to %s", srcPath, dstPath), nil
}

func validateAndResolvePath(inputPath string) (string, error) {
	showsFolder := os.Getenv("JELLYFIN_SHOWS_FOLDER")
	moviesFolder := os.Getenv("JELLYFIN_MOVIES_FOLDER")

	if showsFolder == "" || moviesFolder == "" {
		return "", fmt.Errorf("JELLYFIN_SHOWS_FOLDER or JELLYFIN_MOVIES_FOLDER environment variable not set")
	}

	// Try to resolve path within shows folder
	showsPath := filepath.Join(showsFolder, inputPath)
	if absShowsPath, err := filepath.Abs(showsPath); err == nil {
		if absBasePath, err := filepath.Abs(showsFolder); err == nil {
			if relPath, err := filepath.Rel(absBasePath, absShowsPath); err == nil && relPath != ".." && !(len(relPath) > 2 && relPath[:3] == "../") {
				return showsPath, nil
			}
		}
	}

	// Try to resolve path within movies folder
	moviesPath := filepath.Join(moviesFolder, inputPath)
	if absMoviesPath, err := filepath.Abs(moviesPath); err == nil {
		if absBasePath, err := filepath.Abs(moviesFolder); err == nil {
			if relPath, err := filepath.Rel(absBasePath, absMoviesPath); err == nil && relPath != ".." && !(len(relPath) > 2 && relPath[:3] == "../") {
				return moviesPath, nil
			}
		}
	}

	return "", fmt.Errorf("path must be within Jellyfin media directories")
}
