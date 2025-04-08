package coder

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/muesli/termenv"

	"github.com/adrianliechti/wingman-cli/pkg/markdown"
	"github.com/adrianliechti/wingman-cli/pkg/tool"
	"github.com/adrianliechti/wingman-cli/pkg/tool/cli"
	"github.com/adrianliechti/wingman-cli/pkg/tool/fs"

	wingman "github.com/adrianliechti/wingman/pkg/client"
	"github.com/adrianliechti/wingman/pkg/provider"
)

var (
	//go:embed prompt.txt
	system string
)

func Run(ctx context.Context, client *wingman.Client, model, path string) error {
	path, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	fs, err := fs.New(path)

	if err != nil {
		return err
	}

	var tools []tool.Tool

	if t, err := fs.Tools(ctx); err == nil {
		tools = append(tools, t...)
	}

	for _, name := range []string{"git", "wget", "curl", "docker", "kubectl", "helm", "jq", "yq"} {
		if c, err := cli.New(name); err == nil {
			t, _ := c.Tools(ctx)
			tools = append(tools, t...)
		}
	}

	output := termenv.NewOutput(os.Stdout)

	output.WriteString("🤗 I'm your coding assistant and can help you with your application.\n")
	output.WriteString("🗂️  " + path + "\n")
	output.WriteString("\n")

	input := wingman.CompletionRequest{
		Model: model,

		Messages: []wingman.Message{
			wingman.SystemMessage(system),
		},

		CompleteOptions: wingman.CompleteOptions{
			Tools: toTools(tools),
		},
	}

	for {
		var prompt string

		if err := huh.NewText().
			Lines(2).
			Value(&prompt).
			Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}

			return err
		}

		prompt = strings.TrimSpace(prompt)

		if prompt == "" {
			continue
		}

		output.WriteString("> " + prompt)
		output.WriteString("\n")

		input.Messages = append(input.Messages, wingman.UserMessage(prompt))

		completion, err := client.Completions.New(ctx, input)

		if err != nil {
			return err
		}

		message := completion.Message

		input.Messages = append(input.Messages, *message)

		var calls []provider.ToolCall

		for _, c := range message.Content {
			if c.ToolCall != nil {
				calls = append(calls, *c.ToolCall)
			}
		}

		if len(calls) > 0 {
			for _, c := range calls {
				result := handleToolCall(ctx, tools, c)

				var content string

				switch v := result.(type) {
				case string:
					content = v
				case error:
					content = v.Error()
				default:
					data, _ := json.Marshal(v)
					content = string(data)
				}

				input.Messages = append(input.Messages, wingman.Message{
					Role: provider.MessageRoleUser,

					Content: []provider.Content{
						{
							ToolResult: &provider.ToolResult{
								ID:   c.ID,
								Data: content,
							},
						},
					},
				})
			}

			completion, err = client.Completions.New(ctx, input)

			if err != nil {
				return err
			}

			message = completion.Message
			input.Messages = append(input.Messages, *message)
		}

		markdown.Render(output, message.Text())
	}
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

func handleToolCall(ctx context.Context, tools []tool.Tool, call provider.ToolCall) any {
	println("⚡️ " + call.Name)

	var handler tool.ExecuteFn

	for _, t := range tools {
		if t.Name != call.Name {
			continue
		}

		handler = t.Execute
	}

	if handler == nil {
		return errors.New("Unknown tool: " + call.Name)
	}

	var args map[string]any
	json.Unmarshal([]byte(call.Arguments), &args)

	result, err := handler(ctx, args)

	if err != nil {
		return err
	}

	var content string

	switch v := result.(type) {
	case string:
		content = v
	default:
		data, _ := json.Marshal(v)
		content = string(data)
	}

	return content
}
