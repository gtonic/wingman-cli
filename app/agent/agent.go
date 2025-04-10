package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/muesli/termenv"

	"github.com/adrianliechti/wingman-cli/pkg/cli"
	"github.com/adrianliechti/wingman-cli/pkg/markdown"
	"github.com/adrianliechti/wingman-cli/pkg/tool"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func Run(ctx context.Context, client *wingman.Client, model, system string, tools []tool.Tool) error {
	output := termenv.NewOutput(os.Stdout)

	input := wingman.CompletionRequest{
		Model: model,

		CompleteOptions: wingman.CompleteOptions{
			Tools: toTools(tools),
		},
	}

	if system != "" {
		input.Messages = append(input.Messages, wingman.SystemMessage(system))
	}

	output.WriteString("\n")

	for {
		prompt, _ := cli.Prompt("> ", "")

		if prompt == "" {
			continue
		}

		input.Messages = append(input.Messages, wingman.UserMessage(prompt))

		var message *wingman.Message

		for {
			completion, err := client.Completions.New(ctx, input)

			if err != nil {
				return err
			}

			message = completion.Message
			input.Messages = append(input.Messages, *message)

			calls := message.ToolCalls()

			if len(calls) == 0 {
				break
			}

			for _, call := range calls {
				println("⚡️ " + call.Name)

				content, err := handleToolCall(ctx, tools, call)

				if err != nil {
					content = err.Error()
				}

				input.Messages = append(input.Messages, wingman.ToolMessage(call.ID, content))
			}
		}

		if message == nil {
			return nil
		}

		markdown.Render(output, message.Text())
	}
}

func handleToolCall(ctx context.Context, tools []tool.Tool, call wingman.ToolCall) (string, error) {
	var handler tool.ExecuteFn

	for _, t := range tools {
		if !strings.EqualFold(t.Name, call.Name) {
			continue
		}

		handler = t.Execute
	}

	if handler == nil {
		return "", errors.New("Unknown tool: " + call.Name)
	}

	var args map[string]any
	json.Unmarshal([]byte(call.Arguments), &args)

	result, err := handler(ctx, args)

	if err != nil {
		return "", err
	}

	var content string

	switch v := result.(type) {
	case string:
		content = v
	default:
		data, _ := json.Marshal(v)
		content = string(data)
	}

	return content, nil
}

func toTools(tools []tool.Tool) []wingman.Tool {
	var result []wingman.Tool

	for _, t := range tools {
		result = append(result, toTool(t))
	}

	return result
}

func toTool(t tool.Tool) wingman.Tool {
	return wingman.Tool{
		Name:        t.Name,
		Description: t.Description,

		Parameters: t.Schema,
	}
}
