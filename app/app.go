package app

import (
	"context"
	"os"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var (
	DefaultModel     string
	DefaultModelMini string

	ThinkingModel     string
	ThinkingModelMini string

	EmbeddingModel     string
	EmbeddingModelMini string
)

func MustClient(ctx context.Context) *wingman.Client {
	url := os.Getenv("WINGMAN_URL")

	if url == "" {
		url = "http://localhost:8080"
	}

	var options []wingman.RequestOption

	if token := os.Getenv("WINGMAN_TOKEN"); token != "" {
		options = append(options, wingman.WithToken(token))
	}

	return wingman.New(url, options...)
}
