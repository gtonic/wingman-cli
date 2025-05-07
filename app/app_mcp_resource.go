package app

import (
	"context"
	"os"

	"github.com/adrianliechti/wingman-cli/pkg/mcp"
	"github.com/adrianliechti/wingman-cli/pkg/resource"
)

func MustConnectResources(ctx context.Context) []resource.Resource {
	resources, err := ConnectResources(ctx)

	if err != nil {
		panic(err)
	}

	return resources
}

func ConnectResources(ctx context.Context) ([]resource.Resource, error) {
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

		return mcp.Resources(ctx)
	}

	return nil, nil
}
