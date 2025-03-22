package chat

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/openai/openai-go"
)

func Run(ctx context.Context, client openai.Client, model string) error {
	//output := termenv.NewOutput(os.Stdout)

	param := openai.ChatCompletionNewParams{
		Model:    model,
		Messages: []openai.ChatCompletionMessageParamUnion{},
	}

	println()
	//output.WriteString("\n")

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

		// output.WriteString("> " + prompt)
		// output.WriteString("\n")
		println("> " + prompt)
		println()

		param.Messages = append(param.Messages, openai.UserMessage(prompt))

		//output.HideCursor()
		//output.SaveCursorPosition()

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
		println()

		param.Messages = append(param.Messages, acc.Choices[0].Message.ToParam())

		// output.RestoreCursorPosition()
		// output.ClearLine()
		// output.ShowCursor()

		// content := acc.Choices[0].Message.Content
		// markdown.Render(output, content)
	}
}
