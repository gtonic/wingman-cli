package resource

import (
	"context"
)

type ContentFn func(ctx context.Context) ([]byte, error)

type Resource struct {
	URI string

	Name        string
	Description string

	Content     ContentFn
	ContentType string
}
