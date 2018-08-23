package throttler

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/YangYongZhi/muxy/log"
	"os"
	"os/exec"
	"runtime"
)

const (
	linux           = "linux"
	darwin          = "darwin"
	freebsd         = "freebsd"
	windows         = "windows"
	checkOSXVersion = "sw_vers -productVersion"
	ipfw            = "ipfw"
	pfctl           = "pfctl"
)

// Config specifies options for configuring packet filter rules.
type Config struct {
	Device           string
	Stop             bool
	Latency          int
	TargetBandwidth  int
	DefaultBandwidth int
	PacketLoss       float64
	TargetIps        []string
	TargetIps6       []string
	TargetPorts      []string
	TargetProtos     []string
	DryRun           bool
}

type throttler interface {
	setup(*Config) error
	teardown(*Config) error
	exists() bool
	check() string
}

type commander interface {
	execute(string) error
	executeGetLines(string) ([]string, error)
	commandExists(string) bool
}

type dryRunCommander struct{}

type shellCommander struct{}

var dry bool

func setup(t throttler, cfg *Config) {
	if t.exists() {
		log.Debug("It looks like the packet rules are already setup")
		os.Exit(1)
	}

	if err := t.setup(cfg); err != nil {
		log.Debug("I couldn't setup the packet rules:" + err.Error())
		os.Exit(1)
	}

	log.Debug("Packet rules setup...")
	log.Debug("Run `%s` to double check\n", t.check())
	log.Debug("Run `%s --device %s --stop` to reset\n", os.Args[0], cfg.Device)
}

func teardown(t throttler, cfg *Config) {
	if !t.exists() {
		log.Debug("It looks like the packet rules aren't setup")
		os.Exit(1)
	}

	if err := t.teardown(cfg); err != nil {
		log.Debug("Failed to stop packet controls")
		os.Exit(1)
	}

	log.Debug("Packet rules stopped...")
	log.Debug("Run `%s` to double check\n", t.check())
	log.Debug("Run `%s` to start\n", os.Args[0])
}

// Run executes the packet filter operation, either setting it up or tearing
// it down.
func Run(cfg *Config) {
	log.Debug("Start to run a throttler")
	dry = cfg.DryRun
	var t throttler
	var c commander

	if cfg.DryRun {
		c = &dryRunCommander{}
	} else {
		c = &shellCommander{}
	}

	log.Debug("Your OS \t" + log.Colorize(log.YELLOW, runtime.GOOS))
	switch runtime.GOOS {
	case freebsd:
		if cfg.Device == "" {
			log.Debug("Device not specified, unable to default to eth0 on FreeBSD.")
			os.Exit(1)
		}

		t = &ipfwThrottler{c}
	case darwin:
		// Avoid OS version pinning and choose based on what's available
		if c.commandExists(pfctl) {
			t = &pfctlThrottler{c}
		} else if c.commandExists(ipfw) {
			t = &ipfwThrottler{c}
		} else {
			log.Debug("Could not determine an appropriate firewall tool for OSX (tried pfctl, ipfw), exiting")
			os.Exit(1)
		}

		if cfg.Device == "" {
			cfg.Device = "eth0"
		}

	case linux:
		if cfg.Device == "" {
			cfg.Device = "eth0"
		}
		t = &tcThrottler{c}
	default:
		log.Debug("I don't support your OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	log.Debug("Your device \t" + log.Colorize(log.YELLOW, cfg.Device))

	if !cfg.Stop {
		log.Debug("Start to setup throttler")
		setup(t, cfg)
		log.Debug("Finish to setup")
	} else {
		log.Debug("Start to teardown throttler")
		teardown(t, cfg)
		log.Debug("Finish to teardown")
	}
}

func (c *dryRunCommander) execute(cmd string) error {
	log.Debug(log.Colorize(log.GREEN, cmd))
	return nil
}

func (c *dryRunCommander) executeGetLines(cmd string) ([]string, error) {
	log.Debug(log.Colorize(log.GREEN, cmd))
	return []string{}, nil
}

func (c *dryRunCommander) commandExists(cmd string) bool {
	return true
}

func (c *shellCommander) execute(cmd string) error {
	log.Debug(log.Colorize(log.GREEN, cmd))
	return exec.Command("/bin/sh", "-c", cmd).Run()
}

func (c *shellCommander) executeGetLines(cmd string) ([]string, error) {
	lines := []string{}

	log.Debug("executeGetLines: " + log.Colorize(log.GREEN, cmd))
	child := exec.Command("/bin/sh", "-c", cmd)

	out, err := child.StdoutPipe()
	if err != nil {
		return []string{}, err
	}

	err = child.Start()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []string{}, errors.New(fmt.Sprint("Error reading standard input:", err))
	}

	err = child.Wait()
	if err != nil {
		return []string{}, err
	}

	return lines, nil
}

func (c *shellCommander) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
