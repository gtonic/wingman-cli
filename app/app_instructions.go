package app

import (
	"os"
	"strings"

	"github.com/adrianliechti/wingman/pkg/template"
)

func MustParseInstructions() string {
	prompt, err := ParseInstructions()

	if err != nil {
		panic(err)
	}

	return prompt
}

func ParseInstructions() (string, error) {
	candidates := []string{
		".instructions.md",
		".instructions.txt",

		"instructions.md",
		"instructions.txt",

		".prompt.md",
		".prompt.txt",

		"prompt.md",
		"prompt.txt",
	}

	for _, name := range candidates {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(name)

		if err != nil {
			return "", err
		}

		instructions := strings.TrimSpace(string(data))

		tmpl, err := template.NewTemplate(instructions)

		if err != nil {
			return "", err
		}

		instructions, err = tmpl.Execute(nil)

		if err != nil {
			return "", err
		}

		return instructions, nil
	}

	return "", nil
}
