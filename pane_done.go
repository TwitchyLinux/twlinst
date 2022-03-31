package main

import (
	"fmt"
	"os/exec"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchylinux/twlinst/install"
)

type donePane struct {
	shutdownBtn *gtk.Button
	content     *gtk.Grid
}

func initDonePane(b *gtk.Builder) *donePane {
	obj, err := b.GetObject("contentGrid_done")
	if err != nil {
		panic("couldnt find contentGrid_done")
	}
	content := obj.(*gtk.Grid)

	obj, err = b.GetObject("fullGrid_done")
	if err != nil {
		panic("couldnt find fullGrid_done")
	}
	obj.(*gtk.Grid).Remove(content)

	obj, err = b.GetObject("done_shutdownBtn")
	if err != nil {
		panic("couldnt find done_shutdownBtn")
	}
	shutdownBtn := obj.(*gtk.Button)

	p := &donePane{
		shutdownBtn: shutdownBtn,
		content:     content,
	}

	shutdownBtn.Connect("clicked", p.callbackShutdown)
	return p
}

func (p *donePane) callbackShutdown() {
	exec.Command("sudo", "shutdown", "-h", "0").Run()
}

func (p *donePane) Show(settings *install.Settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)
	return nil
}

func (p *donePane) Hide(settings *install.Settings, fullGrid *gtk.Grid) error {
	currentPane, err := fullGrid.GetChildAt(0, 1)
	if err != nil {
		return fmt.Errorf("Failed to get current pane: %v", err)
	}
	fullGrid.Remove(currentPane)
	return nil
}
