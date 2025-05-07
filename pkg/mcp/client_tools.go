package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/adrianliechti/wingman-cli/pkg/tool"

	"github.com/mark3labs/mcp-go/mcp"
)

func (c *Client) Tools(ctx context.Context) ([]tool.Tool, error) {
	var result []tool.Tool

	for _, c := range c.clients {
		resp, err := c.ListTools(ctx, mcp.ListToolsRequest{})

		if err != nil {
			return nil, err
		}

		for _, t := range resp.Tools {
			var schema tool.Schema

			input, _ := json.Marshal(t.InputSchema)

			if err := json.Unmarshal([]byte(input), &schema); err != nil {
				return nil, err
			}

			if len(t.InputSchema.Properties) == 0 {
				schema = map[string]any{
					"type":                 "object",
					"properties":           map[string]any{},
					"additionalProperties": false,
				}
			}

			tool := tool.Tool{
				Name:        t.Name,
				Description: t.Description,

				Schema: schema,

				Execute: func(ctx context.Context, args map[string]any) (any, error) {
					if args == nil {
						args = map[string]any{}
					}

					req := mcp.CallToolRequest{}
					req.Params.Name = t.Name
					req.Params.Arguments = args

					result, err := c.CallTool(ctx, req)

					if err != nil {
						return nil, err
					}

					if len(result.Content) > 1 {
						return nil, errors.New("multiple content types not supported")
					}

					for _, content := range result.Content {
						switch content := content.(type) {
						case mcp.TextContent:
							text := strings.TrimSpace(content.Text)
							return text, nil

						case mcp.ImageContent:
							return nil, errors.New("image content not supported")

						case mcp.EmbeddedResource:
							return nil, errors.New("embedded resource not supported")
						}
					}

					return nil, errors.New("no content returned")
				},
			}

			result = append(result, tool)
		}
	}

	return result, nil
}
