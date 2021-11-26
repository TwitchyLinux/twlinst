package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchylinux/twlinst/install"
)

type introPane struct {
	content *gtk.Grid
}

func initIntroPane(b *gtk.Builder) *introPane {
	obj, err := b.GetObject("contentGrid")
	if err != nil {
		panic(err)
	}
	return &introPane{
		content: obj.(*gtk.Grid),
	}
}

func (p *introPane) Show(settings *install.Settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)
	return nil
}

func (p *introPane) Hide(settings *install.Settings, fullGrid *gtk.Grid) error {
	currentPane, err := fullGrid.GetChildAt(0, 1)
	if err != nil {
		return fmt.Errorf("Failed to get current pane: %v", err)
	}
	fullGrid.Remove(currentPane)
	return nil
}
