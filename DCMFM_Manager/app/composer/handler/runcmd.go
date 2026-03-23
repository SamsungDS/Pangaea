package handler

import (
	// Built-in PKG
	"os/exec"
	"runtime"

	// External PKG
	"github.com/sirupsen/logrus"
)

func RunCMD(command string) string {
	var shell, flag string

	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/c"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}

	logrus.Debugf("command: %s", command)

	out, err := exec.Command(shell, flag, command).Output()
	logrus.Debugf("err: %s", err)
	if err != nil {
		panic(err)
	}

	return string(out)
}
