package agent

import (
	"context"
	_ "embed"

	"github.com/adrianliechti/wingman-cli/pkg/tool"
	"github.com/adrianliechti/wingman-cli/pkg/tool/cli"
	"github.com/adrianliechti/wingman-cli/pkg/tool/fs"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var (
	//go:embed system_admin.txt
	system_admin string
)

func RunAdmin(ctx context.Context, client *wingman.Client, model string) error {
	println("🤗 Hello, I'm your AI Administrator")
	println("")

	var tools []tool.Tool

	fs, err := fs.New("")

	if err != nil {
		return err
	}

	if t, err := fs.Tools(ctx); err == nil {
		tools = append(tools, t...)
	}

	for _, name := range []string{"docker", "kubectl", "helm"} {
		if c, err := cli.New(name); err == nil {
			println("🔨 " + name)

			t, _ := c.Tools(ctx)

			//w := util.OptimizeContext(&completer{client, model}, t)
			tools = append(tools, t...)
		}
	}

	println("")

	return Run(ctx, client, model, tools, &RunOptions{
		System: system_admin,

		OptimizeTools: true,
	})
}
