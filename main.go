package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No env file found")
	}

	client := anthropic.NewClient()

	scanner := bufio.NewScanner(os.Stdin)

	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}

		return scanner.Text(), true
	}

	tools := []ToolDefinition{
		ReadFileDefinition,
		ListDirectoryDefinition,
	}
	agent := NewAgent(&client, getUserMessage, tools)

	err = agent.Run(context.TODO())

	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
}

type Agent struct {
	client        *anthropic.Client
	getUserMesage func() (string, bool)
	tools         []ToolDefinition
}

func NewAgent(client *anthropic.Client, getUserMesage func() (string, bool), tools []ToolDefinition) *Agent {
	return &Agent{
		client:        client,
		getUserMesage: getUserMesage,
		tools:         tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	convo := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
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
	var toolDef ToolDefinition
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
