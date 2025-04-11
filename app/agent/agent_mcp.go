package agent

import (
	"context"

	"github.com/adrianliechti/wingman-cli/pkg/mcp"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func RunMCP(ctx context.Context, client *wingman.Client, model string) error {
	println("🤗 Hello, I'm your AI Assistant")
	println("")

	cfg, err := mcp.Parse("mcp.json")

	if err != nil {
		return err
	}

	for name, _ := range cfg.Servers {
		println("🛠️ " + name)
	}

	mcp, err := mcp.New(cfg)

	if err != nil {
		return err
	}

	tools, err := mcp.Tools(ctx)

	if err != nil {
		return err
	}

	tools = toolsWrapper(client, model, tools)

	return Run(ctx, client, model, tools, &RunOptions{})
}
