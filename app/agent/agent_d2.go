package agent

import (
	"context"
	_ "embed"
	"os"
	"strings"

	"github.com/adrianliechti/wingman-cli/pkg/tool"
	"github.com/adrianliechti/wingman-cli/pkg/tool/d2"
	"github.com/adrianliechti/wingman-cli/pkg/util"

	"github.com/adrianliechti/go-cli"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var (
	//go:embed prompt_d2.txt
	prompt_d2 string
)

func ParseTemplate() (string, error) {
	for _, name := range []string{"container.d2"} {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(name)

		if err != nil {
			continue
		}

		template := strings.TrimSpace(string(data))
		return template, nil
	}

	return "", nil
}

func RunD2(ctx context.Context, client *wingman.Client, model string) error {
	cli.Info("🤗 Hello, I'm your D2 Drawing Assistant")
	cli.Info()

	tpl, err := ParseTemplate()

	if err != nil {
		return err
	}

	println("🛠️ " + tpl)

	var tools []tool.Tool

	cmd, err := d2.New("d2")
	if err == nil {
		if t, err := cmd.Tools(ctx); err == nil {
			tools = append(tools, t...)
		}
	}

	tools = util.OptimizeTools(client, model, tools)

	cli.Info()

	return Run(ctx, client, model, tools, &RunOptions{
		Prompt:     prompt_d2,
		PromptFile: true,
	})
}
