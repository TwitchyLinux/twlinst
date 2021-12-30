package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchylinux/twlinst/install"
)

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// void createInstallViewTags(GtkTextBuffer* buff) {
//   gtk_text_buffer_create_tag(buff, "cmdMsg", "weight", "700", NULL);
//   gtk_text_buffer_create_tag(buff, "warnMsg", "foreground", "orange", NULL);
//   gtk_text_buffer_create_tag(buff, "errMsg", "foreground", "red", NULL);
// }
// void textbuffDelete(GtkTextBuffer *buff, GtkTextIter *cursor, GtkTextIter* end) {
//   gtk_text_buffer_delete(buff, cursor, end);
// }
import "C"

type installPane struct {
	updateCh chan install.Update
	prev     *gtk.Button
	content  *gtk.Grid

	stepFormatLabel    *gtk.Label
	stepCopyLabel      *gtk.Label
	stepConfigureLabel *gtk.Label
	stepCleanupLabel   *gtk.Label

	outputText   *gtk.TextView
	outputBuffer *gtk.TextBuffer
	scroll       *gtk.ScrolledWindow
}

func initInstallPane(b *gtk.Builder) *installPane {
	obj, err := b.GetObject("contentGrid_progress")
	if err != nil {
		panic("couldnt find contentGrid_progress")
	}
	content := obj.(*gtk.Grid)

	obj, err = b.GetObject("progressstep_1")
	if err != nil {
		panic("couldnt find progressstep_1")
	}
	stepFormatLabel := obj.(*gtk.Label)
	obj, err = b.GetObject("progressstep_2")
	if err != nil {
		panic("couldnt find progressstep_2")
	}
	stepCopyLabel := obj.(*gtk.Label)
	obj, err = b.GetObject("progressstep_3")
	if err != nil {
		panic("couldnt find progressstep_3")
	}
	stepConfigureLabel := obj.(*gtk.Label)
	obj, err = b.GetObject("progressstep_4")
	if err != nil {
		panic("couldnt find progressstep_4")
	}
	stepCleanupLabel := obj.(*gtk.Label)

	obj, err = b.GetObject("outputProgressText")
	if err != nil {
		panic("couldnt find outputProgressText")
	}
	output := obj.(*gtk.TextView)
	obj, err = b.GetObject("outputProgressScroller")
	if err != nil {
		panic("couldnt find outputProgressScroller")
	}
	scroll := obj.(*gtk.ScrolledWindow)

	ttt, err := gtk.TextTagTableNew()
	if err != nil {
		panic(err)
	}
	textBuffer, err := gtk.TextBufferNew(ttt)
	if err != nil {
		panic(err)
	}
	output.SetBuffer(textBuffer)
	C.createInstallViewTags((*C.GtkTextBuffer)(unsafe.Pointer(textBuffer.Native())))

	obj, err = b.GetObject("fullGrid_progress")
	if err != nil {
		panic("couldnt find fullGrid_progress")
	}
	obj.(*gtk.Grid).Remove(content)

	obj, err = b.GetObject("previousBtn")
	if err != nil {
		panic(err)
	}
	prev := obj.(*gtk.Button)

	updateCh := make(chan install.Update)
	p := &installPane{
		updateCh,
		prev,
		content,
		stepFormatLabel,
		stepCopyLabel,
		stepConfigureLabel,
		stepCleanupLabel,
		output,
		textBuffer,
		scroll,
	}

	go p.updater()
	return p
}

func (p *installPane) updater() {
	applyBold := func(lab *gtk.Label, bold bool) {
		sc, _ := lab.GetStyleContext()
		if bold {
			sc.AddClass("active-progress-label")
		} else {
			sc.RemoveClass("active-progress-label")
		}
	}

	var outText string
	sync := make(chan bool)
	for msg := range p.updateCh {
		if msg.Step != "" {
			glib.IdleAdd(func() {
				applyBold(p.stepFormatLabel, msg.Step == "format")
				applyBold(p.stepConfigureLabel, msg.Step == "configure")
				applyBold(p.stepCopyLabel, msg.Step == "copy")
				applyBold(p.stepCleanupLabel, msg.Step == "cleanup")
			})
		}

		if msg.TrimLastLine {
			// Cut out up to the 2nd-last newline.
			i := strings.LastIndex(outText, "\n")
			if i > 0 {
				i = strings.LastIndex(outText[:i], "\n")
				if i > 0 {
					i++
					outText = outText[:i]
					glib.IdleAdd(func() {
						curs := p.outputBuffer.GetIterAtOffset(i)
						end := p.outputBuffer.GetEndIter()
						C.textbuffDelete((*C.GtkTextBuffer)(unsafe.Pointer(p.outputBuffer.Native())), (*C.GtkTextIter)(unsafe.Pointer(curs)), (*C.GtkTextIter)(unsafe.Pointer(end)))
						sync <- true
					})
					<-sync
				}
			}
		}

		if msg.Msg != "" {
			outText += msg.Msg
			glib.IdleAdd(func() {
				p.outputBuffer.InsertAtCursor(msg.Msg)

				switch msg.Level {
				case install.MsgWarn:
					p.outputBuffer.ApplyTagByName("warnMsg", p.outputBuffer.GetIterAtOffset(len(outText)-len(msg.Msg)), p.outputBuffer.GetIterAtOffset(len(outText)))
				case install.MsgCmd:
					p.outputBuffer.ApplyTagByName("cmdMsg", p.outputBuffer.GetIterAtOffset(len(outText)-len(msg.Msg)), p.outputBuffer.GetIterAtOffset(len(outText)))
				case install.MsgErr:
					p.outputBuffer.ApplyTagByName("errMsg", p.outputBuffer.GetIterAtOffset(len(outText)-len(msg.Msg)), p.outputBuffer.GetIterAtOffset(len(outText)))
				}
				sync <- true
			})
			<-sync
		}

		if !msg.TrimLastLine {
			glib.IdleAdd(func() {
				adj := p.scroll.GetVAdjustment()
				adj.SetValue(adj.GetUpper())
				sync <- true
			})
			<-sync
		}
	}
}

func (p *installPane) Show(settings *install.Settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)
	p.prev.SetSensitive(false)

	run := install.Configure(p.updateCh, *settings)
	return run.Start()
}

func (p *installPane) Hide(settings *install.Settings, fullGrid *gtk.Grid) error {
	currentPane, err := fullGrid.GetChildAt(0, 1)
	if err != nil {
		return fmt.Errorf("Failed to get current pane: %v", err)
	}
	fullGrid.Remove(currentPane)
	return nil
}
