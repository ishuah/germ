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
	ui               *widget.TextGrid
	pty              *os.File
	row, col, cursor int
	stream           chan rune
	buffer           [32][]rune
	commandBuffer    []rune
	bufferIndex      int
	redraw           chan bool
}

func NewTerminal(p *os.File) *Terminal {
	ui := widget.NewTextGrid()
	stream := make(chan rune, 0xffff)
	redraw := make(chan bool)
	terminal := &Terminal{pty: p, ui: ui, stream: stream, buffer: [32][]rune{}, redraw: redraw}
	go terminal.ProcessOutput()
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

func (t *Terminal) ProcessOutput() {
	var b rune
	for {
		b = <-t.stream
		if b == ' ' && len(t.buffer[t.bufferIndex]) > 0 && t.buffer[t.bufferIndex][len(t.buffer[t.bufferIndex])-1] == '$' {
			t.buffer[t.bufferIndex] = append(t.buffer[t.bufferIndex], b)
			t.bufferIndex++
			t.cursor = 0
		} else {
			//t.ui.SetCell(t.row, t.col, widget.TextGridCell{Rune: b})
			t.buffer[t.bufferIndex] = append(t.buffer[t.bufferIndex], b)
		}
		t.Draw(b)
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
			os.Exit(0)
		}
		t.stream <- r
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
