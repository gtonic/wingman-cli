package mcp

import (
	"context"
	"errors"

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
