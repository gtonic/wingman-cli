package util

import (
	"context"
	"os"

	"github.com/adrianliechti/wingman-cli/pkg/mcp"
	"github.com/adrianliechti/wingman-cli/pkg/tool"
)

func ParseMCP() ([]tool.Tool, error) {
	ctx := context.Background()

	for _, name := range []string{".mcp.json", ".mcp.yaml", "mcp.json", "mcp.yaml"} {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			continue
		}

		cfg, err := mcp.Parse(name)

		if err != nil {
			return nil, err
		}

		mcp, err := mcp.New(cfg)

		if err != nil {
			return nil, err
		}

		return mcp.Tools(ctx)
	}

	return nil, nil
}
