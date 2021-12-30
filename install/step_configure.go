package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/twitchylinux/twlinst/z"
)

const fsTmpl = `
{config, pkgs, boot, lib, ...}:
	{
		boot.initrd.luks.devices = {
			"cryptroot" = {
				device = "/dev/disk/by-uuid/{{.LuksUUID}}";
			};
		};

		fileSystems = {
			"/" = {
				device = "/dev/disk/by-uuid/{{.Ext4UUID}}";
				fsType = "ext4";
			};
			"/boot" = {
				device = "/dev/disk/by-uuid/{{.BootUUID}}";
				fsType = "vfat";
			};
		};
	}
`

const nixCfgTmpl = `
{...}:
{
	imports = [
		/etc/twl-base
		./filesystems.nix
	];

	users.users.{{.Username}} = {
		isNormalUser = true;
		extraGroups = [ "wheel" "networkmanager" "video" ];
		hashedPassword = "{{.PasswordHash}}";
	};

	time.timeZone = "{{.Timezone}}";
	networking.hostName = "{{.Hostname}}";

	{{if .Autologin -}}
	# Automatically login on startup.
	services.getty.autologinUser = "{{.Username}}";
	system.activationScripts.etc = lib.stringAfter [ "users" "groups" ]
		''
			mkdir -pv /home/{{.Username}}/.config/sway
			ln -s /etc/twl-base/resources/sway.config /home/{{.Username}}/.config/sway/config || true

			if [ ! -f /home/{{.Username}}/.bashrc ]; then
			echo 'if [[ $(tty) == "/dev/tty1" ]]; then' >> /home/{{.Username}}/.bash_login
			echo '  sleep 2 && startsway' >> /home/{{.Username}}/.bash_login
			echo 'fi' >> /home/{{.Username}}/.bash_login
			fi
		'';
	{{- end}}
}
`

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
	if err := s.setupFilesystemConf(updateChan, run, mountBase); err != nil {
		return err
	}
	if err := s.setupNixConf(updateChan, run, mountBase); err != nil {
		return err
	}

	if out, err := exec.Command("sudo", "chown", "-R", "root", filepath.Join(mountBase, "etc")).CombinedOutput(); err != nil {
		return fmt.Errorf("chmod etc (root): %s (%v)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (s *ConfigureStep) setupNixConf(updateChan chan Update, run *Run, mountBase string) error {
	t, err := template.New("configuration.nix").Parse(nixCfgTmpl)
	if err != nil {
		return fmt.Errorf("template parse: %v", err)
	}

	f, err := os.OpenFile(filepath.Join(mountBase, "etc", "nixos", "configuration.nix"), os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	mkpwd := exec.Command("mkpasswd", "-s", "-m", "sha-512")
	mkpwd.Stdin = strings.NewReader(run.config.Password)
	pwHash, err := mkpwd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mkpasswd: %v", err)
	}

	if err := t.Execute(f, map[string]interface{}{
		"Username": run.config.Username,
		"Timezone": run.config.Timezone,
		"PasswordHash": strings.TrimSpace(string(pwHash)),
		"Hostname": run.config.Hostname,
		"Autologin": run.config.Autologin,
	}); err != nil {
		return fmt.Errorf("writing config: %v", err)
	}

	return f.Close()
}

func (s *ConfigureStep) setupFilesystemConf(updateChan chan Update, run *Run, mountBase string) error {
	t, err := template.New("filesystems.nix").Parse(fsTmpl)
	if err != nil {
		return fmt.Errorf("template parse: %v", err)
	}

	f, err := os.OpenFile(filepath.Join(mountBase, "etc", "nixos", "filesystems.nix"), os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	bootInfo, err := z.GetUdevDiskInfo(run.config.Disk.PathForPartition(1), false)
	if err != nil {
		return err
	}
	mainInfo, err := z.GetUdevDiskInfo(run.config.Disk.PathForPartition(2), false)
	if err != nil {
		return err
	}
	cryptInfo, err := z.GetUdevDiskInfo("/dev/mapper/cryptroot", false)
	if err != nil {
		return err
	}

	fmt.Printf("boot info: %+v\n\nmain info: %+v\n\n crypt info: %+v\n\n", bootInfo, mainInfo, cryptInfo)

	if err := t.Execute(f, map[string]interface{}{
		"BootUUID": bootInfo.FsUUID,
		"LuksUUID": mainInfo.FsUUID,
		"Ext4UUID": cryptInfo.FsUUID,
	}); err != nil {
		return fmt.Errorf("writing filesystems: %v", err)
	}

	return f.Close()
}

func (s *ConfigureStep) setupEtc(updateChan chan Update, run *Run, mountBase string) error {
	if _, err := exec.Command("sudo", "mkdir", "-p", filepath.Join(mountBase, "etc")).CombinedOutput(); err != nil {
		return err
	}
	if out, err := exec.Command("sudo", "chown", "nixos", filepath.Join(mountBase, "etc")).CombinedOutput(); err != nil {
		return fmt.Errorf("chmod etc (nixos): %s (%v)", strings.TrimSpace(string(out)), err)
	}

	progressInfo(updateChan, "\n  Staging configuration:\n")
	e := exec.Command("cp", "-ar", "/etc/nixos", filepath.Join(mountBase, "etc"))
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
	e = exec.Command("cp", "-ar", "/etc/twl-base", filepath.Join(mountBase, "etc"))
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

	a := Applyer{
		Run:         run,
		TwlBasePath: filepath.Join(mountBase, "etc", "twl-base"),
	}
	if err := a.Exec(); err != nil {
		return err
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

	progressInfo(updateChan, "Mounting %s -> %s\n", run.config.Disk.PathForPartition(1), filepath.Join(mountBase, "boot"))
	cmd = exec.Command("sudo", "mount", run.config.Disk.PathForPartition(1), filepath.Join(mountBase, "boot"))
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
