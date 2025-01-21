package completer

import (
	"context"
	"os"
	"strings"

	"github.com/adrianliechti/wingman/pkg/markdown"

	"github.com/muesli/termenv"
	"github.com/openai/openai-go"
)

func Run(ctx context.Context, client *openai.Client, model, prompt string) error {
	prompt = strings.TrimSpace(prompt)

	if prompt == "" {
		return nil
	}

	output := termenv.NewOutput(os.Stdout)

	param := openai.ChatCompletionNewParams{
		Model: openai.F(model),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
	}

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

	output.RestoreCursorPosition()
	output.ClearLine()
	output.ShowCursor()

	content := acc.Choices[0].Message.Content
	markdown.Render(output, content)

	return nil
}
