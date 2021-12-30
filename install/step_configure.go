package install

import (
	"os/exec"
	"time"
)

type ConfigureStep struct{}

func (s *ConfigureStep) Exec(updateChan chan Update, run *Run) error {
	time.Sleep(2 * time.Second)
	cmd := exec.Command("sudo", "mkdir", "-p", "/mnt")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	progressInfo(updateChan, "\n  Mounting %s -> /mnt\n", "/dev/mapper/cryptroot")
	cmd = exec.Command("sudo", "mount", "/dev/mapper/cryptroot", "/mnt")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	progressInfo(updateChan, "  Output: %q\n", string(out))
	progressInfo(updateChan, "Mounted root fs.\n\n")
	time.Sleep(3 * time.Second)

	cmd = exec.Command("sudo", "mkdir", "-p", "/mnt/boot")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	progressInfo(updateChan, "Mounting %s -> /mnt/boot\n", run.config.Disk.pathForPartition(1))
	cmd = exec.Command("sudo", "mount", run.config.Disk.pathForPartition(1), "/mnt/boot")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	progressInfo(updateChan, "  Output: %q\n", string(out))
	progressInfo(updateChan, "Mounted boot fs.\n")
	time.Sleep(2 * time.Second)

	return nil
}

func (s *ConfigureStep) Name() string {
	return "Configure"
}

func (s *ConfigureStep) Stage() string {
	return "configure"
}
