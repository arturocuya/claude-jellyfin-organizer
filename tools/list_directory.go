package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

type ListDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"The directory path to list. Can be absolute or relative path."`
}

var ListDirectoryInputSchema = GenerateSchema[ListDirectoryInput]()

var ListDirectoryDefinition = ToolDefinition{
	Name:        "list_directory",
	Description: "List the contents of a directory.",
	InputSchema: ListDirectoryInputSchema,
	Function:    ListDirectory,
}

func ListDirectory(input json.RawMessage) (string, error) {
	listDirInput := ListDirectoryInput{}
	err := json.Unmarshal(input, &listDirInput)
	if err != nil {
		return "", err
	}

	dirPath := listDirInput.Path

	// Validate that the path is within allowed directories
	err = ValidatePath(dirPath)
	if err != nil {
		return "", fmt.Errorf("access denied: %v", err)
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
