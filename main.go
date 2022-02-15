package main

import (
	"fmt"
	"os"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/creack/pty"
)

func eval(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("germ")

	c := exec.Command("/bin/bash")
	p, err := pty.Start(c)
	eval(err)

	defer c.Process.Kill()

	os.Setenv("TERM", "xterm-256color")
	terminal := NewTerminal(p)

	w.SetContent(
		container.New(
			layout.NewGridWrapLayout(fyne.NewSize(630, 630)),
			terminal,
		),
	)
	w.Canvas().Focus(terminal)
	// w.Canvas().SetOnTypedKey(terminal.OnTypedKey)
	// w.Canvas().SetOnTypedRune(terminal.OnTypedRune)

	w.ShowAndRun()
}
