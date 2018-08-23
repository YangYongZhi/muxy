package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/YangYongZhi/muxy/command"
	_ "github.com/YangYongZhi/muxy/middleware"
	_ "github.com/YangYongZhi/muxy/protocol"
	_ "github.com/YangYongZhi/muxy/symptom"
	"github.com/mitchellh/cli"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	rand.Seed(time.Now().Unix())
	cli := cli.NewCLI(strings.ToLower(ApplicationName), Version)
	cli.Args = os.Args[1:]
	cli.Commands = command.Commands

	exitStatus, err := cli.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	return exitStatus
}
