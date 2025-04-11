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
	//go:embed system_coder.txt
	system_coder string
)

func RunCoder(ctx context.Context, client *wingman.Client, model string) error {
	println("🤗 Hello, I'm your AI Coder")
	println()

	fs, err := fs.New("")

	if err != nil {
		return err
	}

	var tools []tool.Tool

	if t, err := fs.Tools(ctx); err == nil {
		tools = append(tools, t...)
	}

	for _, name := range []string{"git", "wget", "curl", "docker", "kubectl", "helm", "jq", "yq"} {
		if c, err := cli.New(name); err == nil {
			println("🔨 " + name)

			t, _ := c.Tools(ctx)
			t = toolsWrapper(client, model, t)

			tools = append(tools, t...)
		}
	}

	println()

	return Run(ctx, client, model, tools, &RunOptions{
		System: system_coder,
	})
}
