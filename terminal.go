package main

import (
	"bufio"
	"image/color"
	"io"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Terminal struct {
	ui               *widget.TextGrid
	scroll           *container.Scroll
	pty              *os.File
	row, col, cursor int
	buffer           [32][]rune
	commandBuffer    []rune
	bufferIndex      int
}

func NewTerminal(p *os.File) *Terminal {
	ui := widget.NewTextGrid()
	scroll := container.NewScroll(ui)
	terminal := &Terminal{pty: p, ui: ui, scroll: scroll, buffer: [32][]rune{}}
	go terminal.Read()
	go terminal.Blink()
	return terminal
}

func (t *Terminal) Draw(r rune) {
	if r == '\n' {
		t.row++
		t.col = 0
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
		if b == ' ' && len(t.buffer[t.bufferIndex]) > 0 && t.buffer[t.bufferIndex][len(t.buffer[t.bufferIndex])-1] == '$' {
			t.buffer[t.bufferIndex] = append(t.buffer[t.bufferIndex], b)
			t.bufferIndex++
			t.cursor = 0
		} else {
			t.buffer[t.bufferIndex] = append(t.buffer[t.bufferIndex], b)
		}
		t.Draw(b)
	}
	t.scroll.ScrollToBottom()
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

func (t *Terminal) OnTypedKey(e *fyne.KeyEvent) {
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
			//_, _ = t.pty.Write([]byte{8})
		}
	case fyne.KeyUp:
		_, _ = t.pty.Write([]byte{27, '[', 'A'})
	case fyne.KeyDown:
		_, _ = t.pty.Write([]byte{27, '[', 'B'})
	}
}

func (t *Terminal) OnTypedRune(r rune) {
	t.ui.SetCell(t.row, t.col+t.cursor, widget.TextGridCell{Rune: r})
	t.cursor++
	t.commandBuffer = append(t.commandBuffer, r)
	//_, _ = t.pty.WriteString(string(r))
}
