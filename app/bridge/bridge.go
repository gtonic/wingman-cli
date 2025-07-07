package bridge

import (
	"context"

	"github.com/adrianliechti/wingman-cli/app"
	"github.com/adrianliechti/wingman-cli/pkg/bridge"

	"github.com/adrianliechti/go-cli"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func Run(ctx context.Context, client *wingman.Client) error {
	tools := app.MustConnectTools(ctx)
	instructions := app.MustParseInstructions()

	//tools = util.OptimizeTools(client, app.DefaultModel, tools)

	cli.Info()
	cli.Info("🖥️ Wingman Bridge")
	cli.Info()

	for _, tool := range tools {
		println("🛠️ " + tool.Name)
	}

	cli.Info()

	return bridge.Run(ctx, client, instructions, tools)
}
