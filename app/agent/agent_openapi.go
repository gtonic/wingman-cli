package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/adrianliechti/wingman-cli/pkg/cli"
	"github.com/adrianliechti/wingman-cli/pkg/openapi"
	openapiclient "github.com/adrianliechti/wingman-cli/pkg/openapi/client"

	wingman "github.com/adrianliechti/wingman/pkg/client"
)

var (
	//go:embed system_openapi.txt
	system_openapi string
)

func RunOpenAPI(ctx context.Context, client *wingman.Client, model string, path, url, bearer, username, password string) error {
	println("🤗 Hello, I'm your OpenAPI AI Assistant")
	println()

	c, err := openapiclient.New(url,
		openapiclient.WithBearer(bearer),
		openapiclient.WithBasicAuth(username, password),
		openapiclient.WithConfirm(handleConfirm),
	)

	if err != nil {
		return err
	}

	catalog, err := openapi.New(path, c)

	if err != nil {
		return err
	}

	tools := catalog.Tools()
	tools = toolsWrapper(client, model, tools)

	return Run(ctx, client, model, tools, &RunOptions{
		System: system_openapi,
	})
}

func handleConfirm(method, path, contentType string, body io.Reader) error {
	fmt.Printf("⚡️ %s %s", method, path)
	fmt.Println()

	if body != nil && contentType == "application/json" {
		var val map[string]any

		json.NewDecoder(body).Decode(&val)
		data, _ := json.MarshalIndent(val, "", "  ")

		fmt.Println(string(data))
	}

	if strings.EqualFold(method, "HEAD") || strings.EqualFold(method, "GET") {
		return nil
	}

	ok, err := cli.Confirm("Are you sure?", true)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New("operation cancelled by user")
	}

	return nil
}
