package bridge

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/cors"

	"github.com/adrianliechti/wingman-cli/pkg/tool"
	wingman "github.com/adrianliechti/wingman/pkg/client"
)

func Run(ctx context.Context, client *wingman.Client, tools []tool.Tool) error {
	s := mcp.NewServer("Wingman MCP Server", "1.0.0", nil)

	for _, t := range tools {
		if _, ok := t.Schema["additionalProperties"]; !ok {
			t.Schema["additionalProperties"] = false
		}

		if _, ok := t.Schema["required"]; !ok {
			required := []string{}

			for k := range t.Schema["properties"].(map[string]any) {
				required = append(required, k)
			}

			t.Schema["required"] = required
		}

		data, _ := json.Marshal(t.Schema)

		schema := new(jsonschema.Schema)

		if err := schema.UnmarshalJSON(data); err != nil {
			return err
		}

		handler := func(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[any]) (*mcp.CallToolResultFor[any], error) {
			args, err := convertArgs(params.Arguments)

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

			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: content,
					},
				},
			}, nil
		}

		s.AddTools(mcp.NewServerTool(t.Name, t.Description, handler, mcp.Input(mcp.Schema(schema))))
	}

	addr := "localhost:4200"

	mux := http.NewServeMux()

	mux.HandleFunc("GET /.well-known/wingman", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"name": "wingman",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	h := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return s
	})

	mux.Handle("/sse", h)
	// mux.Handle("/message", h)

	server := &http.Server{
		Addr:    addr,
		Handler: cors.AllowAll().Handler(mux),
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
