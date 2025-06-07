package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ListDirectoryInput struct {
	Type    string `json:"type" jsonschema_description:"The type of media directory to list. Must be 'shows' or 'movies'."`
	Subpath string `json:"subpath" jsonschema_description:"The relative path within the media directory. Leave empty for the root of the media directory."`
}

var ListDirectoryInputSchema = GenerateSchema[ListDirectoryInput]()

var ListDirectoryDefinition = ToolDefinition{
	Name:        "list_directory",
	Description: "List the contents of a Jellyfin media directory. Access is restricted to paths within JELLYFIN_SHOWS_FOLDER and JELLYFIN_MOVIES_FOLDER.",
	InputSchema: ListDirectoryInputSchema,
	Function:    ListDirectory,
}

func ListDirectory(input json.RawMessage) (string, error) {
	listDirInput := ListDirectoryInput{}
	err := json.Unmarshal(input, &listDirInput)
	if err != nil {
		return "", err
	}

	var basePath string
	switch listDirInput.Type {
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
		return "", fmt.Errorf("invalid type '%s': must be 'shows' or 'movies'", listDirInput.Type)
	}

	dirPath := basePath
	if listDirInput.Subpath != "" {
		dirPath = filepath.Join(basePath, listDirInput.Subpath)
	}

	// Ensure the resolved path is still within the base directory
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %v", err)
	}
	
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve directory path: %v", err)
	}
	
	relPath, err := filepath.Rel(absBasePath, absDirPath)
	if err != nil || relPath == ".." || len(relPath) > 2 && relPath[:3] == "../" {
		return "", fmt.Errorf("access denied: path outside of allowed directory")
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	result := ""
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			result += name + "/\n"
		} else {
			info, err := entry.Info()
			if err != nil {
				result += name + "\n"
			} else {
				result += fmt.Sprintf("%s (%d bytes)\n", name, info.Size())
			}
		}
	}

	return result, nil
}
