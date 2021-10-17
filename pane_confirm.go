package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

type confirmPane struct {
	content *gtk.Grid
}

func initConfirmPane(b *gtk.Builder) *confirmPane {
	obj, err := b.GetObject("contentGrid_confirmation")
	if err != nil {
		panic("couldnt find contentGrid_confirmation")
	}

	content := obj.(*gtk.Grid)
	obj, err = b.GetObject("fullGrid_confirmation")
	if err != nil {
		panic("couldnt find fullGrid_confirmation")
	}
	obj.(*gtk.Grid).Remove(content)

	return &confirmPane{content}
}

func (p *confirmPane) Show(settings *settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)
	return nil
}

func (p *confirmPane) Hide(settings *settings, fullGrid *gtk.Grid) error {
	currentPane, err := fullGrid.GetChildAt(0, 1)
	if err != nil {
		return fmt.Errorf("Failed to get current pane: %v", err)
	}
	fullGrid.Remove(currentPane)
	return nil
}
