package main

import (
	"bufio"
	"image/color"
	"io"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Terminal struct {
	ui       *widget.TextGrid
	pty      *os.File
	row, col int
	stream   chan rune
}

func NewTerminal(p *os.File) *Terminal {
	ui := widget.NewTextGrid()
	stream := make(chan rune, 0xffff)
	terminal := &Terminal{pty: p, ui: ui, stream: stream}
	go terminal.Write()
	go terminal.Read()
	go terminal.Blink()
	return terminal
}

func (t *Terminal) Write() {
	var b rune
	for {
		b = <-t.stream
		if b == '\n' {
			t.row++
			t.col = 0
		} else {
			t.ui.SetCell(t.row, t.col, widget.TextGridCell{Rune: b})
		}

		if t.col >= 60 {
			t.row++
			t.col = 0
		} else {
			t.col++
		}
	}
}

func (t *Terminal) Read() {
	reader := bufio.NewReader(t.pty)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return
			}
			os.Exit(1)
		}
		t.stream <- r
	}
}

func (t *Terminal) Blink() {
	blink := true
	for {
		time.Sleep(1 * time.Second)
		if blink {
			t.ui.SetCell(t.row, t.col, widget.TextGridCell{Style: &widget.CustomTextGridStyle{BGColor: color.White}})
		} else {
			t.ui.SetCell(t.row, t.col, widget.TextGridCell{Style: &widget.CustomTextGridStyle{BGColor: theme.BackgroundColor()}})
		}
		blink = !blink
	}
}

func (t *Terminal) OnTypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		_, _ = t.pty.Write([]byte{'\n'})
	case fyne.KeyEscape:
		_, _ = t.pty.Write([]byte{27})
	case fyne.KeyBackspace:
		t.ui.SetCell(t.row, t.col, widget.TextGridCell{Rune: ' '})
		t.col--
		t.ui.SetCell(t.row, t.col, widget.TextGridCell{Rune: ' '})
		_, _ = t.pty.Write([]byte{8})
	case fyne.KeyUp:
		_, _ = t.pty.Write([]byte{27, '[', 'A'})
	case fyne.KeyDown:
		_, _ = t.pty.Write([]byte{27, '[', 'B'})
	}
}

func (t *Terminal) OnTypedRune(r rune) {
	t.ui.SetCell(t.row, t.col, widget.TextGridCell{Rune: r})
	t.col++
	_, _ = t.pty.WriteString(string(r))
}
