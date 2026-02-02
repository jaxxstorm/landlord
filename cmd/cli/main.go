package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/fang"
)

func main() {
	cmd := newRootCommand()
	if err := fang.Execute(
		context.Background(),
		cmd,
		fang.WithErrorHandler(func(w io.Writer, styles fang.Styles, err error) {
			if err == nil {
				return
			}
			fmt.Fprintln(w, errorStyle.Render(err.Error()))
		}),
	); err != nil {
		os.Exit(1)
	}
}
