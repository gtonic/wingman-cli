package util

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/adrianliechti/wingman-cli/pkg/tool"
)

type Model interface {
	Complete(ctx context.Context, input string) (string, error)
}

func OptimizeContext(model Model, tools []tool.Tool) []tool.Tool {
	var wrapped []tool.Tool

	for _, t := range tools {
		wrapped = append(wrapped, createWrapper(model, t))
	}

	return wrapped
}

func createWrapper(m Model, t tool.Tool) tool.Tool {
	schema := tool.Schema{
		"type": "object",

		"properties": map[string]any{
			"goal": map[string]any{
				"type":        "string",
				"description": "The goal of the task including the expected record, fields and information you expect or search in the result. This goal is used to compress and filter large results.",
			},

			"input": t.Schema,
		},
	}

	return tool.Tool{
		Name:        t.Name,
		Description: t.Description,

		Schema: schema,

		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			goal, ok := args["goal"].(string)

			if !ok {
				return nil, errors.New("goal is required")
			}

			input, ok := args["input"].(map[string]any)

			if !ok {
				return nil, errors.New("input is required")
			}

			println("🥅", "goal", goal)

			result, err := t.Execute(ctx, input)

			if err != nil {
				return nil, err
			}

			var data string

			switch val := result.(type) {
			case string:
				data = val
			case []any, map[string]any:
				json, _ := json.Marshal(val)
				data = string(json)
			}

			println("#######")
			println("data", data)
			println("#######")

			summary, err := m.Complete(ctx, "Extract the relevant information from the following data based on the goal:\n"+goal+"\n\nData: "+data)

			if err != nil {
				return nil, err
			}

			println("#######")
			println("summary", summary)
			println("#######")

			return summary, nil
		},
	}
}
