package symptom

import (
	"bytes"
	"github.com/YangYongZhi/muxy/log"
	"github.com/YangYongZhi/muxy/muxy"
	"github.com/YangYongZhi/muxy/throttler"
	"github.com/mefellows/plugo/plugo"
	"io"
	l "log"
	"net"
	"os"
	"strconv"
	"strings"
)

// NetworkShaperSymptom allows you to modify the network speed on a host
// e.g. shape bandwidth to mobile, slower speeds
type NetworkShaperSymptom struct {
	Config           throttler.Config
	Device           string
	Latency          int      `default:"-1"`
	TargetBandwidth  int      `mapstructure:"target_bw" default:"-1"`
	DefaultBandwidth int      `mapstructure:"default_bw" default:"-1"`
	PacketLoss       float64  `mapstructure:"packet_loss"`
	TargetIps        []string `mapstructure:"target_ips"`
	TargetIps6       []string `mapstructure:"target_ips6"`
	TargetPorts      []string `mapstructure:"target_ports"`
	TargetProtos     []string `mapstructure:"target_protos" required:"true" default:"tcp,icmp,udp"`
	out              io.Writer
	err              io.Writer
}

func init() {
	plugo.PluginFactories.Register(func() (interface{}, error) {
		return &NetworkShaperSymptom{}, nil
	}, "network_shape")

}

// Setup sets up the plugin
func (s *NetworkShaperSymptom) Setup() {
	log.Debug("NetworkShaperSymptom - start setup")

	ports := parsePorts(strings.Join(s.TargetPorts, ","))
	log.Debug("NetworkShaperSymptom - ports %s", ports)
	targetIPv4, targetIPv6 := parseAddrs(strings.Join(append(s.TargetIps, s.TargetIps6...), ","))
	log.Debug("NetworkShaperSymptom - \tIPv4 %s \tIPv6 %s")

	s.Config = throttler.Config{
		Device:           s.Device,
		Latency:          s.Latency,
		TargetBandwidth:  s.TargetBandwidth,
		DefaultBandwidth: s.DefaultBandwidth,
		PacketLoss:       s.PacketLoss,
		TargetIps:        targetIPv4,
		TargetIps6:       targetIPv6,
		TargetPorts:      ports,
		TargetProtos:     s.TargetProtos,
		DryRun:           false,
	}

	executeThrottler(&s.Config)

	log.Debug("NetworkShaperSymptom has been setup")

}

var executeThrottler = func(config *throttler.Config) {
	log.Debug("NetworkShaperSymptom - start to execute throttler [network shape]")
	supressOutput(func() {
		log.Debug("thottler - start to run [network shape]")
		throttler.Run(config)
		log.Debug("NetworkShaperSymptom - end to run [network shape]")
	})
}

// HandleEvent is the hook into the event system
func (s NetworkShaperSymptom) HandleEvent(e muxy.ProxyEvent, ctx *muxy.Context) {
	/*
	 * Because the network shaper plugin is a layer 4 tamperer,
	 * it only need to configure the local firewall and network devices. don't require any proxy.
	 */
	switch e {
	case muxy.EventPreDispatch:
		s.Muck(ctx)
	}
}

// Muck is where the plugin can do any context-specific chaos
func (s *NetworkShaperSymptom) Muck(ctx *muxy.Context) {
	log.Debug("NetworkShaperSymptom - Nothing to muck")
}

// Teardown shuts down the plugin
func (s *NetworkShaperSymptom) Teardown() {
	log.Debug("NetworkShaperSymptom - Teardown()")
	s.Config.Stop = true
	executeThrottler(&s.Config)
}

// Supress output of function to keep logs clean
func supressOutput(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("supressOutput", err)
		}
	}()

	old := os.Stdout
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	//l.SetOutput(w)

	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		o := buf.String()
		log.Trace(o)
		outC <- o
	}()

	// log to TRACE

	// back to normal state
	w.Close()
	os.Stdout = old
	os.Stderr = oldErr
	l.SetOutput(old)
}

func parseLoss(loss string) float64 {
	loss = strings.TrimSpace(loss)
	val := loss
	if strings.Contains(loss, "%") {
		val = loss[:len(loss)-1]
	}
	l, err := strconv.ParseFloat(val, 64)
	if err != nil {
		fail("NetworkShaperSymptom - Incorrectly specified packet loss:", loss)
	}
	return l
}

func parseAddrs(addrs string) ([]string, []string) {
	adrs := strings.Split(strings.TrimSpace(addrs), ",")
	parsedIPv4 := []string{}
	parsedIPv6 := []string{}

	if addrs != "" {
		for _, adr := range adrs {
			ip := net.ParseIP(adr)
			if ip != nil {
				if ip.To4() != nil {
					parsedIPv4 = append(parsedIPv4, adr)
				} else {
					parsedIPv6 = append(parsedIPv6, adr)
				}
			} else { //Not a valid single IP, could it be a CIDR?
				parsedIP, net, err := net.ParseCIDR(adr)
				if err == nil {
					if parsedIP.To4() != nil {
						parsedIPv4 = append(parsedIPv4, net.String())
					} else {
						parsedIPv6 = append(parsedIPv6, net.String())
					}
				} else {
					fail("NetworkShaperSymptom - Incorrectly specified target IP or CIDR:", adr)
				}
			}
		}
	}

	return parsedIPv4, parsedIPv6
}

func parsePort(port string) int {
	prt, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil {
		return 0
	}

	return prt
}

func validPort(port string) bool {
	prt := parsePort(port)
	return prt > 0 && prt < 65536
}

func parsePorts(ports string) []string {
	prts := strings.Split(ports, ",")
	parsed := []string{}

	if ports != "" {
		for _, prt := range prts {
			if strings.Contains(prt, ":") {
				if validRange(prt) {
					parsed = append(parsed, prt)
				} else {
					fail("NetworkShaperSymptom - Incorrectly specified port range:", prt)
				}
			} else { //Isn't a range, check if just a single port
				if validPort(prt) {
					parsed = append(parsed, prt)
				} else {
					fail("NetworkShaperSymptom - Incorrectly specified port:", prt)
				}
			}
		}
	}

	return parsed
}

func validRange(ports string) bool {
	pr := strings.Split(ports, ":")

	if len(pr) == 2 {
		if !validPort(pr[0]) || !validPort(pr[1]) {
			return false
		}

		if portHigher(pr[0], pr[1]) {
			return false
		}
	} else {
		return false
	}

	return true
}

func portHigher(prt1, prt2 string) bool {
	p1 := parsePort(prt1)
	p2 := parsePort(prt2)

	return p1 > p2
}

func parseProtos(protos string) []string {
	ptcs := strings.Split(protos, ",")
	parsed := []string{}

	if protos != "" {
		for _, ptc := range ptcs {
			p := strings.ToLower(ptc)
			if p == "udp" ||
				p == "tcp" ||
				p == "icmp" {
				parsed = append(parsed, p)
			} else {
				fail("NetworkShaperSymptom - Incorrectly specified protocol:", p)
			}
		}
	}

	return parsed
}

var fail = func(reason string, i ...interface{}) {
	log.Fatalf(reason, i)
}
