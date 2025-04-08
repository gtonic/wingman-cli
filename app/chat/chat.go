package chat

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/huh"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func Run(ctx context.Context, client *wingman.Client, model string) error {
	println()

	input := wingman.CompletionRequest{
		Model: model,

		CompleteOptions: wingman.CompleteOptions{
			Stream: func(ctx context.Context, completion wingman.Completion) error {
				print(completion.Message.Text())
				return nil
			},
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

		println("> " + prompt)
		println()

		input.Messages = append(input.Messages, wingman.UserMessage(prompt))

		completion, err := client.Completions.New(ctx, input)

		if err != nil {
			return err
		}

		input.Messages = append(input.Messages, *completion.Message)

		println()
		println()
	}
}
