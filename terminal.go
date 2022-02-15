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
	widget.BaseWidget
	fyne.ShortcutHandler
	ui               *widget.TextGrid
	pty              *os.File
	row, col, cursor int
	commandBuffer    []rune
}

func NewTerminal(p *os.File) *Terminal {
	ui := widget.NewTextGrid()
	terminal := &Terminal{}
	terminal.ExtendBaseWidget(terminal)
	terminal.ui = ui
	terminal.pty = p

	go terminal.Read()
	go terminal.Blink()
	return terminal
}

func (t *Terminal) Draw(r rune) {
	if r == '\n' {
		t.row++
		t.col = 0
		t.cursor = 0
		return
	}

	t.ui.SetCell(t.row, t.col, widget.TextGridCell{Rune: r})
	if t.col >= 60 {
		t.row++
		t.col = 0
	} else {
		t.col++
	}
}

func (t *Terminal) ProcessOutput(buffer []byte) {
	for _, b := range string(buffer) {
		t.Draw(b)
	}
}

func (t *Terminal) Read() {
	reader := bufio.NewReader(t.pty)
	bufferSize := 4069
	buffer := make([]byte, bufferSize)
	for {
		size, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return
			}
			os.Exit(0)
		}
		t.ProcessOutput(buffer[:size])
	}
}

func (t *Terminal) Blink() {
	blink := true
	for {
		time.Sleep(500 * time.Millisecond)
		if blink {
			t.ui.SetCell(t.row, t.col+t.cursor, widget.TextGridCell{Style: &widget.CustomTextGridStyle{BGColor: color.White}})
		} else {
			t.ui.SetCell(t.row, t.col+t.cursor, widget.TextGridCell{Style: &widget.CustomTextGridStyle{BGColor: theme.BackgroundColor()}})
		}
		blink = !blink
	}
}

func (t *Terminal) TypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		_, _ = t.pty.WriteString(string(t.commandBuffer))
		t.commandBuffer = nil
		_, _ = t.pty.Write([]byte{'\n'})
	case fyne.KeyEscape:
		_, _ = t.pty.Write([]byte{27})
	case fyne.KeyBackspace:
		if t.cursor > 0 {
			t.ui.SetCell(t.row, t.col+t.cursor, widget.TextGridCell{Rune: ' '})
			t.cursor--
			t.ui.SetCell(t.row, t.col+t.cursor, widget.TextGridCell{Rune: ' '})
			t.commandBuffer = t.commandBuffer[:len(t.commandBuffer)-1]
		}
	case fyne.KeyUp:
		_, _ = t.pty.Write([]byte{27, '[', 'A'})
	case fyne.KeyDown:
		_, _ = t.pty.Write([]byte{27, '[', 'B'})
	}
}

func (t *Terminal) TypedRune(r rune) {
	t.ui.SetCell(t.row, t.col+t.cursor, widget.TextGridCell{Rune: r})
	t.cursor++
	t.commandBuffer = append(t.commandBuffer, r)
}

func (t *Terminal) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*fyne.ShortcutCopy); ok {
		_, _ = t.pty.Write([]byte{0x3})
	}
}

func (t *Terminal) FocusGained() {
	t.Refresh()
}

func (t *Terminal) FocusLost() {
	t.Refresh()
}
