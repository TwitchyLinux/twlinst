package install

import (
	"fmt"
	"os"
	"strings"
)

// Settings contains all the configuration for the installer.
type Settings struct {
	Username string
	Hostname string
	Password string
	Timezone string

	Disk  Disk
	Scrub bool
}

// Disk describes a block device or partition on the system.
type Disk struct {
	Name, Path string

	Model    string
	Serial   string
	Bus, Rev string
	Symlinks []string

	NumBlocks int

	Major, Minor int
	PartN        int

	PartTabType string
	PartUUID    string
	FsUUID      string
	Partitions  []*Disk
	FS, Label   string
}

func (d *Disk) pathForPartition(partNum int) string {
	if _, err := os.Stat(fmt.Sprintf("%s%d", d.Path, partNum)); err == nil {
		return fmt.Sprintf("%s%d", d.Path, partNum)
	}
	if _, err := os.Stat(fmt.Sprintf("%sp%d", d.Path, partNum)); err == nil {
		return fmt.Sprintf("%sp%d", d.Path, partNum)
	}

	// fallback
	if strings.Contains(d.Path, "/sd") {
		return d.Path + fmt.Sprint(partNum)
	}
	return d.Path + "p" + fmt.Sprint(partNum)
}
