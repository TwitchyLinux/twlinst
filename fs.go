package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/twitchylinux/twlinst/z"
)

func getDevBlockSize(name string) (int, error) {
	d, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/block/%s/size", name))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.Trim(string(d), "\n\t\r "))
}

func getDiskInfo() ([]z.Disk, error) {
	stdout, err := exec.Command("lsblk", "-Jadp").Output()
	if err != nil {
		return nil, err
	}
	var blockDevs map[string][]struct {
		Name string `json:"name"`
		Size string `json:"size"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(stdout, &blockDevs); err != nil {
		fmt.Println(string(stdout))
		return nil, err
	}

	var out []z.Disk
	for _, blkDev := range blockDevs["blockdevices"] {
		if blkDev.Type == "disk" {
			diskInfo, err := z.GetUdevDiskInfo(blkDev.Name, true)
			if err != nil {
				return nil, err
			}
			diskInfo.NumBlocks, err = getDevBlockSize(diskInfo.Name)
			if err != nil {
				return nil, err
			}
			out = append(out, *diskInfo)
		}
	}

	return out, nil
}
