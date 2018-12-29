package main

import (
	"os/exec"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"strings"
	"os"
	"time"
)
var log = elog.Log

func main() {
	i := 0
	for {
		i++
		log.Debug("start test program:", i)
		out, err := runCmd("./run.sh")
		errors.CheckErrorPanic(err)
		time.Sleep(time.Second*10)
		out, err = runCmd("find ecoball_log/ -name ecoball.log | xargs grep 123456789")
		log.Notice(out)
		out, err = runCmd("find ecoball_log/ -name ecoball.log | xargs grep ERROR")
		errors.CheckErrorPanic(err)
		log.Info(out)
		if strings.Contains(out, "connection reset") {
			os.Exit(0)
		}
	}
}


func runCmd(shell string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", shell)
	out, err := cmd.Output()
	if err != nil {
		log.Warn("exec ", cmd.Args, "failed, ", err.Error())
		log.Warn(string(out))
		return "", err
	}
	return string(out), err
}