package http

import (
	"encoding/json"
	"fmt"
	"github.com/YangYongZhi/muxy/log"
	"github.com/YangYongZhi/muxy/middleware"
	m "github.com/YangYongZhi/muxy/run"
	"github.com/YangYongZhi/muxy/stressng"
	"github.com/YangYongZhi/muxy/symptom"
	"github.com/YangYongZhi/muxy/throttler"
	"io/ioutil"
	"net/http"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	port                             = ":13003"
	sleepTimeBetweenDisableAndEnable = 5 * time.Second
)

var Muxy *m.Muxy

type MuxyApiServer struct {
	Name string
	//server *http.Server
}

func New(name string) *MuxyApiServer {
	return &MuxyApiServer{Name: name}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Request method : [%s]", log.Colorize(log.YELLOW, r.URL.Path[1:]))
	method := r.URL.Path[1:]
	//fmt.Fprintf(w, "Hi, your resource is %s\n", method)

	switch method {
	case "middlewares":
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
	case "networkshape":
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
	case "clear":
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
	case "reset":
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
	case "enable":
		log.Info("Enable the network shape of the current Muxy, please wait.")

		fmt.Fprint(w, "Enable\n")
		for _, m := range Muxy.MiddleWares() {
			log.Info("Enable type %s", reflect.TypeOf(m))
			m.Setup()
			fmt.Fprintf(w, "%s.\n", reflect.TypeOf(m))
		}

		log.Info("Enable Muxy successfully")
		fmt.Fprint(w, "Enable Muxy successfully")
	case "disable":
		log.Info("Disable all rules with current Muxy, please wait.")

		fmt.Fprint(w, "Disable\n")
		for _, m := range Muxy.MiddleWares() {
			log.Info("Disable type %s", reflect.TypeOf(m))
			m.Teardown()
			fmt.Fprintf(w, "%s.\n", reflect.TypeOf(m))
		}

		log.Info("Disable the network shape of the current Muxy successfully")
		fmt.Fprint(w, "Disable Muxy successfully")

	case "stressor":
		requestBody, _ := ioutil.ReadAll(r.Body)
		requestBodyStr := string(requestBody)
		log.Debug("Received body : \n%s", requestBodyStr)

		var param stressng.Param
		if err := json.Unmarshal(requestBody, &param); err != nil {
			fmt.Fprintf(w, "Unmashaling has an error: %s", err.Error())
			return
		}

		// List all iptables
		cmdStrings := []string{"stress-ng", fmt.Sprintf("--%s %d", param.Stressor, param.StressorCount),
			fmt.Sprintf("--timeout %d%s", param.Timeout, param.TimeoutUnit),
		}

		log.Debug("Stressor name:[%s], count:[%d]", log.Colorize(log.GREEN, param.Stressor), param.StressorCount)

		switch param.Stressor {
		case "cpu":
			cmdStrings = append(cmdStrings, fmt.Sprintf("-l %d", param.CpuLoad))
			log.Debug("Cpu load : %s", string(param.CpuLoad))
		case "vm":

		case "iomix":

		case "hdd":

		default:
			log.Debug("Not support stressor named : %s", log.Colorize(log.RED, param.Stressor))
			fmt.Fprintf(w, "Not support stressor named : %s", param.Stressor)
			return
		}

		if param.Abort {
			cmdStrings = append(cmdStrings, "--abort")
		}

		if param.Metrics {
			cmdStrings = append(cmdStrings, "--metrics-brief")
		}

		//--vm 2 --timeout 30s --metrics-brief --abort
		//--cpu 0 --cpu-method all --timeout 60s --metrics-brief --abort
		//--iomix 10 --timeout 30s --metrics-brief --abort
		cmdStr := strings.Join(cmdStrings, " ")
		cmd := exec.Command("/bin/bash", "-c", cmdStr)

		if err := cmd.Start(); err != nil {
			log.Fatalf("Execute stress-ng has an error:%s", err.Error())
			fmt.Fprintf(w, err.Error())
			return
		} else {
			pid := cmd.Process.Pid
			wait := make(chan int)

			stressng.RunningStressor[cmd.Process.Pid] = *cmd

			// Wait for a stressor finished in another goroutine.
			go func() {
				pid := <-wait

				/*
				 * 1. The process exit automatically
				 * 2. Kill this process manually.
				 */
				cmd.Wait()

				if pid != 0 {
					delete(stressng.RunningStressor, pid)
					log.Info("Deleting the stressor with key [%d] from running stressor map if it finished automatically or it has been killed manually.",
						pid)
				}
			}()

			wait <- pid
		}
		//	cmd.Wait()
		//}()
		//<-wait

		//time.Sleep(time.Second * 2)

		//go cmd.Run()

		log.Debug("Command : %s, pid : [%d]", log.Colorize(log.GREEN, cmdStr), cmd.Process.Pid)
		//cmd.Process.Release()
		fmt.Fprintf(w, "Commit command successfully: \n%s\n%s\n%d", cmd.Args, cmdStr, cmd.Process.Pid)

	case "runningStressors":
		runingStressorsJson, err := json.MarshalIndent(stressng.RunningStressor, "", "")
		if err == nil {
			stressorsJon := string(runingStressorsJson)
			log.Debug("Running stressors : %s", stressorsJon)
			fmt.Fprint(w, stressorsJon)
		} else {
			log.Error(err.Error())
			fmt.Fprint(w, err.Error())
		}
	case "killStressor":
		pidStrs, ok := r.URL.Query()["pid"]
		if !ok || len(pidStrs[0]) < 1 {
			fmt.Fprint(w, "You must specify a pid!")
			log.Error("Can not find a pid")
			return
		}

		pid, err := strconv.Atoi(pidStrs[0])
		if err == nil {
			cmd := stressng.RunningStressor[pid]
			if cmd.Process == nil {
				fmt.Fprintf(w, "The process with pid [%d] has not exist, it may be finished or be killed already.", pid)
				log.Info("The process with pid [%d] has not exist, it may be finished or be killed already.", pid)
				return
			}

			if err := cmd.Process.Kill(); err == nil {
				//delete(stressng.RunningStressor, pid)
				log.Info("Killing the stressor with key [%d] manually.", pid)
				fmt.Fprintf(w, "Process with pid [%d] has been killed successfully", pid)
			} else {
				fmt.Fprint(w, err.Error())
				log.Error(err.Error())
			}
		} else {
			fmt.Fprint(w, err.Error())
			log.Error(err.Error())
		}
	case "stopAllStressors":
		if len(stressng.RunningStressor) <= 0 {
			fmt.Fprint(w, "No running streesor here")
			log.Warn("No running streesor here")
			return
		}

		for pid, cmd := range stressng.RunningStressor {
			if cmd.Process != nil {
				if err := cmd.Process.Kill(); err == nil {
					log.Info("Killing the stressor with key [%d] manually.", pid)
					fmt.Fprintf(w, "Process with pid [%d] has been killed successfully\n", pid)
				} else {
					fmt.Fprint(w, err.Error())
					log.Error(err.Error())
				}
			}
		}

	default:
		fmt.Fprintf(w, "Can not support %s method", r.URL.Path[1:])
		log.Debug("Can not support %s method", log.Colorize(log.RED, r.URL.Path[1:]))
	}

}

func (s *MuxyApiServer) Start() error {
	server := &http.Server{
		Addr:           port,
		Handler:        http.HandlerFunc(apiHandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(server.ListenAndServe())

	return nil
}
