package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
	"ojm/tools"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No env file found")
	}

	client := anthropic.NewClient()

	// Collect user input
	inputPath := getInput("Enter the path of the file or folder to organize: ")

	// Get env vars
	moviesFolder := os.Getenv("JELLYFIN_MOVIES_FOLDER")
	showsFolder := os.Getenv("JELLYFIN_SHOWS_FOLDER")

	if moviesFolder == "" || showsFolder == "" {
		log.Fatal("JELLYFIN_MOVIES_FOLDER and JELLYFIN_SHOWS_FOLDER environment variables must be set")
	}

	// Read Jellyfin docs
	jellyfinDocs, err := readJellyfinDocs()
	if err != nil {
		log.Fatalf("Error reading Jellyfin docs: %v", err)
	}

	// Process prompt template
	prompt, err := processPromptTemplate(inputPath, moviesFolder, showsFolder, jellyfinDocs)
	if err != nil {
		log.Fatalf("Error processing prompt template: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	toolDefinitions := tools.AllTools
	agent := NewAgent(&client, getUserMessage, toolDefinitions)

	err = agent.RunWithInitialPrompt(context.TODO(), prompt)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
}

func getInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readJellyfinDocs() (string, error) {
	var docs strings.Builder

	err := filepath.WalkDir("prompt/jellyfin-docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			docs.Write(content)
			docs.WriteString("\n")
		}

		return nil
	})

	return docs.String(), err
}

type PromptData struct {
	InputPath      string
	MoviesFolder   string
	ShowsFolder    string
	JellyfinDocs   string
}

func processPromptTemplate(inputPath, moviesFolder, showsFolder, jellyfinDocs string) (string, error) {
	templateContent, err := os.ReadFile("prompt/main.md")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("prompt").Parse(string(templateContent))
	if err != nil {
		return "", err
	}

	data := PromptData{
		InputPath:    inputPath,
		MoviesFolder: moviesFolder,
		ShowsFolder:  showsFolder,
		JellyfinDocs: jellyfinDocs,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type Agent struct {
	client        *anthropic.Client
	getUserMesage func() (string, bool)
	tools         []tools.ToolDefinition
}

func NewAgent(client *anthropic.Client, getUserMesage func() (string, bool), toolDefs []tools.ToolDefinition) *Agent {
	return &Agent{
		client:        client,
		getUserMesage: getUserMesage,
		tools:         toolDefs,
	}
}

func (a *Agent) RunWithInitialPrompt(ctx context.Context, initialPrompt string) error {
	convo := []anthropic.MessageParam{}

	// Add initial prompt as first message
	userMsg := anthropic.NewUserMessage(anthropic.NewTextBlock(initialPrompt))
	convo = append(convo, userMsg)

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")
	fmt.Println("Sending initial prompt...")

	// Process initial prompt
	message, err := a.runInference(ctx, convo)
	if err != nil {
		return err
	}

	convo = append(convo, message.ToParam())

	toolResults := []anthropic.ContentBlockParamUnion{}
	for _, content := range message.Content {
		switch content.Type {
		case "text":
			fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
		case "tool_use":
			result := a.executeTool(content.ID, content.Name, content.Input)
			toolResults = append(toolResults, result)
		}
	}

	if len(toolResults) > 0 {
		convo = append(convo, anthropic.NewUserMessage(toolResults...))
	}

	// Continue with regular conversation loop
	readUserInput := len(toolResults) == 0
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")

			userInput, ok := a.getUserMesage()
			if !ok {
				break
			}

			userMsg := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			convo = append(convo, userMsg)
		}

		message, err := a.runInference(ctx, convo)
		if err != nil {
			return err
		}

		convo = append(convo, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}

		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}

		readUserInput = false
		convo = append(convo, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}

	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})
	return message, err
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef tools.ToolDefinition
	var found bool

	for _, tool := range a.tools {
		if tool.Name == name {
			found = true
			toolDef = tool
			break
		}
	}

	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)

	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}

	return anthropic.NewToolResultBlock(id, response, false)
}
