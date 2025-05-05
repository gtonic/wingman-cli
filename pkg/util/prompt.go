package util

import (
	"os"
	"strings"

	"github.com/adrianliechti/wingman/pkg/template"
)

func ParsePrompt() (string, error) {
	for _, name := range []string{".prompt.md", ".prompt.txt", "prompt.md", "prompt.txt"} {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(name)

		if err != nil {
			return "", err
		}

		prompt := strings.TrimSpace(string(data))

		tmpl, err := template.NewTemplate(prompt)

		if err != nil {
			return "", err
		}

		prompt, err = tmpl.Execute(nil)

		if err != nil {
			return "", err
		}

		return prompt, nil
	}

	return "", nil
}
