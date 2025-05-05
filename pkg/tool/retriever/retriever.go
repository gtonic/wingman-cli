package retriever

import (
	"context"
	"encoding/json"

	"github.com/adrianliechti/wingman-cli/pkg/tool"
	"github.com/adrianliechti/wingman/pkg/index"
)

type Retriever struct {
	index index.Provider
}

func New(index index.Provider) *Retriever {
	return &Retriever{
		index: index,
	}
}

func (r *Retriever) Tools(ctx context.Context) ([]tool.Tool, error) {
	tools := []tool.Tool{
		{
			Name:        "retrieve_documents",
			Description: "Query the knowledge base to find relevant documents to answer questions",

			Schema: map[string]any{
				"type": "object",

				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The natural language query input. The query input should be clear and standalone",
					},
				},

				"required": []string{"query"},
			},

			Execute: func(ctx context.Context, args map[string]any) (any, error) {
				data, err := json.Marshal(args)

				if err != nil {
					return nil, err
				}

				var parameters struct {
					Query string `json:"query"`
				}

				if err := json.Unmarshal(data, &parameters); err != nil {
					return nil, err
				}

				limit := 5

				documents, err := r.index.Query(ctx, parameters.Query, &index.QueryOptions{
					Limit: &limit,
				})

				if err != nil {
					return nil, err
				}

				var texts []string

				for _, d := range documents {
					texts = append(texts, d.Content)
				}

				return texts, nil
			},
		},
	}

	return tools, nil
}
