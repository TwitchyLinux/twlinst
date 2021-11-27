package install

import (
	"os"
	"os/exec"
	"time"
)

type ConfigureStep struct{}

func (s *ConfigureStep) Exec(updateChan chan Update, run *Run) error {
	if err := os.Mkdir("/tmp/install_mounts", 0755); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.Mkdir("/tmp/install_mounts/root", 0755); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.Mkdir("/tmp/install_mounts/boot", 0755); err != nil && !os.IsExist(err) {
		return err
	}
	time.Sleep(1 * time.Second)

	progressInfo(updateChan, "Mounting %s -> /tmp/install_mounts/boot\n", run.config.Disk.pathForPartition(1))
	cmd := exec.Command("sudo", "mount", run.config.Disk.pathForPartition(1), "/tmp/install_mounts/boot")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	progressInfo(updateChan, "  Output: %q\n", string(out))
	progressInfo(updateChan, "Mounted boot fs.\n")
	time.Sleep(3 * time.Second)

	progressInfo(updateChan, "\n  Mounting %s -> /tmp/install_mounts/root\n", "/dev/mapper/cryptroot")
	cmd = exec.Command("sudo", "mount", "/dev/mapper/cryptroot", "/tmp/install_mounts/root")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	progressInfo(updateChan, "  Output: %q\n", string(out))
	progressInfo(updateChan, "Mounted root fs.\n\n")
	time.Sleep(2 * time.Second)
	return nil
}

func (s *ConfigureStep) Name() string {
	return "Configure"
}

func (s *ConfigureStep) Stage() string {
	return "configure"
}
