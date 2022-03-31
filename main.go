package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchylinux/twlinst/install"
)

const styling = `
.view.debugTree {
  background-color: #F2F1F0;
}
#confirmInfo {
  background-color: #F2F1F0;
}
.danger > label {
	color: #E23243;
	font-weight: 800;
}
.invalidPassword {
	color: #CC1111;
	font-weight: bold;
}
.active-progress-label {
	font-weight: bold;
	font-size: 135%;
}
`

type pane interface {
	Hide(settings *install.Settings, fullGrid *gtk.Grid) error
	Show(settings *install.Settings, fullGrid *gtk.Grid) error
}

type nextHookPane interface {
	ShouldNext(settings *install.Settings, fullGrid *gtk.Grid) (bool, error)
}

type App struct {
	win      *gtk.Window
	cssProv  *gtk.CssProvider
	fullGrid *gtk.Grid

	cursor   uint
	panes    []pane
	settings install.Settings

	nextBtn, prevBtn *gtk.Button
	abortBtn         *gtk.Button
	versionLab       *gtk.Label
}

func makeApp() (*App, error) {
	a := App{}

	gladePath := "layout.glade"
	if _, err := os.Stat(gladePath); err != nil {
		gladePath = "/etc/twlinst.glade"
	}

	b, err := gtk.BuilderNewFromFile(gladePath)
	if err != nil {
		return nil, err
	}

	obj, err := b.GetObject("window")
	if err != nil {
		return nil, errors.New("couldnt find window in glade file")
	}
	if w, ok := obj.(*gtk.Window); ok {
		a.win = w
	}

	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		return nil, err
	}
	a.cssProv, err = gtk.CssProviderNew()
	if err != nil {
		return nil, err
	}
	gtk.AddProviderForScreen(screen, a.cssProv, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	if err := a.cssProv.LoadFromData(styling); err != nil {
		return nil, err
	}

	obj, err = b.GetObject("fullGrid")
	if err != nil {
		return nil, errors.New("couldnt find fullGrid")
	}
	a.fullGrid = obj.(*gtk.Grid)

	a.win.SetTitle("TwitchyLinux - installer")
	a.win.Connect("destroy", a.callbackWindowDestroy)

	a.panes = []pane{
		initIntroPane(b),
		initSettingsPane(b),
		initHardwarePane(b),
		initConfirmPane(b),
		initInstallPane(b),
		initDonePane(b),
	}
	if err := a.panes[0].Show(&a.settings, a.fullGrid); err != nil {
		return nil, err
	}

	version, err := ioutil.ReadFile("/twl-installer-version")
	if err == nil {
		obj, err = b.GetObject("versionLabel")
		if err != nil {
			return nil, errors.New("couldnt find versionLabel")
		}
		obj.(*gtk.Label).SetText("TwitchyLinux " + strings.TrimSpace(string(version)))
	}

	return &a, a.build(b)
}

func (a *App) build(b *gtk.Builder) error {
	// Buttons on the bottom.
	obj, err := b.GetObject("abortBtn")
	if err != nil {
		return errors.New("couldnt find abortBtn")
	}
	a.abortBtn = obj.(*gtk.Button)
	a.abortBtn.Connect("clicked", a.callbackWindowDestroy)
	obj, err = b.GetObject("nextBtn")
	if err != nil {
		return errors.New("couldnt find nextBtn")
	}
	a.nextBtn = obj.(*gtk.Button)
	a.nextBtn.Connect("clicked", a.callbackNext)
	obj, err = b.GetObject("previousBtn")
	if err != nil {
		return errors.New("couldnt find prevBtn")
	}
	a.prevBtn = obj.(*gtk.Button)
	a.prevBtn.Connect("clicked", a.callbackPrev)
	a.prevBtn.SetSensitive(false)

	return nil
}

func (a *App) callbackWindowDestroy() {
	gtk.MainQuit()
}

func (a *App) callbackNext() {
	if int(a.cursor+1) >= len(a.panes) {
		return
	}

	if p, ok := a.panes[a.cursor].(nextHookPane); ok {
		shouldNext, err := p.ShouldNext(&a.settings, a.fullGrid)
		if err != nil {
			panic(err)
		}
		if !shouldNext {
			return
		}
	}

	if err := a.panes[a.cursor].Hide(&a.settings, a.fullGrid); err != nil {
		panic(err)
	}
	a.cursor++
	a.prevBtn.SetSensitive(a.cursor > 0)
	if err := a.panes[a.cursor].Show(&a.settings, a.fullGrid); err != nil {
		panic(err)
	}
	a.nextBtn.SetSensitive(int(a.cursor+1) < len(a.panes))
}

func (a *App) callbackPrev() {
	if a.cursor == 0 {
		return
	}
	if err := a.panes[a.cursor].Hide(&a.settings, a.fullGrid); err != nil {
		panic(err)
	}
	a.cursor--
	if err := a.panes[a.cursor].Show(&a.settings, a.fullGrid); err != nil {
		panic(err)
	}
	a.nextBtn.SetSensitive(int(a.cursor+1) < len(a.panes))
	a.prevBtn.SetSensitive(a.cursor > 0)
}

func (a *App) mainloop() {
	a.win.SetDefaultSize(800, 600)
	a.win.ShowAll()
	gtk.Main()
}

func mainApp(args []string) {
	gtk.Init(&args)

	a, err := makeApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	a.mainloop()
}

var (
	configFlag = flag.String("config", "", "Install using the provided configuration instead of automatically.")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if *configFlag == "" {
		mainApp(args)
	} else {
		// Non-interactive!
		noninteractiveInstall()
	}
}

func noninteractiveInstall() {
	f, err := os.Open(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Reading config: %v\n", err)
		os.Exit(1)
	}
	var conf install.Settings
	if err := json.NewDecoder(f).Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Decoding config: %v\n", err)
		f.Close()
		os.Exit(1)
	}
	f.Close()

	disks, err := getDiskInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Reading disks: %v\n", err)
		os.Exit(1)
	}
	for _, d := range disks {
		if d.Path == conf.ConfigOnlyDisk {
			conf.Disk = d
		}
	}
	if conf.Disk.Path == "" {
		fmt.Fprintf(os.Stderr, "Couldnt find install disk %q\n", conf.ConfigOnlyDisk)
		os.Exit(1)
	}

	// Print updates on the screen
	upChan := make(chan install.Update, 1)
	go func() {
		for msg := range upChan {
			if msg.Step != "" {
				fmt.Printf("Starting stage %s\n", msg.Step)
			}
			if msg.Msg != "" {
				fmt.Print(msg.Msg)
			}
		}
	}()

	run := install.Configure(upChan, conf)
	if err := run.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Install init failed: %v\n", err)
		os.Exit(1)
	}
	run.Wait()
}
