package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type CopyFileInput struct {
	InitialPath string `json:"initial_path" jsonschema_description:"The source file path to copy from. Use an absolute path"`
	EndingPath  string `json:"ending_path" jsonschema_description:"The destination file path to copy to. Use an absolute path"`
}

var CopyFileInputSchema = GenerateSchema[CopyFileInput]()

var CopyFileDefinition = ToolDefinition{
	Name:        "copy_file",
	Description: "Copy a file from any source path to a destination within Jellyfin media directories. Source and destination should be absolute paths",
	InputSchema: CopyFileInputSchema,
	Function:    CopyFile,
}

func CopyFile(input json.RawMessage) (string, error) {
	copyFileInput := CopyFileInput{}
	err := json.Unmarshal(input, &copyFileInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %v", err)
	}

	srcPath := copyFileInput.InitialPath

	// Validate destination path within Jellyfin directories
	if err = ValidatePath(copyFileInput.EndingPath); err != nil {
		return "", err
	}

	dstPath := copyFileInput.EndingPath

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
