package install

import (
	"github.com/twitchylinux/twlinst/z"
)

// Settings contains all the configuration for the installer.
type Settings struct {
	Username string
	Hostname string
	Password string
	Timezone string

	Disk      z.Disk
	Scrub     bool
	Autologin bool
}
