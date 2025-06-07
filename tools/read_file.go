package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ReadFileInput struct {
	Type  string `json:"type" jsonschema_description:"The type of media directory to read from. Must be 'shows' or 'movies'."`
	Path  string `json:"path" jsonschema_description:"The relative path of a file within the media directory."`
	Bytes int    `json:"bytes" jsonschema_description:"Number of bytes to read from the start of the file. If 0, reads the entire file."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a file within Jellyfin media directories. Can read entire file or a specified number of bytes from the start. Access is restricted to files within JELLYFIN_SHOWS_FOLDER and JELLYFIN_MOVIES_FOLDER.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %v", err)
	}

	var basePath string
	switch readFileInput.Type {
	case "shows":
		basePath = os.Getenv("JELLYFIN_SHOWS_FOLDER")
		if basePath == "" {
			return "", fmt.Errorf("JELLYFIN_SHOWS_FOLDER environment variable not set")
		}
	case "movies":
		basePath = os.Getenv("JELLYFIN_MOVIES_FOLDER")
		if basePath == "" {
			return "", fmt.Errorf("JELLYFIN_MOVIES_FOLDER environment variable not set")
		}
	default:
		return "", fmt.Errorf("invalid type '%s': must be 'shows' or 'movies'", readFileInput.Type)
	}

	filePath := filepath.Join(basePath, readFileInput.Path)

	// Ensure the resolved path is still within the base directory
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %v", err)
	}
	
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %v", err)
	}
	
	relPath, err := filepath.Rel(absBasePath, absFilePath)
	if err != nil || relPath == ".." || len(relPath) > 2 && relPath[:3] == "../" {
		return "", fmt.Errorf("access denied: path outside of allowed directory")
	}

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
