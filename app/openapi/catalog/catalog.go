package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"slices"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"

	wingman "github.com/adrianliechti/wingman/pkg/client"

	"github.com/adrianliechti/wingman-cli/app/openapi/client"
)

type Catalog struct {
	doc *openapi3.T

	api *client.Client
	llm *wingman.Client

	tools    []wingman.Tool
	messages []wingman.Message

	operations map[string]Operation
}

type Client interface {
	Execute(ctx context.Context, method, path string, body io.Reader) ([]byte, string)
}

func New(path string, api *client.Client, llm *wingman.Client) (*Catalog, error) {
	doc, err := parse(path)

	if err != nil {
		return nil, err
	}

	messages := []wingman.Message{
		wingman.SystemMessage("You are connected to an API Server defined by your Tools. You can interact with it by sending messages. Keep answers short and to the point."),
	}

	operations, err := getOperations(doc)

	if err != nil {
		return nil, err
	}

	var tools []wingman.Tool

	for _, o := range operations {
		tools = append(tools, wingman.Tool{
			Name:        o.Name,
			Description: o.Description,

			Parameters: o.Schema,

			Strict: wingman.Ptr(true),
		})
	}

	return &Catalog{
		doc: doc,

		api: api,
		llm: llm,

		tools:    tools,
		messages: messages,

		operations: operations,
	}, nil
}

func (c *Catalog) Query(ctx context.Context, model, prompt string) (string, error) {
	c.messages = append(c.messages, wingman.UserMessage(prompt))

	completion, err := c.llm.Completions.New(ctx, wingman.CompletionRequest{
		Model: model,

		Messages: c.messages,

		CompleteOptions: wingman.CompleteOptions{
			Tools: c.tools,
		},
	})

	if err != nil {
		return "", err
	}

	message := completion.Message
	c.messages = append(c.messages, *message)

	for {
		var calls []wingman.ToolCall

		for _, c := range message.Content {
			if c.ToolCall != nil {
				calls = append(calls, *c.ToolCall)
			}
		}

		if len(calls) > 0 {
			for _, call := range calls {
				data, err := c.handleToolCall(ctx, call.Name, call.Arguments)

				if err != nil {
					return "", err
				}

				c.messages = append(c.messages, wingman.ToolMessage(call.ID, data))
			}

			completion, err = c.llm.Completions.New(ctx, wingman.CompletionRequest{
				Model: model,

				Messages: c.messages,

				CompleteOptions: wingman.CompleteOptions{
					Tools: c.tools,
				},
			})

			if err != nil {
				return "", err
			}

			message = completion.Message
			c.messages = append(c.messages, *message)
		}

		for _, c := range message.Content {
			if c.ToolCall != nil {
				calls = append(calls, *c.ToolCall)
			}
		}

		if len(calls) == 0 {
			break
		}
	}

	return message.Text(), nil
}

func (c *Catalog) handleToolCall(ctx context.Context, name string, arguments string) (string, error) {
	o, found := c.operations[name]

	if !found {
		return "", errors.New("function tool not found")
	}

	parameters := map[string]any{}
	json.Unmarshal([]byte(arguments), &parameters)

	path := o.Path

	query := url.Values{}

	for k, v := range parameters {
		key := "{" + k + "}"
		value, ok := v.(string)

		if !ok {
			continue
		}

		path = strings.ReplaceAll(path, key, value)

		if slices.Contains(o.Queries, strings.ToLower(k)) {
			query.Set(k, value)
		}
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var body io.Reader

	if val, ok := parameters["body"]; ok {
		data, _ := json.Marshal(val)
		body = strings.NewReader(string(data))
	}

	data, err := c.api.Execute(ctx, o.Method, path, o.Type, body)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getOperations(doc *openapi3.T) (map[string]Operation, error) {
	result := map[string]Operation{}

	for p, path := range doc.Paths.Map() {
		operations := path.Operations()

		for m, o := range operations {
			if o.OperationID == "" {
				continue
			}

			name := camelToSnake(o.OperationID)

			description := o.Summary
			description = strings.TrimRight(description, ". \n")
			description += "."

			if o.Description != "" {
				description += " " + o.Description
				description = strings.TrimRight(description, ". \n")
				description += "."
			}

			queries := []string{}

			required := []string{}
			properties := map[string]any{}

			for _, p := range o.Parameters {
				if p.Value == nil {
					continue
				}

				if strings.EqualFold(p.Value.In, "query") {
					query := strings.ToLower(p.Value.Name)
					queries = append(queries, query)
				} else if strings.EqualFold(p.Value.In, "path") {
					//
				} else if strings.EqualFold(p.Value.In, "header") {
					continue
				} else {
					// unknown
					continue
				}

				name := p.Value.Name
				//types := p.Value.Schema.Value.Type.Slice()

				property := map[string]any{
					"type":        "string",
					"description": p.Value.Description,
				}

				// if len(types) == 0 {
				// 	definition["type"] = types[0]
				// }

				// if len(types) > 1 {
				// 	definition["type"] = types
				// }

				properties[name] = property

				if p.Value.Required {
					required = append(required, name)
				}
			}

			var contentType string

			if o.RequestBody != nil {
				content := o.RequestBody.Value.Content.Get("application/json")

				if content != nil {
					contentType = "application/json"

					properties["body"] = content.Schema.Value.Properties
					required = append(required, "body")
				}
			}

			if len(properties) == 0 {
				properties["body"] = map[string]any{}
			}

			schema := map[string]any{
				"type": "object",

				"properties": properties,
			}

			if len(required) > 0 {
				schema["required"] = required
			}

			result[name] = Operation{
				Name:        name,
				Description: description,

				Method: m,
				Path:   p,

				Queries: queries,

				Type:   contentType,
				Schema: schema,
			}
		}
	}

	return result, nil
}

func camelToSnake(s string) string {
	var result strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}

			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
