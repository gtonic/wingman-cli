package mcp

import (
	"context"
	"errors"

	"github.com/adrianliechti/wingman-cli/pkg/resource"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (c *Client) Resources(ctx context.Context) ([]resource.Resource, error) {
	var result []resource.Resource

	for name := range c.transports {
		session, err := c.createSession(ctx, name)

		if err != nil {
			return nil, err
		}

		defer session.Close()

		resp, err := session.ListResources(ctx, nil)

		if err != nil {
			return nil, err
		}

		for _, r := range resp.Resources {
			resource := resource.Resource{
				URI: r.URI,

				Name:        r.Name,
				Description: r.Description,

				ContentType: r.MIMEType,

				Content: func(ctx context.Context) ([]byte, error) {
					session, err := c.createSession(ctx, name)

					if err != nil {
						return nil, err
					}

					defer session.Close()

					resp, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
						URI: r.URI,
					})

					if err != nil {
						return nil, err
					}

					if len(resp.Contents) > 1 {
						return nil, errors.New("multiple contents not supported")
					}

					if len(resp.Contents) == 1 {
						content := resp.Contents[0]
						if len(content.Blob) > 0 {
							return content.Blob, nil
						}

						if len(content.Text) > 0 {
							return []byte(content.Text), nil
						}
					}

					return nil, errors.New("no content returned")
				},
			}

			result = append(result, resource)
		}
	}

	return result, nil
}
