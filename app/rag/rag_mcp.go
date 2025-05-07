package rag

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/adrianliechti/go-cli"
	"github.com/adrianliechti/wingman-cli/pkg/resource"
	wingman "github.com/adrianliechti/wingman/pkg/client"
	"github.com/adrianliechti/wingman/pkg/index"
)

func IndexResources(ctx context.Context, client *wingman.Client, i index.Provider, resources []resource.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	supported := []string{
		"text/plain",
		"text/markdown",
	}

	var cursor string

	mapping := make(map[string]string)

	for {
		page, err := i.List(ctx, &index.ListOptions{
			Cursor: cursor,
		})

		if err != nil {
			return err
		}

		cursor = page.Cursor

		if len(page.Items) == 0 {
			break
		}

		for _, i := range page.Items {
			uri := i.Metadata["uri"]

			if uri == "" {
				continue
			}

			mapping[uri] = i.ID
		}
	}

	var result error

	for _, r := range resources {
		if !slices.Contains(supported, r.ContentType) {
			continue
		}

		if _, found := mapping[r.URI]; found {
			continue
		}

		data, err := r.Content(ctx)

		if err != nil {
			result = errors.Join(result, err)
			continue
		}

		md5_hash := md5.Sum(data)
		md5_text := hex.EncodeToString(md5_hash[:])

		cli.Infof("Indexing /%s...", r.URI)

		// extraction, err := client.Extractions.New(ctx, wingman.ExtractionRequest{
		// 	Name:   name,
		// 	Reader: bytes.NewReader(data),
		// })

		// if err != nil {
		// 	result = errors.Join(result, err)
		// 	return nil
		// }

		segments, err := client.Segments.New(ctx, wingman.SegmentRequest{
			Name:   "content.txt",
			Reader: strings.NewReader(string(data)),

			SegmentLength:  wingman.Ptr(3000),
			SegmentOverlap: wingman.Ptr(1500),
		})

		if err != nil {
			result = errors.Join(result, err)
			continue
		}

		var documents []index.Document

		for i, segment := range segments {
			document := index.Document{
				Content: segment.Text,

				Metadata: map[string]string{
					"uri": r.URI,

					"index":    fmt.Sprintf("%d", i),
					"revision": md5_text,
				},
			}

			documents = append(documents, document)
		}

		if err := i.Index(ctx, documents...); err != nil {
			result = errors.Join(result, err)
			continue
		}
	}

	return result
}
