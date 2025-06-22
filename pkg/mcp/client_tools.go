package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/adrianliechti/wingman-cli/pkg/tool"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (c *Client) Tools(ctx context.Context) ([]tool.Tool, error) {
	var result []tool.Tool

	for name := range c.transports {
		session, err := c.createSession(ctx, name)

		if err != nil {
			return nil, err
		}

		defer session.Close()

		resp, err := session.ListTools(ctx, nil)

		if err != nil {
			return nil, err
		}

		for _, t := range resp.Tools {
			var schema tool.Schema

			input, _ := t.InputSchema.MarshalJSON()

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

					session, err := c.createSession(ctx, name)

					if err != nil {
						return nil, err
					}

					defer session.Close()

					resp, err := session.CallTool(ctx, &mcp.CallToolParams{
						Name:      t.Name,
						Arguments: args,
					})

					if err != nil {
						return nil, err
					}

					if len(resp.Content) > 1 {
						return nil, errors.New("multiple content types not supported")
					}

					for _, content := range resp.Content {
						switch content.Type {
						case "text":
							text := strings.TrimSpace(content.Text)
							return text, nil
						case "image":
							return nil, errors.New("image content not supported")
						case "audio":
							return nil, errors.New("audio content not supported")
						case "resource":
							return nil, errors.New("embedded resource not supported")
						default:
							return nil, errors.New("unknown content type")
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
