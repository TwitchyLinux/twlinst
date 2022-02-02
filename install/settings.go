package install

import (
	"github.com/twitchylinux/twlinst/z"
)

// Settings contains all the configuration for the installer.
type Settings struct {
	Username string `json:"username"`
	Hostname string `json:"hostname"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`

	Disk      z.Disk `json:"-"`
	Scrub     bool   `json:"scrub_disk"`
	Autologin bool   `json:"autologin"`

	ConfigOnlyDisk string `json:"install_disk"`
}
