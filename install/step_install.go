package install

import (
	"os/exec"
	"path/filepath"
)

type InstallStep struct{}

func (s *InstallStep) Exec(updateChan chan Update, run *Run) error {
	mountBase := "/mnt"
	progressInfo(updateChan, "Commencing installation.\n")

	e := exec.Command("sudo", "nixos-install", "--no-root-passwd")
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

	// Copy any network connections the user configured during installation.
	e = exec.Command("sudo", "cp", "-rv", filepath.Join("/etc", "NetworkManager", "system-connections"), filepath.Join(mountBase, "etc", "NetworkManager"))
	out, err := e.CombinedOutput()
	if err != nil {
		progressInfo(updateChan, "  Output: %q\n", string(out))
		return err
	}

	return nil
}

func (s *InstallStep) Name() string {
	return "Install system"
}

func (s *InstallStep) Stage() string {
	return "copy"
}
