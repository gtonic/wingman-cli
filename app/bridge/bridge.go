package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adrianliechti/wingman-cli/app"
	"github.com/adrianliechti/wingman-cli/pkg/util"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/cors"

	"github.com/adrianliechti/go-cli"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func Run(ctx context.Context, client *wingman.Client) error {
	tools := app.MustConnectTools(ctx)

	tools = util.OptimizeTools(client, app.DefaultModel, tools)

	cli.Info()
	cli.Info("🖥️ MCP Server")
	cli.Info()

	for _, tool := range tools {
		println("🛠️ " + tool.Name)
	}

	cli.Info()

	s := server.NewMCPServer(
		"Wingman MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	for _, t := range tools {
		schema, _ := json.Marshal(t.Schema)

		tool := mcp.Tool{
			Name:           t.Name,
			Description:    t.Description,
			RawInputSchema: schema,
		}

		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, err := convertArgs(request.Params.Arguments)

			if err != nil {
				return nil, err
			}

			result, err := t.Execute(ctx, args)

			if err != nil {
				return nil, err
			}

			var content string

			switch v := result.(type) {
			case string:
				content = v
			default:
				data, _ := json.Marshal(v)
				content = string(data)
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent(content),
				},
			}, nil
		})
	}

	addr := "localhost:4200"

	server := server.NewSSEServer(s,
		server.WithBaseURL(fmt.Sprintf("http://%s", addr)),
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /.well-known/wingman", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"name": "wingman",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	mux.Handle("/sse", server)
	mux.Handle("/message", server)

	if err := http.ListenAndServe(addr, cors.AllowAll().Handler(mux)); err != nil {
		return err
	}

	return nil
}

func convertArgs(val any) (map[string]any, error) {
	data, err := json.Marshal(val)

	if err != nil {
		return nil, err
	}

	var args map[string]any

	if err := json.Unmarshal(data, &args); err == nil {
		return args, nil
	}

	return map[string]any{
		"input": val,
	}, nil
}
