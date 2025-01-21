package markdown

import (
	"fmt"
	"io"

	"github.com/charmbracelet/glamour"
)

func Render(w io.Writer, content string) {
	md, err := glamour.Render(content, "auto")

	if err != nil {
		fmt.Fprintln(w, content)
		return
	}

	fmt.Fprintln(w, md)
}
