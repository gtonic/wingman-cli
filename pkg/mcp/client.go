package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/adrianliechti/wingman-cli/pkg/tool"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Client struct {
	clients map[string]*client.Client
}

func New(config *Config) (*Client, error) {
	c := &Client{
		clients: make(map[string]*client.Client),
	}

	ctx := context.Background()

	for n, s := range config.Servers {
		switch s.Type {
		case "stdio":
			env := []string{}

			for k, v := range s.Env {
				env = append(env, k+"="+v)
			}

			client, err := client.NewStdioMCPClient(s.Command, env, s.Args...)

			if err != nil {
				return nil, err
			}

			c.clients[n] = client

		case "sse":
			client, err := client.NewSSEMCPClient(s.URL, client.WithHeaders(s.Headers))

			if err != nil {
				return nil, err
			}

			c.clients[n] = client
		default:
			return nil, errors.New("invalid server type")
		}
	}

	for _, c := range c.clients {
		if err := c.Start(ctx); err != nil {
			return nil, err
		}
	}

	for _, c := range c.clients {
		req := mcp.InitializeRequest{}
		req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		req.Params.ClientInfo = mcp.Implementation{
			Name:    "wingman",
			Version: "1.0.0",
		}

		if _, err := c.Initialize(ctx, req); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) Close() {
	for _, c := range c.clients {
		c.Close()
	}
}

func (c *Client) Tools(ctx context.Context) ([]tool.Tool, error) {
	var result []tool.Tool

	for _, c := range c.clients {
		tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})

		if err != nil {
			return nil, err
		}

		for _, t := range tools.Tools {
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
