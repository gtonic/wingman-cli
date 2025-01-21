package coder

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/adrianliechti/wingman/pkg/fs"

	"github.com/openai/openai-go"
)

var (
	toolListFiles = openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("list_files"),
			Description: openai.String("list all files and directories recursively in the file system"),

			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",

				"properties": map[string]any{},
			}),
		}),
	}

	toolReadFile = openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("read_file"),
			Description: openai.String("read the (text) content of a file at path"),

			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",

				"properties": map[string]any{
					"path": map[string]string{
						"type": "string",
					},
				},

				"required": []string{"path"},
			}),
		}),
	}

	toolCreateFile = openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("create_file"),
			Description: openai.String("create or overwrite file at path with content (text)"),

			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",

				"properties": map[string]any{
					"path": map[string]string{
						"type": "string",
					},

					"content": map[string]string{
						"type": "string",
					},
				},

				"required": []string{"path", "content"},
			}),
		}),
	}

	toolDeleteFile = openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("delete_file"),
			Description: openai.String("delete a file at path and all empty parent directories"),

			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",

				"properties": map[string]any{
					"path": map[string]string{
						"type": "string",
					},
				},

				"required": []string{"path"},
			}),
		}),
	}

	toolCreateDir = openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("create_dir"),
			Description: openai.String("create a directroy at path and all missing parent directories"),

			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",

				"properties": map[string]any{
					"path": map[string]string{
						"type": "string",
					},
				},

				"required": []string{"path"},
			}),
		}),
	}

	toolDeleteDir = openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),

		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("delete_dir"),
			Description: openai.String("delete a dir at path and all child files and directories"),

			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",

				"properties": map[string]any{
					"path": map[string]string{
						"type": "string",
					},
				},

				"required": []string{"path"},
			}),
		}),
	}
)

func handleToolCall(ctx context.Context, fs *fs.FileSystem, call openai.ChatCompletionMessageToolCall) (bool, openai.ChatCompletionToolMessageParam) {
	var args map[string]any
	json.Unmarshal([]byte(call.Function.Arguments), &args)

	switch strings.ToLower(call.Function.Name) {
	case "list_files":
		println("📖 list files")

		files, err := fs.ListFiles()

		if err != nil {
			return true, openai.ToolMessage(call.ID, err.Error())
		}

		content, _ := json.Marshal(files)
		return true, openai.ToolMessage(call.ID, string(content))

	case "create_file":
		path := args["path"].(string)
		data := args["content"].(string)

		println("📚 create " + path)

		if err := fs.CreateFile(path, data); err != nil {
			return true, openai.ToolMessage(call.ID, err.Error())
		}

		return true, openai.ToolMessage(call.ID, "file created")

	case "read_file":
		path := args["path"].(string)

		println("📖 read " + path)

		content, err := fs.ReadFile(path)

		if err != nil {
			return true, openai.ToolMessage(call.ID, err.Error())
		}

		return true, openai.ToolMessage(call.ID, content)

	case "delete_file":
		path := args["path"].(string)

		println("🗑️ delete " + path)

		if err := fs.DeleteFile(path); err != nil {
			return true, openai.ToolMessage(call.ID, err.Error())
		}

		return true, openai.ToolMessage(call.ID, "file deleted")

	case "create_dir":
		path := args["path"].(string)

		println("📚 craete " + path)

		if err := fs.CreateDir(path); err != nil {
			return true, openai.ToolMessage(call.ID, err.Error())
		}

		return true, openai.ToolMessage(call.ID, "directory created")

	case "delete_dir":
		path := args["path"].(string)

		println("🗑️ delete " + path)

		if err := fs.DeleteDir(path); err != nil {
			return true, openai.ToolMessage(call.ID, err.Error())
		}

		return true, openai.ToolMessage(call.ID, "directory deleted")
	}

	return false, openai.ToolMessage(call.ID, "unknown function")
}
