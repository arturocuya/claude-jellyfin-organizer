package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

type ListDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a directory in the working directory. Leave empty for current directory."`
}

var ListDirectoryInputSchema = GenerateSchema[ListDirectoryInput]()

var ListDirectoryDefinition = ToolDefinition{
	Name:        "list_directory",
	Description: "List the contents of a directory. Returns file and directory names in the specified path.",
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
	if dirPath == "" {
		dirPath = "."
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
