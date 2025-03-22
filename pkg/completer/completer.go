package completer

import (
	"context"
	"strings"

	"github.com/openai/openai-go"
)

func Run(ctx context.Context, client openai.Client, model, prompt string) error {
	prompt = strings.TrimSpace(prompt)

	if prompt == "" {
		return nil
	}

	//output := termenv.NewOutput(os.Stdout)

	param := openai.ChatCompletionNewParams{
		Model: model,

		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	// output.HideCursor()
	// output.SaveCursorPosition()

	acc := openai.ChatCompletionAccumulator{}
	stream := client.Chat.Completions.NewStreaming(ctx, param)

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// output.RestoreCursorPosition()
		// output.ClearLine()

		// content := acc.Choices[0].Message.Content
		// markdown.Render(output, content)

		content := chunk.Choices[0].Delta.Content
		print(content)
	}

	if err := stream.Err(); err != nil {
		return err
	}

	println()

	// output.RestoreCursorPosition()
	// output.ClearLine()
	// output.ShowCursor()

	// content := acc.Choices[0].Message.Content
	// markdown.Render(output, content)

	return nil
}
