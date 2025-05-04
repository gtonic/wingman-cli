package index

import (
	"context"

	_ "github.com/ncruces/go-sqlite3/embed"
)

type Index interface {
	List(ctx context.Context, options *ListOptions) (*Page[Record], error)
	Index(ctx context.Context, record ...Record) error
	Search(ctx context.Context, vector []float32, topK int) ([]Record, error)
	Delete(ctx context.Context, ids ...string) error
}

type Page[T any] struct {
	Items []T

	Cursor string
}

type Record struct {
	ID string

	Text   string
	Vector []float32

	Metadata map[string]string
}

type ListOptions struct {
	Limit  *int
	Cursor string
}
