package remote

import (
	"context"

	"github.com/adrianliechti/wingman/pkg/index"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var _ index.Provider = (*Index)(nil)

type Index struct {
	index  string
	client *wingman.Client
}

func (i *Index) List(ctx context.Context, options *index.ListOptions) (*index.Page[index.Document], error) {
	document, err := i.client.Documents.List(ctx, i.index)

	if err != nil {
		return nil, err
	}

	var items []index.Document

	for _, d := range document {
		items = append(items, index.Document{
			ID: d.ID,

			Content:   d.Content,
			Embedding: d.Embedding,

			Metadata: d.Metadata,
		})
	}

	return &index.Page[index.Document]{
		Items: items,
	}, nil
}

func (i *Index) Index(ctx context.Context, documents ...index.Document) error {
	var input []wingman.Document

	for _, d := range documents {
		input = append(input, wingman.Document{
			ID: d.ID,

			Title:   d.Title,
			Source:  d.Source,
			Content: d.Content,

			Metadata: d.Metadata,

			Embedding: d.Embedding,
		})
	}

	_, err := i.client.Documents.Index(ctx, i.index, input, nil)
	return err
}

func (i *Index) Delete(ctx context.Context, ids ...string) error {
	return i.client.Documents.Delete(ctx, i.index, ids)
}

func (i *Index) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	resp, err := i.client.Documents.Query(ctx, i.index, wingman.DocumentQueryRequest{
		Text: query,

		Limit: options.Limit,
	})

	if err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range resp {
		result := index.Result{
			Document: index.Document{
				ID: r.ID,

				Title:   r.Title,
				Source:  r.Source,
				Content: r.Content,

				Metadata: r.Metadata,
			},
		}

		if r.Score != nil {
			result.Score = float32(*r.Score)
		}

		results = append(results, result)
	}

	return results, nil
}
