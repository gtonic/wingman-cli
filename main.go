package main

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/muesli/termenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	ctx := context.Background()

	output := termenv.NewOutput(os.Stdout)

	client := openai.NewClient(
		option.WithBaseURL("http://localhost:8080/v1/"),
		option.WithAPIKey("-"),
	)

	param := openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{}),
		Model:    openai.F(openai.ChatModelGPT4o),
	}

	println()

	for {
		var prompt string

		if err := huh.NewText().
			Value(&prompt).
			Run(); err != nil {
			break
		}

		prompt = strings.TrimSpace(prompt)

		if prompt == "" {
			continue
		}

		param.Messages.Value = append(param.Messages.Value, openai.UserMessage(prompt))

		println("> " + prompt)
		output.HideCursor()
		output.SaveCursorPosition()

		stream := client.Chat.Completions.NewStreaming(ctx, param)
		acc := openai.ChatCompletionAccumulator{}

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			content := acc.Choices[0].Message.Content

			output.RestoreCursorPosition()
			output.ClearLine()

			out, _ := glamour.Render(content, "auto")
			output.WriteString(out)
		}

		if err := stream.Err(); err != nil {
			panic(err)
		}

		output.RestoreCursorPosition()
		output.ClearLine()
		output.ShowCursor()

		content := acc.Choices[0].Message.Content
		out, _ := glamour.Render(content, "auto")

		output.WriteString(out)

		println()

		param.Messages.Value = append(param.Messages.Value, acc.Choices[0].Message)
	}
}
