package d2

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/adrianliechti/wingman-cli/pkg/tool"

	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
)

func New(name string) (*Command, error) {
	_, err := exec.LookPath(name)

	if err != nil {
		return nil, err
	}

	c := &Command{
		name: name,
	}

	return c, nil
}

var (
	_ tool.Provider = (*Command)(nil)
)

type Command struct {
	name string
}

func (c *Command) Tools(ctx context.Context) ([]tool.Tool, error) {
	return []tool.Tool{
		{
			Name:        "d2",
			Description: "run the d2 api to generate diagrams",

			Schema: tool.Schema{
				"type": "object",

				"properties": map[string]any{
					"args": map[string]any{
						"type": "array",

						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},

			Execute: func(ctx context.Context, args map[string]any) (any, error) {
				data, err := json.Marshal(args)

				if err != nil {
					return nil, err
				}

				var parameters struct {
					Args []string `json:"args"`
				}

				if err := json.Unmarshal(data, &parameters); err != nil {
					return nil, err
				}

				if len(parameters.Args) == 0 {
					return nil, nil
				}

				source := parameters.Args[0]

				output, err := RunOracle(ctx, source)

				if err != nil {
					return nil, err
				}

				return output, nil
			},
		},
	}, nil
}

func RunOracle(ctx context.Context, source string) (string, error) {
	// Compile the source into a diagram
	diagram, graph, err := d2lib.Compile(ctx, source, nil, nil)
	if err != nil {
		return "", err
	}

	// Layout the graph using elk layout with default options
	err = d2elklayout.Layout(ctx, graph, nil)
	if err != nil {
		return "", err
	}

	// Render the diagram to SVG with default options
	svgBytes, err := d2svg.Render(diagram, nil)
	if err != nil {
		return "", err
	}

	return string(svgBytes), nil
}
