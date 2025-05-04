package remote

import (
	"context"

	"github.com/adrianliechti/wingman-cli/pkg/index"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var _ index.Index = (*Index)(nil)

type Index struct {
	index  string
	client *wingman.Client
}

func (i *Index) List(ctx context.Context, options *index.ListOptions) (*index.Page[index.Record], error) {
	document, err := i.client.Documents.List(ctx, i.index)

	if err != nil {
		return nil, err
	}

	var items []index.Record

	for _, d := range document {
		items = append(items, index.Record{
			ID: d.ID,

			Text:   d.Content,
			Vector: d.Embedding,

			Metadata: d.Metadata,
		})
	}

	return &index.Page[index.Record]{
		Items: items,
	}, nil
}

func (i *Index) Index(ctx context.Context, record ...index.Record) error {
	var documents []wingman.Document

	for _, r := range record {
		documents = append(documents, wingman.Document{
			ID: r.ID,

			Content:   r.Text,
			Embedding: r.Vector,

			Metadata: r.Metadata,
		})
	}

	_, err := i.client.Documents.New(ctx, i.index, documents)
	return err
}

func (i *Index) Delete(ctx context.Context, ids ...string) error {
	return i.client.Documents.Delete(ctx, i.index, ids)
}

func (i *Index) Search(ctx context.Context, vector []float32, topK int) ([]index.Record, error) {
	panic("unimplemented")
}
