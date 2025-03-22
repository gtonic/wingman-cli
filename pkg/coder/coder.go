package coder

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

func Run(ctx context.Context, client *openai.Client, model, path string) error {
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

	params := openai.ChatCompletionNewParams{
		Model: openai.F(model),

		Tools: openai.F(toTools(tools)),

		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
		}),
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

		params.Messages.Value = append(params.Messages.Value, openai.UserMessage(prompt))

		completion, err := client.Chat.Completions.New(ctx, params)

		if err != nil {
			return err
		}

		message := completion.Choices[0].Message
		params.Messages.Value = append(params.Messages.Value, message)

		for len(message.ToolCalls) > 0 {
			for _, call := range message.ToolCalls {
				m := handleToolCall(ctx, tools, call)
				params.Messages.Value = append(params.Messages.Value, m)
			}

			completion, err = client.Chat.Completions.New(ctx, params)

			if err != nil {
				return err
			}

			message = completion.Choices[0].Message
			params.Messages.Value = append(params.Messages.Value, message)
		}

		content := message.Content
		markdown.Render(output, content)
	}
}

func handleToolCall(ctx context.Context, tools []tool.Tool, call openai.ChatCompletionMessageToolCall) openai.ChatCompletionToolMessageParam {
	println("⚡️ " + call.Function.Name)

	var handler tool.ExecuteFn

	for _, t := range tools {
		if t.Name != call.Function.Name {
			continue
		}

		handler = t.Execute
	}

	if handler == nil {
		return openai.ToolMessage(call.ID, "Unknown tool: "+call.Function.Name)
	}

	var args map[string]any
	json.Unmarshal([]byte(call.Function.Arguments), &args)

	result, err := handler(ctx, args)

	if err != nil {
		return openai.ToolMessage(call.ID, err.Error())
	}

	var content string

	switch v := result.(type) {
	case string:
		content = v
	default:
		data, _ := json.Marshal(v)
		content = string(data)
	}

	return openai.ToolMessage(call.ID, content)
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
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(shared.FunctionDefinitionParam{
			Name:        openai.F(t.Name),
			Description: openai.F(t.Description),

			Parameters: openai.F(shared.FunctionParameters(t.Schema)),
		}),
	}
}
