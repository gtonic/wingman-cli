package tool

import (
	"context"
)

type Provider interface {
	Tools(ctx context.Context) ([]Tool, error)
}

type Schema map[string]any
type ExecuteFn func(ctx context.Context, args map[string]any) (any, error)

type Tool struct {
	Name        string
	Description string

	Schema  Schema
	Execute ExecuteFn
}
