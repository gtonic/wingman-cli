package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adrianliechti/wingman-cli/pkg/mcp/config"
	"github.com/adrianliechti/wingman-cli/pkg/tool"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Client struct {
	clients map[string]client.MCPClient
}

func New(config *config.Config) (*Client, error) {
	c := &Client{
		clients: make(map[string]client.MCPClient),
	}

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

	ctx := context.Background()

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "wingman",
		Version: "1.0.0",
	}

	for _, c := range c.clients {
		result, err := c.Initialize(ctx, initRequest)

		if err != nil {
			return nil, err
		}

		fmt.Printf(
			"Initialized with server: %s %s\n\n",
			result.ServerInfo.Name,
			result.ServerInfo.Version,
		)
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

			tool := tool.Tool{
				Name:        t.Name,
				Description: t.Description,

				Schema: schema,

				Execute: func(ctx context.Context, args map[string]any) (any, error) {
					req := mcp.CallToolRequest{
						Request: mcp.Request{
							Method: "tools/call",
						},
					}

					req.Params.Name = t.Name
					req.Params.Arguments = args

					result, err := c.CallTool(ctx, req)

					if err != nil {
						return nil, err
					}

					if len(result.Content) == 0 {
						return nil, errors.New("no content returned")
					}

					if len(result.Content) > 1 {
						return nil, errors.New("multiple content types not supported")
					}

					for _, content := range result.Content {
						switch content := content.(type) {
						case mcp.TextContent:
							return content.Text, nil
						case mcp.ImageContent:
							return nil, errors.New("image content not supported")
						case mcp.EmbeddedResource:
							return nil, errors.New("embedded resource not supported")
						}
					}

					return nil, nil
				},
			}

			result = append(result, tool)
		}

	}

	return result, nil
}
