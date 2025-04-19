package d2

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/adrianliechti/wingman-cli/pkg/tool"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2exporter"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"
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

				println("parsed source: " + source)

				output, err := RunOracle(ctx, source)

				println("resulting svg: " + output)

				if err != nil {
					return nil, err
				}

				return output, nil
			},
		},
	}, nil
}

func RunOracle(ctx context.Context, source string) (string, error) {
	graph, config, _ := d2compiler.Compile("", strings.NewReader(source), nil)
	graph.ApplyTheme(d2themescatalog.NeutralDefault.ID)
	ruler, _ := textmeasure.NewRuler()
	_ = graph.SetDimensions(nil, ruler, nil)
	_ = d2dagrelayout.Layout(context.Background(), graph, nil)
	diagram, _ := d2exporter.Export(context.Background(), graph, nil)
	diagram.Config = config
	out, _ := d2svg.Render(diagram, &d2svg.RenderOpts{
		ThemeID: &d2themescatalog.NeutralDefault.ID,
	})
	return string(out), nil
}
