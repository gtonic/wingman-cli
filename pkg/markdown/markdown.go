package markdown

import (
	"fmt"
	"io"

	"github.com/charmbracelet/glamour"
)

func Render(w io.Writer, content string) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	md, err := r.Render(content)

	if err != nil {
		fmt.Fprintln(w, content)
		return
	}

	fmt.Fprintln(w, md)
}
