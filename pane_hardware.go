package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchylinux/twlinst/install"
)

var (
	nixosHardwareJson = flag.String("nixos-hardware-json", "/nixos-hardware-info.json", "Path to a JSON file describing available hardware profiles.")
)

func makeColumn(name string, id int) *gtk.TreeViewColumn {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		panic(err)
	}
	column, err := gtk.TreeViewColumnNewWithAttribute(name, cellRenderer, "text", id)
	if err != nil {
		panic(err)
	}
	return column
}

type hardwarePane struct {
	treeView  *gtk.TreeView
	treeStore *gtk.TreeStore
	content   *gtk.Grid

	selection string
}

func initHardwarePane(b *gtk.Builder) *hardwarePane {
	obj, err := b.GetObject("contentGrid_hardware")
	if err != nil {
		panic("couldnt find contentGrid_hardware")
	}
	content := obj.(*gtk.Grid)

	obj, err = b.GetObject("fullGrid_hardware")
	if err != nil {
		panic("couldnt find fullGrid_hardware")
	}
	obj.(*gtk.Grid).Remove(content)

	treeView, err := gtk.TreeViewNew()
	if err != nil {
		panic(fmt.Errorf("creating treeview: %w", err))
	}
	treeView.SetVExpand(true)
	treeView.SetHExpand(true)
	treeView.SetMarginTop(4)
	treeView.SetMarginBottom(4)
	treeView.SetMarginStart(4)
	treeView.SetMarginEnd(4)
	treeView.AppendColumn(makeColumn("Device", 0))

	selection, err := treeView.GetSelection()
	if err != nil {
		panic(err)
	}
	selection.SetMode(gtk.SELECTION_SINGLE)

	treeStore, err := gtk.TreeStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		panic(fmt.Errorf("creating treeview: %w", err))
	}
	treeView.SetModel(treeStore)

	sw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		panic(fmt.Errorf("creating scrolled window: %w", err))
	}
	sw.Add(treeView)

	content.Attach(sw, 0, 1, 1, 1)
	content.ShowAll()

	pane := &hardwarePane{treeView, treeStore, content, ""}
	selection.Connect("changed", pane.selectionChanged)

	pane.appendRow(nil, "<none>", "<none>")
	pane.populateFromFile()

	return pane
}

func (p *hardwarePane) populateFromFile() {
	f, err := os.Open(*nixosHardwareJson)
	if err != nil {
		return
	}
	defer f.Close()

	var profiles map[string]string
	if err := json.NewDecoder(f).Decode(&profiles); err != nil {
		return
	}

	sortedPrefixes := make([]string, 0, 32)
	seen := make(map[string]struct{})
	for k := range profiles {
		if idx := strings.Index(k, "-"); idx > 0 {
			if _, exists := seen[k[:idx]]; !exists {
				seen[k[:idx]] = struct{}{}
				sortedPrefixes = append(sortedPrefixes, k[:idx])
			}
		}
	}
	sort.Strings(sortedPrefixes)

	prefixesTree := map[string]*gtk.TreeIter{}
	for _, prefix := range sortedPrefixes {
		prefixesTree[prefix] = p.appendRow(nil, prefix, "")
	}

	for profile, path := range profiles {
		if strings.HasPrefix(profile, "_") {
			continue
		}
		if idx := strings.Index(profile, "-"); idx > 0 && prefixesTree[profile[:idx]] != nil {
			p.appendRow(prefixesTree[profile[:idx]], profile[idx+1:], path)
		} else {
			p.appendRow(nil, profile, path)
		}
	}
}

func (p *hardwarePane) Show(settings *install.Settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)
	return nil
}

func (p *hardwarePane) Hide(settings *install.Settings, fullGrid *gtk.Grid) error {
	currentPane, err := fullGrid.GetChildAt(0, 1)
	if err != nil {
		return fmt.Errorf("Failed to get current pane: %v", err)
	}
	fullGrid.Remove(currentPane)
	return nil
}

func (p *hardwarePane) selectionChanged(selection *gtk.TreeSelection) {
	var (
		iter  *gtk.TreeIter
		model gtk.ITreeModel
		ok    bool
	)

	if model, iter, ok = selection.GetSelected(); ok {
		v, _ := model.(*gtk.TreeModel).GetValue(iter, 1)
		p.selection, _ = v.GetString()
	} else {
		p.selection = ""
	}
}

func (p *hardwarePane) ShouldNext(settings *install.Settings, fullGrid *gtk.Grid) (bool, error) {
	if p.selection == "<none>" {
		settings.NixosHardwareImport = ""
	} else {
		settings.NixosHardwareImport = p.selection
	}
	return p.selection != "", nil
}

func (p *hardwarePane) appendRow(maybeParent *gtk.TreeIter, text, path string) *gtk.TreeIter {
	i := p.treeStore.Append(maybeParent)
	err := p.treeStore.SetValue(i, 0, text)
	if err != nil {
		panic(err)
	}
	err = p.treeStore.SetValue(i, 1, path)
	if err != nil {
		panic(err)
	}
	return i
}
