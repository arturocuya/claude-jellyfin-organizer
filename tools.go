package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gocolly/colly/v2"
	"github.com/invopop/jsonschema"
)

type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

var ListDirectoryDefinition = ToolDefinition{
	Name:        "list_directory",
	Description: "List the contents of a directory. Returns file and directory names in the specified path.",
	InputSchema: ListDirectoryInputSchema,
	Function:    ListDirectory,
}

var SearchIMDbDefinition = ToolDefinition{
	Name:        "search_imdb",
	Description: "Search for a title on IMDb. Returns a JSON string with title, id, and description.",
	InputSchema: SearchIMDbInputSchema,
	Function:    SearchIMDb,
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

type ListDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a directory in the working directory. Leave empty for current directory."`
}

type SearchIMDbInput struct {
	SearchTerm string `json:"search_term" jsonschema_description:"The search term to look for on IMDb."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()
var ListDirectoryInputSchema = GenerateSchema[ListDirectoryInput]()
var SearchIMDbInputSchema = GenerateSchema[SearchIMDbInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
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

func SearchIMDb(input json.RawMessage) (string, error) {
	searchInput := SearchIMDbInput{}
	err := json.Unmarshal(input, &searchInput)
	if err != nil {
		return "", err
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	type IMDbResult struct {
		RawResult string `json:"rawResult"`
		ID        string `json:"id"`
	}

	var results []IMDbResult

	// Scrape each search result
	c.OnHTML(".ipc-metadata-list-summary-item__tc", func(e *colly.HTMLElement) {
		textContent := e.Text
		
		// Get the ID from the first child anchor tag
		href := e.ChildAttr("a", "href")
		
		// Extract ID from href like "/title/tt4955642/?ref_=fn_all_ttl_1"
		var id string
		if href != "" {
			parts := strings.Split(href, "/")
			if len(parts) > 2 {
				// Get the third part and remove query parameters
				idPart := strings.Split(parts[2], "?")[0]
				id = idPart
			}
		}

		results = append(results, IMDbResult{
			RawResult: textContent,
			ID:        id,
		})
	})

	// Construct IMDB search URL
	searchURL := fmt.Sprintf("https://www.imdb.com/find/?q=%s&ref_=nv_sr_sm", url.QueryEscape(searchInput.SearchTerm))
	
	err = c.Visit(searchURL)
	if err != nil {
		return "", fmt.Errorf("failed to scrape IMDB: %w", err)
	}

	// Convert results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(jsonData), nil
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}
