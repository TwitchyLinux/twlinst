package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"unicode"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twitchylinux/twlinst/install"
	"github.com/twitchylinux/twlinst/z"
)

type settingsPane struct {
	content *gtk.Grid

	hostCtrl, userCtrl *gtk.Entry

	tzCtrl, diskCtrl   *gtk.ComboBoxText
	pwCtrl, pwConfirm  *gtk.Entry
	pwLabel, hostLabel *gtk.Label
	userLabel          *gtk.Label
	scrubCheck         *gtk.CheckButton
	loginCheck         *gtk.CheckButton

	disks []z.Disk
}

func initSettingsPane(b *gtk.Builder) *settingsPane {
	obj, err := b.GetObject("contentGrid_settings")
	if err != nil {
		panic("couldnt find contentGrid_settings")
	}

	content := obj.(*gtk.Grid)
	obj, err = b.GetObject("fullGrid_settings")
	if err != nil {
		panic("couldnt find fullGrid_settings")
	}
	obj.(*gtk.Grid).Remove(content)

	obj, err = b.GetObject("timezoneCombo")
	if err != nil {
		panic("couldnt find timezoneCombo")
	}
	tzCtrl := obj.(*gtk.ComboBoxText)
	obj, err = b.GetObject("installDiskCombo")
	if err != nil {
		panic("couldnt find installDiskCombo")
	}
	diskCtrl := obj.(*gtk.ComboBoxText)

	obj, err = b.GetObject("passwordInput")
	if err != nil {
		panic("couldnt find passwordInput")
	}
	pwCtrl := obj.(*gtk.Entry)

	obj, err = b.GetObject("confirmPasswordInput")
	if err != nil {
		panic("couldnt find confirmPasswordInput")
	}
	pwConfirm := obj.(*gtk.Entry)

	obj, err = b.GetObject("passwordLabel")
	if err != nil {
		panic("couldnt find passwordLabel")
	}
	pwLabel := obj.(*gtk.Label)
	obj, err = b.GetObject("hostLabel")
	if err != nil {
		panic("couldnt find hostLabel")
	}
	hostLabel := obj.(*gtk.Label)
	obj, err = b.GetObject("userLabel")
	if err != nil {
		panic("couldnt find userLabel")
	}
	userLabel := obj.(*gtk.Label)

	obj, err = b.GetObject("hostnameInput")
	if err != nil {
		panic("couldnt find hostnameInput")
	}
	hostCtrl := obj.(*gtk.Entry)
	// hostCtrl.Connect("changed", mw.callbackSettingsTyped)

	obj, err = b.GetObject("usernameInput")
	if err != nil {
		panic("couldnt find usernameInput")
	}
	userCtrl := obj.(*gtk.Entry)
	// userCtrl.Connect("changed", mw.callbackSettingsTyped)

	obj, err = b.GetObject("clearDiskCheck")
	if err != nil {
		panic("couldnt find clearDiskCheck")
	}
	scrubCheck := obj.(*gtk.CheckButton)
	// scrubCheck.Connect("toggled", mw.callbackSettingsTyped)
	obj, err = b.GetObject("autologinCheck")
	if err != nil {
		panic("couldnt find autologinCheck")
	}
	loginCheck := obj.(*gtk.CheckButton)
	// loginCheck.Connect("toggled", mw.callbackSettingsTyped)

	disks, err := getDiskInfo()
	if err != nil {
		panic(err)
	}
	for i, d := range disks {
		diskCtrl.Append(fmt.Sprint(i), fmt.Sprintf("%s (%s) - %s bus, %s partition table", d.Path, d.Model, d.Bus, d.PartTabType))
	}
	diskCtrl.SetActive(0)

	tzs := timezones()
	for _, tz := range tzs {
		tzCtrl.Append(tz, tz)
	}
	tzCtrl.SetActive(0)

	p := &settingsPane{
		content,
		hostCtrl, userCtrl,
		tzCtrl, diskCtrl,
		pwCtrl, pwConfirm, pwLabel,
		hostLabel, userLabel,
		scrubCheck, loginCheck,
		disks,
	}
	pwCtrl.Connect("changed", p.callbackPwChanged)
	pwConfirm.Connect("changed", p.callbackPwChanged)
	hostCtrl.Connect("changed", p.callbackHostChanged)
	userCtrl.Connect("changed", p.callbackUserChanged)
	return p
}

func (p *settingsPane) callbackPwChanged() {
	mainPw, _ := p.pwCtrl.GetText()
	confPw, _ := p.pwConfirm.GetText()

	sc, _ := p.pwLabel.GetStyleContext()

	if mainPw == confPw && mainPw != "" {
		sc.AddClass("validPassword")
		sc.RemoveClass("invalidPassword")
	} else {
		sc.AddClass("invalidPassword")
		sc.RemoveClass("validPassword")
	}
}

func (p *settingsPane) callbackHostChanged() {
	h, _ := p.hostCtrl.GetText()

	sc, _ := p.hostLabel.GetStyleContext()
	if !strings.ContainsAny(h, "!@#$%^&*()=+[]{}~`\\| ?,./<>") && h != "" {
		sc.AddClass("validPassword")
		sc.RemoveClass("invalidPassword")
	} else {
		sc.AddClass("invalidPassword")
		sc.RemoveClass("validPassword")
	}
}

func (p *settingsPane) callbackUserChanged() {
	u, _ := p.userCtrl.GetText()

	sc, _ := p.userLabel.GetStyleContext()
	if !strings.ContainsAny(u, "!@#$%^&*()=+[]{}~`\\| ?,./<>") && u != "" {
		sc.AddClass("validPassword")
		sc.RemoveClass("invalidPassword")
	} else {
		sc.AddClass("invalidPassword")
		sc.RemoveClass("validPassword")
	}
}

func (p *settingsPane) doUpdateValidation() {
	p.callbackPwChanged()
	p.callbackHostChanged()
	p.callbackUserChanged()
}

func (p *settingsPane) Show(settings *install.Settings, fullGrid *gtk.Grid) error {
	fullGrid.Attach(p.content, 0, 1, 1, 1)
	return nil
}

func (p *settingsPane) Hide(settings *install.Settings, fullGrid *gtk.Grid) error {
	currentPane, err := fullGrid.GetChildAt(0, 1)
	if err != nil {
		return fmt.Errorf("Failed to get current pane: %v", err)
	}
	fullGrid.Remove(currentPane)
	return nil
}

func (p *settingsPane) ShouldNext(settings *install.Settings, fullGrid *gtk.Grid) (bool, error) {
	p.doUpdateValidation()

	mainPw, _ := p.pwCtrl.GetText()
	confPw, _ := p.pwConfirm.GetText()
	if mainPw == "" || confPw != mainPw {
		return false, nil
	}
	h, _ := p.hostCtrl.GetText()
	if strings.ContainsAny(h, "!@#$%^&*()=+[]{}~`\\| ?,./<>") || h == "" {
		return false, nil
	}
	u, _ := p.userCtrl.GetText()
	if strings.ContainsAny(u, "!@#$%^&*()=+[]{}~`\\| ?,./<>") || u == "" {
		return false, nil
	}
	disk := p.disks[p.diskCtrl.GetActive()]
	tz := p.tzCtrl.GetActiveText()

	// Otherwise lets populate the settings struct!
	settings.Hostname = h
	settings.Username = u
	settings.Password = mainPw
	settings.Disk = disk
	settings.Timezone = tz
	settings.Scrub = p.scrubCheck.GetActive()
	settings.Autologin = p.loginCheck.GetActive()

	return true, nil
}

func readTzDir(p string) []string {
	var timezones []string
	fs, _ := ioutil.ReadDir(p)
	for _, f := range fs {
		if strings.HasPrefix(f.Name(), "posix") || strings.HasPrefix(f.Name(), "right") || !unicode.Is(unicode.Upper, rune(f.Name()[0])) {
			continue
		}

		if f.IsDir() {
			readTzDir(path.Join(p, f.Name()))
		} else {
			timezones = append(timezones, path.Join(p, f.Name()))
		}
	}

	return timezones
}

func timezones() []string {
	// Combo boxes are messed up rn.
	return []string{"America/Los_Angeles", "America/New_York", "Asia/Tokyo", "Australia/Sydney", "Europe/Amsterdam", "Europe/Berlin"}

	// for _, base := range []string{"/usr/share/zoneinfo", "/etc/zoneinfo"} {
	// 	if _, err := os.Stat(base); err != nil {
	// 		continue
	// 	}
	//
	// 	var timezones []string
	// 	fs, _ := ioutil.ReadDir(base)
	// 	for _, f := range fs {
	// 		if strings.HasPrefix(f.Name(), "posix") || strings.HasPrefix(f.Name(), "right") || strings.HasPrefix(f.Name(), "Etc") || !unicode.Is(unicode.Upper, rune(f.Name()[0])) {
	// 			continue
	// 		}
	//
	// 		if f.IsDir() {
	// 			timezones = append(timezones, readTzDir(path.Join(base, f.Name()))...)
	// 		} else {
	// 			timezones = append(timezones, path.Join(base, f.Name()))
	// 		}
	// 	}
	//
	// 	for i := range timezones {
	// 		timezones[i] = strings.TrimPrefix(timezones[i], base+"/")
	// 	}
	// 	return timezones
	// }
	//
	// return nil
}
