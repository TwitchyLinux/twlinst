package install

import (
	"fmt"
	"os/exec"
	"strings"
)

type MsgLevel uint8

// Valid MsgLevel values.
const (
	MsgInfo MsgLevel = iota
	MsgWarn
	MsgErr
	MsgCmd
)

// Update represents an update for the UI that signals a log message
// or some kind of progress.
type Update struct {
	Step string

	Msg   string
	Level MsgLevel

	TrimLastLine bool
	Complete bool
}

func progressInfo(updateChan chan Update, fmtStr string, args ...interface{}) {
	updateChan <- Update{
		Msg: fmt.Sprintf("  "+fmtStr, args...),
	}
}

type cmdInteractiveWriter struct {
	updateChan chan Update
	logPrefix  string
	IsErr      bool
	IsProgress bool
}

func (c *cmdInteractiveWriter) Write(in []byte) (int, error) {
	for _, line := range strings.Split(string(in), "\n") {
		line = strings.Trim(line, " \r\n")
		if len(line) < 2 {
			continue
		}

		out := Update{
			Msg:          fmt.Sprintf("  %s %s\n", c.logPrefix, line),
			TrimLastLine: c.IsProgress,
		}
		if c.IsErr {
			out.Level = MsgWarn
		}
		c.updateChan <- out
	}
	return len(in), nil
}

func runCmdInteractive(updateChan chan Update, logPrefix, cmd string, args ...string) error {
	e := exec.Command(cmd, args...)
	updateChan <- Update{Msg: fmt.Sprintf("  %s %s %s\n", logPrefix, cmd, args)}

	e.Stdout = &cmdInteractiveWriter{
		updateChan: updateChan,
		logPrefix:  logPrefix,
	}
	e.Stderr = &cmdInteractiveWriter{
		updateChan: updateChan,
		logPrefix:  logPrefix,
		IsErr:      true,
	}

	return e.Run()
}
