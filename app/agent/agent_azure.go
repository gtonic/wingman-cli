package agent

import (
	"context"
	_ "embed"

	"github.com/adrianliechti/wingman-cli/app"
	"github.com/adrianliechti/wingman-cli/pkg/tool"
	"github.com/adrianliechti/wingman-cli/pkg/tool/azure"
	"github.com/adrianliechti/wingman-cli/pkg/util"

	"github.com/adrianliechti/go-cli"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var (
	//go:embed prompt_azure.txt
	prompt_azure string
)

func RunAzure(ctx context.Context, client *wingman.Client) error {
	azure, err := azure.New()

	if err != nil {
		return err
	}

	prompt := app.MustParsePrompt()

	if prompt == "" {
		prompt = prompt_azure
	}

	var tools []tool.Tool

	if t, err := azure.Tools(ctx); err == nil {
		tools = append(tools, t...)
	}

	tools = util.OptimizeTools(client, app.DefaultModel, tools)

	cli.Info()
	cli.Info("🤗 Hello, I'm your Azure AI Assistant")
	cli.Info()

	return Run(ctx, client, app.ThinkingModel, prompt, tools)
}
