package http

import (
	"encoding/json"
	"fmt"
	"github.com/YangYongZhi/muxy/log"
	"github.com/YangYongZhi/muxy/middleware"
	"github.com/YangYongZhi/muxy/symptom"
	"github.com/YangYongZhi/muxy/throttler"
	"io/ioutil"
	"net/http"
	"os/exec"
	"reflect"
	"time"
)

/**

 */
func middlewaresHanlder(w http.ResponseWriter, r *http.Request) {
	middlewares := Muxy.MiddleWares()
	if len(middlewares) > 0 {
		log.Debug("Muxy is running, middleware count : %d\n", len(middlewares))
	}

	//middlewareJson, _ := json.Marshal(middlewares)
	//middlewareJsonString := string(middlewareJson)

	middlewareJson, _ := json.MarshalIndent(middlewares, "", "    ")
	middlewareJsonString := string(middlewareJson)
	log.Debug("middlewares are :\n%s", middlewareJsonString)

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, middlewareJsonString)
}

/**
 *
 */
func networkshapeHanlder(w http.ResponseWriter, r *http.Request) {
	// List all iptables
	iptCmdStr := fmt.Sprintf(throttler.IptList, throttler.Ip4Tables)
	iptCmd := exec.Command("/bin/bash", "-c", iptCmdStr)
	iptOut, err := iptCmd.Output()
	log.Debug("Executed command : %s", log.Colorize(log.GREEN, iptCmdStr))
	if err != nil {
		log.Error("Error: %s", err.Error())
	}
	fmt.Fprintf(w, "### %s ###:\n%s", iptCmdStr, string(iptOut))

	// List all iptables 6
	ipt6CmdStr := fmt.Sprintf(throttler.IptList, throttler.Ip6Tables)
	ipt6Cmd := exec.Command("/bin/bash", "-c", ipt6CmdStr)
	ipt6Out, err := ipt6Cmd.Output()
	log.Debug("Executed command : %s", log.Colorize(log.GREEN, ipt6CmdStr))
	if err != nil {
		log.Error("Error: %s", err.Error())
	}
	fmt.Fprintf(w, "### %s ###:\n%s", ipt6CmdStr, string(ipt6Out))

	// Show tc qdisc
	tcListCmd := exec.Command("/bin/bash", "-c", throttler.TcList)
	tcOut, err := tcListCmd.Output()
	log.Debug("Executed command : %s", log.Colorize(log.GREEN, throttler.TcList))
	if err != nil {
		log.Error("Error: %s", err.Error())
	}
	fmt.Fprintf(w, "### %s ###:\n%s", throttler.TcList, string(tcOut))
}

/**
 *
 */
func clearHanlder(w http.ResponseWriter, r *http.Request) {
	log.Info("Start forcing to clear all tc qdisc")
	for _, m := range Muxy.MiddleWares() {
		log.Debug("%s", reflect.TypeOf(m))

		switch v := m.(type) {
		case *middleware.LoggerMiddleware:
			log.Debug("Not support %v yet.", v)
		case *symptom.HTTPDelaySymptom:
			log.Debug("Not support %v yet.", v)
		case *symptom.NetworkShaperSymptom:
			log.Debug("Not support %v yet.", v)
		}

	}

	log.Info("Forcing clear all tc qdisc successfully")
	fmt.Fprint(w, "Forcing to clear all tc qdisc successfully.")
}

/**
 *
 */
func resetHanlder(w http.ResponseWriter, r *http.Request) {
	//e.g. body = {"Device":"ens33","Latency": 2000, "TargetBandWidth":20,"PacketLoss":70,"TargetPorts": ["5001","10090"], "TargetProtos":["tcp","icmp"]}
	log.Info("Reset the middlewares, please wait.")

	//decode := json.NewDecoder(r.Body)

	body, _ := ioutil.ReadAll(r.Body)
	//    r.Body.Close()
	body_str := string(body)
	log.Debug("Body : %s", body_str)

	var config throttler.Config

	if err := json.Unmarshal(body, &config); err == nil {
		log.Debug("latency :%d", config.Latency)
		fmt.Fprintf(w, "%s\n", body_str)
	} else {
		log.Error("Reset a network shape has an error", err)
		fmt.Fprint(w, err)
		return
	}

	log.Debug("The parameters for reset the work shape on current muxy: %s", body_str)

	for _, m := range Muxy.MiddleWares() {
		log.Debug("%s", reflect.TypeOf(m))

		switch v := m.(type) {
		case *middleware.LoggerMiddleware:
			log.Debug("Not support %v yet.", v)
		case *symptom.HTTPDelaySymptom:
			log.Debug("Not support %v yet.", v)
		case *symptom.NetworkShaperSymptom:
			v.Teardown()

			// We should sleep this current goroutine in order to make the ip tables change less frequently.
			log.Debug("We should sleep this current goroutine in order to make the ip tables change less frequently.")
			time.Sleep(sleepTimeBetweenDisableAndEnable)

			v.Device = config.Device
			v.Latency = config.Latency
			v.LatencyJitter = config.LatencyJitter
			v.LatencyCorrelation = config.LatencyCorrelation
			v.LatencyDistribution = config.LatencyDistribution
			v.LatencyReorder = config.LatencyReorder
			v.LatencyDuplicate = config.LatencyDuplicate
			v.LatencyCorrupt = config.LatencyCorrupt
			v.PacketLoss = config.PacketLoss
			v.TargetBandwidth = config.TargetBandwidth
			v.TargetPorts = config.TargetPorts
			v.TargetProtos = config.TargetProtos
			if len(config.TargetIps) > 0 {
				v.TargetIps = config.TargetIps
			}
			if len(config.TargetIps6) > 0 {
				v.TargetIps6 = config.TargetIps6
			}

			v.Config.Device = config.Device
			v.Config.Latency = config.Latency
			v.Config.LatencyJitter = config.LatencyJitter
			v.Config.LatencyCorrelation = config.LatencyCorrelation
			v.Config.LatencyDistribution = config.LatencyDistribution
			v.Config.LatencyReorder = config.LatencyReorder
			v.Config.LatencyDuplicate = config.LatencyDuplicate
			v.Config.LatencyCorrupt = config.LatencyCorrupt
			v.Config.PacketLoss = config.PacketLoss
			v.Config.TargetBandwidth = config.TargetBandwidth
			v.Config.TargetPorts = config.TargetPorts
			v.Config.TargetProtos = config.TargetProtos
			if len(config.TargetIps) > 0 {
				v.Config.TargetIps = config.TargetIps
			}
			if len(config.TargetIps6) > 0 {
				v.Config.TargetIps6 = config.TargetIps6
			}

			v.Setup()
		}

	}

	log.Info("Create some middleware for current Muxy successfully")
	fmt.Fprint(w, "Reset the network shape successfully.")
}

/**
 *
 */
func enableHanlder(w http.ResponseWriter, r *http.Request) {
	log.Info("Enable the network shape of the current Muxy, please wait.")

	fmt.Fprint(w, "Enable\n")
	for _, m := range Muxy.MiddleWares() {
		log.Info("Enable type %s", reflect.TypeOf(m))
		m.Setup()
		fmt.Fprintf(w, "%s.\n", reflect.TypeOf(m))
	}

	log.Info("Enable Muxy successfully")
	fmt.Fprint(w, "Enable Muxy successfully")
}

/**
 *
 */
func disableHanlder(w http.ResponseWriter, r *http.Request) {
	log.Info("Disable all rules with current Muxy, please wait.")

	fmt.Fprint(w, "Disable\n")
	for _, m := range Muxy.MiddleWares() {
		log.Info("Disable type %s", reflect.TypeOf(m))
		m.Teardown()
		fmt.Fprintf(w, "%s.\n", reflect.TypeOf(m))
	}

	log.Info("Disable the network shape of the current Muxy successfully")
	fmt.Fprint(w, "Disable Muxy successfully")
}
