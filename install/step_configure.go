package install

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ConfigureStep struct{}

func (s *ConfigureStep) Exec(updateChan chan Update, run *Run) error {
	mountBase := "/mnt"

	time.Sleep(2 * time.Second)
	if err := s.setupMounts(updateChan, run, mountBase); err != nil {
		return err
	}
	if err := s.setupEtc(updateChan, run, mountBase); err != nil {
		return err
	}
	return nil
}

func (s *ConfigureStep) setupEtc(updateChan chan Update, run *Run, mountBase string) error {
	if _, err := exec.Command("sudo", "mkdir", "-p", filepath.Join(mountBase, "etc")).CombinedOutput(); err != nil {
		return err
	}
	if out, err := exec.Command("sudo", "chown", "nixos", filepath.Join(mountBase, "etc")).CombinedOutput(); err != nil {
		return fmt.Errorf("chmod etc (nixos): %s (%v)", strings.TrimSpace(string(out)), err)
	}

	progressInfo(updateChan, "\n  Staging configuration:\n")
	e := exec.Command("cp", "-arv", "/etc/nixos", filepath.Join(mountBase, "etc"))
	e.Stdout = &cmdInteractiveWriter{
		updateChan: updateChan,
		logPrefix:  "  ",
		IsProgress: true,
	}
	e.Stderr = e.Stdout
	if err := e.Start(); err != nil {
		return err
	}
	e.Wait()
	e = exec.Command("cp", "-arv", "/etc/twl-base", filepath.Join(mountBase, "etc"))
	e.Stdout = &cmdInteractiveWriter{
		updateChan: updateChan,
		logPrefix:  "  ",
		IsProgress: true,
	}
	e.Stderr = e.Stdout
	if err := e.Start(); err != nil {
		return err
	}
	e.Wait()

	if out, err := exec.Command("sudo", "root", "nixos", filepath.Join(mountBase, "etc")).CombinedOutput(); err != nil {
		return fmt.Errorf("chmod etc (root): %s (%v)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (s *ConfigureStep) setupMounts(updateChan chan Update, run *Run, mountBase string) error {
	cmd := exec.Command("sudo", "mkdir", "-p", mountBase)
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	progressInfo(updateChan, "\n  Mounting %s -> %s\n", "/dev/mapper/cryptroot", mountBase)
	cmd = exec.Command("sudo", "mount", "/dev/mapper/cryptroot", mountBase)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	progressInfo(updateChan, "  Output: %q\n", string(out))
	progressInfo(updateChan, "Mounted root fs.\n\n")
	time.Sleep(3 * time.Second)

	cmd = exec.Command("sudo", "mkdir", "-p", filepath.Join(mountBase, "boot"))
	out, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	progressInfo(updateChan, "Mounting %s -> %s\n", run.config.Disk.pathForPartition(1), filepath.Join(mountBase, "boot"))
	cmd = exec.Command("sudo", "mount", run.config.Disk.pathForPartition(1), filepath.Join(mountBase, "boot"))
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
