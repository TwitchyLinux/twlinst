package install

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

const bootPartSizeMB = 256

// ByteCountDecimal pretty-formats the number of bytes.
func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

type PartitionStep struct{}

func (s *PartitionStep) Exec(updateChan chan Update, run *Run) error {
	progressInfo(updateChan, "Partitioning %q\n", run.config.Disk.Path)
	progressInfo(updateChan, "Device has a capacity of %s\n", ByteCountDecimal(int64(run.config.Disk.NumBlocks*512)))
	progressInfo(updateChan, "\n  New partition table:\n")
	progressInfo(updateChan, "    [FAT32]  Boot partition (%s)\n", ByteCountDecimal(bootPartSizeMB))
	progressInfo(updateChan, "    [LUKS2]  Encrypted root partition\n")

	cmd := exec.Command("sudo", "parted", "--script", run.config.Disk.Path, "mklabel", "gpt",
		"mkpart", "'EFI system partition'", "fat32", "1MiB", strconv.Itoa(bootPartSizeMB)+"MiB",
		"mkpart", "TWL", strconv.Itoa(bootPartSizeMB)+"MiB", "100%",
		"set", "1", "esp", "on")

	progressInfo(updateChan, "\n  Parted invocation: %v\n", cmd.Args)

	out, err := cmd.CombinedOutput()
	progressInfo(updateChan, "  Output: %q\n", string(out))
	if err != nil {
		return err
	}
	time.Sleep(time.Second)

	cmd = exec.Command("sudo", "partprobe", run.config.Disk.Path)
	progressInfo(updateChan, "\n  Probing: %v\n", run.config.Disk.Path)
	out, err = cmd.CombinedOutput()
	progressInfo(updateChan, "  Output: %q\n", string(out))
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)

	cmd = exec.Command("sudo", "mkfs.fat", "-F32", "-n", "SYSTEM-EFI", run.config.Disk.pathForPartition(1))
	progressInfo(updateChan, "\n  Creating fat32 EFI filesystem on %v\n", run.config.Disk.pathForPartition(1))
	out, err = cmd.CombinedOutput()
	progressInfo(updateChan, "  Output: %q\n", string(out))
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	cmd = exec.Command("sudo", "cryptsetup", "luksFormat", "--type", "luks2", run.config.Disk.pathForPartition(2), "--key-file", "-",
		"--hash", "sha256", "--cipher", "aes-xts-plain64", "--key-size", "512", "--iter-time", "2600", "--use-random")
	progressInfo(updateChan, "\n  Creating encrypted filesystem on %v\n", run.config.Disk.pathForPartition(2))
	progressInfo(updateChan, "  Invocation: %v\n", cmd.Args)
	cmd.Stdin = bytes.NewReader([]byte(run.config.Password))
	out, err = cmd.CombinedOutput()
	progressInfo(updateChan, "  Output: %q\n", string(out))
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	progressInfo(updateChan, "\n  Unlocking root filesystem\n")
	cmd = exec.Command("sudo", "cryptsetup", "luksOpen", "--key-file", "-", run.config.Disk.pathForPartition(2), "cryptroot")
	progressInfo(updateChan, "  Invocation: %v\n", cmd.Args)
	cmd.Stdin = bytes.NewReader([]byte(run.config.Password))
	out, err = cmd.CombinedOutput()
	progressInfo(updateChan, "  Output: %q\n", string(out))
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	if run.config.Scrub {
		if err := s.scrubEncrypted(updateChan, run); err != nil {
			return err
		}
	}

	cmd = exec.Command("sudo", "mkfs.ext4", "-qF", "/dev/mapper/cryptroot")
	progressInfo(updateChan, "\n  Creating ext4 filesystem on %v\n", "/dev/mapper/cryptroot")
	out, err = cmd.CombinedOutput()
	progressInfo(updateChan, "  Output: %q\n", string(out))
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

func (s *PartitionStep) scrubEncrypted(updateChan chan Update, run *Run) error {
	progressInfo(updateChan, "\n  Scrubbing encrypted partition:\n")
	e := exec.Command("sudo", "dd", "if=/dev/zero", "of=/dev/mapper/cryptroot", "bs=1M", "status=progress")
	progressInfo(updateChan, "  Invocation: %v\n", e.Args)

	e.Stdout = &cmdInteractiveWriter{
		updateChan: updateChan,
		logPrefix:  "  ",
		IsProgress: true,
	}
	e.Stderr = e.Stdout

	if err := e.Start(); err != nil {
		return err
	}

	e.Wait() // Will error when we exhaust the space on the device (intended).
	return nil
}

func (s *PartitionStep) Name() string {
	return "Format disk"
}

func (s *PartitionStep) Stage() string {
	return "format"
}
