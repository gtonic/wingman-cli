package coder

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrianliechti/wingman/pkg/fs"
	"github.com/adrianliechti/wingman/pkg/markdown"

	"github.com/charmbracelet/huh"
	"github.com/muesli/termenv"
	"github.com/openai/openai-go"
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

	fs := fs.New(path)

	output := termenv.NewOutput(os.Stdout)

	output.WriteString("🤗 I'm your coding assistant and can help you with your application.\n")
	output.WriteString("🗂️  " + path + "\n")
	output.WriteString("\n")

	params := openai.ChatCompletionNewParams{
		Model: openai.F(model),

		Tools: openai.F([]openai.ChatCompletionToolParam{
			toolListFiles,

			toolReadFile,
			toolCreateFile,
			toolDeleteFile,

			toolCreateDir,
			toolDeleteDir,
		}),

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
				_, m := handleToolCall(ctx, fs, call)
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
