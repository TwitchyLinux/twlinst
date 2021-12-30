package install

import "os/exec"

type InstallStep struct{}

func (s *InstallStep) Exec(updateChan chan Update, run *Run) error {
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

	return nil
}

func (s *InstallStep) Name() string {
	return "Install system"
}

func (s *InstallStep) Stage() string {
	return "copy"
}
