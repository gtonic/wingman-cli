package rag

import (
	"context"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

type embeder struct {
	client *wingman.Client
}

func (e *embeder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.client.Embeddings.New(ctx, wingman.EmbeddingsRequest{
		Texts: []string{text},
	})

	if err != nil {
		return nil, err
	}

	return embeddings.Embeddings[0], nil
}
