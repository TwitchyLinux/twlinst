package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
)

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// void createConfirmTags(GtkTextBuffer* buff) {
//   gtk_text_buffer_create_tag(buff, "settingName", "weight", "700", NULL);
//   gtk_text_buffer_create_tag(buff, "warning", "foreground", "red", NULL);
// }
// void createInstallViewTags(GtkTextBuffer* buff) {
//   gtk_text_buffer_create_tag(buff, "cmdMsg", "weight", "700", NULL);
//   gtk_text_buffer_create_tag(buff, "warnMsg", "foreground", "orange", NULL);
//   gtk_text_buffer_create_tag(buff, "errMsg", "foreground", "red", NULL);
// }
import "C"

type confirmPane struct {
	ttt        *gtk.TextTagTable
	textBuffer *gtk.TextBuffer
	content    *gtk.Grid
}

func initConfirmPane(b *gtk.Builder) *confirmPane {
	obj, err := b.GetObject("contentGrid_confirmation")
	if err != nil {
		panic("couldnt find contentGrid_confirmation")
	}
	content := obj.(*gtk.Grid)

	obj, err = b.GetObject("confirmInfo")
	if err != nil {
		panic("couldnt find confirmInfo")
	}
	confirmView := obj.(*gtk.TextView)

	obj, err = b.GetObject("fullGrid_confirmation")
	if err != nil {
		panic("couldnt find fullGrid_confirmation")
	}
	obj.(*gtk.Grid).Remove(content)

	ttt, err := gtk.TextTagTableNew()
	if err != nil {
		panic(err)
	}
	textBuffer, err := gtk.TextBufferNew(ttt)
	if err != nil {
		panic(err)
	}
	confirmView.SetBuffer(textBuffer)
	C.createConfirmTags((*C.GtkTextBuffer)(unsafe.Pointer(textBuffer.Native())))

	return &confirmPane{ttt, textBuffer, content}
}

type textviewStyleSelection struct {
	Start, End int
	Class      string
}

func (p *confirmPane) Show(settings *settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)

	var outText string
	var styles []textviewStyleSelection
	writeStyled := func(text, class string) {
		if class != "" {
			styles = append(styles, textviewStyleSelection{
				Start: len(outText),
				End:   len(outText) + len(text),
				Class: class,
			})
		}
		outText += text
	}

	writeStyled("Hostname: ", "settingName")
	writeStyled(settings.Hostname, "")
	writeStyled("\n", "")
	writeStyled("Username: ", "settingName")
	writeStyled(settings.Username, "")
	writeStyled("\n", "")
	writeStyled("Password: ", "settingName")
	writeStyled(strings.Repeat("*", len(settings.Password)), "")
	writeStyled("\n", "")
	writeStyled("Timezone: ", "settingName")
	writeStyled(settings.Timezone, "")
	writeStyled("\n", "")

	writeStyled("Install to:\n  Path: ", "settingName")
	writeStyled(settings.Disk.Path+"\n", "")
	writeStyled("  Name: ", "settingName")
	if settings.Disk.Label != "" {
		writeStyled(settings.Disk.Label+"\n        ", "")
	}
	writeStyled(settings.Disk.Model+" ("+settings.Disk.Serial+")", "")
	writeStyled("\n\n", "")
	writeStyled("  Capacity: ", "settingName")
	writeStyled(byteCountDecimal(int64(settings.Disk.NumBlocks*512))+"\n", "")
	writeStyled("  UUID: ", "settingName")
	writeStyled(settings.Disk.PartUUID, "")
	writeStyled("\n", "")
	writeStyled("  Partitions: ", "settingName")
	writeStyled(fmt.Sprintf("%d read from %s table", len(settings.Disk.Partitions), settings.Disk.PartTabType), "")
	writeStyled("\n", "")
	for i, part := range settings.Disk.Partitions {
		writeStyled(fmt.Sprintf("   %2d: ", i), "settingName")
		writeStyled(fmt.Sprintf("%s filesystem on %s\n", part.FS, part.Name), "")
		writeStyled(fmt.Sprintf("      Filesystem UUID: %s\n", part.FsUUID), "")
		writeStyled(fmt.Sprintf("      Partition UUID: %s\n", part.PartUUID), "")
	}
	writeStyled("  WARNING: Any existing data on this disk will be lost.\n", "warning")
	if settings.Scrub {
		writeStyled("  Zeros will be written to disk (scrubbing) before formatting.\n", "")
	} else {
		writeStyled("  WARNING: Encrypted area will not be scrubbed. This may reveal information\n", "warning")
		writeStyled("  about the usage patterns of your system if your disk is examined.\n", "warning")
	}

	p.textBuffer.SetText(outText)
	for _, t := range styles {
		if t.Class != "" {
			p.textBuffer.ApplyTagByName(t.Class, p.textBuffer.GetIterAtOffset(t.Start), p.textBuffer.GetIterAtOffset(t.End))
		}
	}

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
