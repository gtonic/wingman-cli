package mcp

import (
	"context"
	"errors"

	"github.com/adrianliechti/wingman-cli/pkg/resource"

	"github.com/mark3labs/mcp-go/mcp"
)

func (c *Client) Resources(ctx context.Context) ([]resource.Resource, error) {
	var result []resource.Resource

	for _, c := range c.clients {
		resp, err := c.ListResources(ctx, mcp.ListResourcesRequest{})

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
					req := mcp.ReadResourceRequest{}
					req.Params.URI = r.URI

					result, err := c.ReadResource(ctx, req)

					if err != nil {
						return nil, err
					}

					if len(result.Contents) > 1 {
						return nil, errors.New("multiple contents not supported")
					}

					for _, content := range result.Contents {
						switch content := content.(type) {
						case mcp.TextResourceContents:
							return []byte(content.Text), nil

						case mcp.BlobResourceContents:
							return []byte(content.Blob), nil
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
