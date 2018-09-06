package stressng

import "os/exec"

var RunningStressor map[int]exec.Cmd

func init() {
	RunningStressor = make(map[int]exec.Cmd)
}
