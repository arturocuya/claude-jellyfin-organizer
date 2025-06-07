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

	sourcePath := renameInput.SourcePath
	targetPath := renameInput.TargetPath

	// Validate both source and target paths are within Jellyfin directories
	err = ValidatePath(renameInput.SourcePath)
	if err != nil {
		return "", fmt.Errorf("invalid source path: %v", err)
	}

	err = ValidatePath(renameInput.TargetPath)
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
