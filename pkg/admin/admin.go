package admin

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrianliechti/wingman/pkg/markdown"
	"github.com/adrianliechti/wingman/pkg/tool"
	"github.com/adrianliechti/wingman/pkg/tool/cli"
	"github.com/adrianliechti/wingman/pkg/tool/fs"

	"github.com/charmbracelet/huh"
	"github.com/muesli/termenv"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

var (
	//go:embed prompt.txt
	system string
)

func Run(ctx context.Context, client openai.Client, model, path string) error {
	path, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	fs, err := fs.New(path)

	if err != nil {
		return err
	}

	var tools []tool.Tool

	for _, name := range []string{"kubectl", "helm", "docker"} {
		if c, err := cli.New(name); err == nil {
			t, _ := c.Tools(ctx)
			tools = append(tools, t...)
		}
	}

	if t, err := fs.Tools(ctx); err == nil {
		tools = append(tools, t...)
	}

	output := termenv.NewOutput(os.Stdout)

	output.WriteString("🤗 I'm your system admin assistant and can help you with your platform.\n")
	output.WriteString("🗂️  " + path + "\n")
	output.WriteString("\n")

	params := openai.ChatCompletionNewParams{
		Model: model,

		Tools: toTools(tools),

		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
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

		params.Messages = append(params.Messages, openai.UserMessage(prompt))

		completion, err := client.Chat.Completions.New(ctx, params)

		if err != nil {
			return err
		}

		message := completion.Choices[0].Message
		params.Messages = append(params.Messages, message.ToParam())

		for len(message.ToolCalls) > 0 {
			for _, c := range message.ToolCalls {
				params.Messages = append(params.Messages, handleToolCall(ctx, tools, c))
			}

			completion, err = client.Chat.Completions.New(ctx, params)

			if err != nil {
				return err
			}

			message = completion.Choices[0].Message
			params.Messages = append(params.Messages, message.ToParam())
		}

		content := message.Content
		markdown.Render(output, content)
	}
}

func handleToolCall(ctx context.Context, tools []tool.Tool, call openai.ChatCompletionMessageToolCall) openai.ChatCompletionMessageParamUnion {
	println("⚡️ " + call.Function.Name)

	var handler tool.ExecuteFn

	for _, t := range tools {
		if t.Name != call.Function.Name {
			continue
		}

		handler = t.Execute
	}

	if handler == nil {
		return openai.ToolMessage("Unknown tool: "+call.Function.Name, call.ID)
	}

	var args map[string]any
	json.Unmarshal([]byte(call.Function.Arguments), &args)

	result, err := handler(ctx, args)

	if err != nil {
		return openai.ToolMessage(err.Error(), call.ID)
	}

	var content string

	switch v := result.(type) {
	case string:
		content = v
	default:
		data, _ := json.Marshal(v)
		content = string(data)
	}

	return openai.ToolMessage(content, call.ID)
}

func toTools(tools []tool.Tool) []openai.ChatCompletionToolParam {
	var result []openai.ChatCompletionToolParam

	for _, t := range tools {
		result = append(result, toTool(t))
	}

	return result
}

func toTool(t tool.Tool) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Function: shared.FunctionDefinitionParam{
			Name:        t.Name,
			Description: openai.String(t.Description),

			Parameters: shared.FunctionParameters(t.Schema),
		},
	}
}
