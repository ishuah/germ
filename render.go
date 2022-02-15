package main

import "fyne.io/fyne/v2"

type render struct {
	term *Terminal
}

func (r *render) Layout(s fyne.Size) {
	r.term.ui.Resize(s)
}

func (r *render) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *render) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.term.ui}
}

func (r *render) Refresh() {
	r.term.ui.Refresh()
}

func (r *render) Destroy() {
}

func (t *Terminal) CreateRenderer() fyne.WidgetRenderer {
	r := &render{term: t}
	return r
}
