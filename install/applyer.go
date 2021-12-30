package install

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	fileMarkerExp  = regexp.MustCompile("(?m)\\s*# File-marker: (.+)$")
	lineMarkerExp  = regexp.MustCompile("(?m)\\s*# Line-marker: (.+)$")
	startMarkerExp = regexp.MustCompile("\\s*# Start-marker: (.+)$")
	endMarkerExp   = regexp.MustCompile("\\s*# End-marker: (.+)$")
)

// Applyer mutates the nix installer config to setup an installation.
type Applyer struct {
	TwlBasePath string
	Run         *Run
}

func (a *Applyer) applyFile(srcPath string, size int64) error {
	if !strings.HasSuffix(srcPath, ".nix") {
		return nil
	}
	contents, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return err
	}

	if fileMarker := fileMarkerExp.FindSubmatch(contents); len(fileMarker) > 0 {
		return a.handleFileMarker(srcPath, string(fileMarker[1]))
	}

	return a.handleInlineMarkers(srcPath, contents, size)
}

func (a *Applyer) handleInlineMarkers(srcPath string, contents []byte, size int64) error {
	var (
		out        bytes.Buffer
		hasChanges bool
	)
	out.Grow(int(size))

	var (
		scanner   = bufio.NewScanner(bytes.NewReader(contents))
		lastStart = ""
	)
	for scanner.Scan() {
		switch {
		case lastStart == "": // Not within any section
			if m := startMarkerExp.FindSubmatch(scanner.Bytes()); len(m) > 0 {
				lastStart = string(m[1])

				if string(m[1]) == "Trim on install" {
					continue
				}
			} else {
				// Check for any line markers
				if m := lineMarkerExp.FindSubmatch(scanner.Bytes()); len(m) > 0 {
					switch string(m[1]) {
					case "Trim on install":
						hasChanges = true
						continue
					default:
						return fmt.Errorf("unknown line marker: %q", string(m[1]))
					}
				}
			}

		case lastStart != "": // within a section
			if m := endMarkerExp.FindSubmatch(scanner.Bytes()); len(m) > 0 {
				if string(m[1]) != lastStart {
					return fmt.Errorf("unmatched start/end: %s != %s", string(m[1]), lastStart)
				}
				lastStart = ""
				hasChanges = true

				if string(m[1]) == "Trim on install" {
					continue
				}
			} else {
				// Handle logic for specific section
				switch lastStart {
				case "Trim on install":
					continue
				}
			}
		}

		out.Write(scanner.Bytes())
		out.WriteRune('\n')
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning: %v", err)
	}

	if hasChanges {
		return ioutil.WriteFile(srcPath, out.Bytes(), 0644)
	}
	return nil
}

func (a *Applyer) handleFileMarker(srcPath, marker string) error {
	switch marker {
	case "Trim on install":
		return os.Remove(srcPath)
	default:
		return fmt.Errorf("unknown file marker: %q", marker)
	}
}

func (a *Applyer) Exec() error {
	entries, err := ioutil.ReadDir(a.TwlBasePath)
	if err != nil {
		return err
	}

	for _, f := range entries {
		srcPath := filepath.Join(a.TwlBasePath, f.Name())
		if err := a.applyFile(srcPath, f.Size()); err != nil {
			return fmt.Errorf("applying %s: %v", srcPath, err)
		}
	}

	return nil
}
