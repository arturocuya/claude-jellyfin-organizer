package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadFileInput struct {
	Path  string `json:"path" jsonschema_description:"The file path to read. Can be absolute or relative path. File should not be an image or video."`
	Bytes int    `json:"bytes" jsonschema_description:"Number of bytes to read from the start of the file. If 0, reads the entire file."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a file. Can read entire file or a specified number of bytes from the start. Access is restricted to files within JELLYFIN_SHOWS_FOLDER, JELLYFIN_MOVIES_FOLDER, or SOURCE_FOLDER.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %v", err)
	}

	filePath := readFileInput.Path

	// Validate that the path is within allowed directories
	err = ValidatePath(filePath)
	if err != nil {
		return "", fmt.Errorf("access denied: %v", err)
	}

	// Check if the file is an image or video
	ext := strings.ToLower(filepath.Ext(filePath))
	imageVideoExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tiff", ".tif", ".ico", ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".3gp", ".ogv", ".vob", ".ts", ".mts", ".m2ts"}
	for _, blockedExt := range imageVideoExts {
		if ext == blockedExt {
			return "", fmt.Errorf("cannot read image or video files: %s", filePath)
		}
	}

	// Path validation is already done by ValidateJellyfinPath

	if readFileInput.Bytes == 0 {
		// Read entire file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	} else {
		// Read specified number of bytes
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		buffer := make([]byte, readFileInput.Bytes)
		n, err := file.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return "", err
		}
		return string(buffer[:n]), nil
	}
}
