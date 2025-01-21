package chat

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/adrianliechti/wingman/pkg/markdown"

	"github.com/openai/openai-go"

	"github.com/charmbracelet/huh"
	"github.com/muesli/termenv"
)

func Run(ctx context.Context, client *openai.Client, model string) error {
	output := termenv.NewOutput(os.Stdout)

	param := openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{}),
		Model:    openai.F(model),
	}

	output.WriteString("\n")

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

		param.Messages.Value = append(param.Messages.Value, openai.UserMessage(prompt))

		output.HideCursor()
		output.SaveCursorPosition()

		acc := openai.ChatCompletionAccumulator{}
		stream := client.Chat.Completions.NewStreaming(ctx, param)

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			output.RestoreCursorPosition()
			output.ClearLine()

			content := acc.Choices[0].Message.Content
			markdown.Render(output, content)
		}

		if err := stream.Err(); err != nil {
			return err
		}

		param.Messages.Value = append(param.Messages.Value, acc.Choices[0].Message)

		output.RestoreCursorPosition()
		output.ClearLine()
		output.ShowCursor()

		content := acc.Choices[0].Message.Content
		markdown.Render(output, content)
	}
}
