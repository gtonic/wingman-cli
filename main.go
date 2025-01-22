package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/adrianliechti/wingman/pkg/chat"
	"github.com/adrianliechti/wingman/pkg/cli"
	"github.com/adrianliechti/wingman/pkg/coder"
	"github.com/adrianliechti/wingman/pkg/completer"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var version string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	defer stop()

	app := initApp()

	if err := app.Run(ctx, os.Args); err != nil {
		cli.Fatal(err)
	}
}

func initApp() cli.Command {
	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		apiKey = "-"
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")

	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/"

		if apiKey == "-" {
			baseURL = "http://localhost:8080/v1/"
		}
	}

	defaultModel := os.Getenv("OPENAI_MODEL")

	if defaultModel == "" {
		defaultModel = openai.ChatModelGPT4o
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return cli.Command{
		Usage: "Wingman AI CLI",

		Suggest: true,
		Version: version,

		HideHelpCommand: true,

		Action: func(ctx context.Context, cmd *cli.Command) error {
			prompt := strings.TrimSpace(strings.Join(cmd.Args().Slice(), " "))

			if input := readInput(); input != "" {
				if prompt == "" {
					prompt += "Analyze the following input\n"
					prompt += "Explain your findings\n"
					prompt += "Give reommendations based on your observations\n"
					prompt += "If you see problems or errors, propose solutions\n"
					prompt += "\n"
					prompt += "Input:\n"
					prompt += input
				}

				return completer.Run(ctx, client, defaultModel, prompt)
			}

			if cmd.Args().Len() > 0 {
				return completer.Run(ctx, client, defaultModel, prompt)
			}

			cli.ShowAppHelp(cmd)
			return nil
		},

		Commands: []*cli.Command{
			{
				Name:  "chat",
				Usage: "AI Chat",

				Action: func(ctx context.Context, cmd *cli.Command) error {
					return chat.Run(ctx, client, defaultModel)
				},
			},

			{
				Name:  "coder",
				Usage: "AI Coder",

				Action: func(ctx context.Context, cmd *cli.Command) error {
					return coder.Run(ctx, client, defaultModel, "")
				},
			},
		},
	}
}

func readInput() string {
	fi, err := os.Stdin.Stat()

	if err != nil {
		return ""
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		return ""
	}

	data, err := io.ReadAll(os.Stdin)

	if err != nil {
		return ""
	}

	return string(data)
}
