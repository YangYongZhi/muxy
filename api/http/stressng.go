package http

import (
	"encoding/json"
	"fmt"
	"github.com/YangYongZhi/muxy/log"
	"github.com/YangYongZhi/muxy/stressng"
	"github.com/mux"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type response struct {
	Cmd       string
	Args      []string
	ProcessId int
	State     int
}

/**
 *
 */
func stressorHandler(w http.ResponseWriter, r *http.Request) {
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

	var resp = response{}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Execute stress-ng has an error:%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	//fmt.Fprintf(w, "Commit command successfully: \n%s\n%s\n%d", cmd.Args, cmdStr, cmd.Process.Pid)
	resp.Cmd = cmdStr
	resp.Args = cmd.Args
	resp.ProcessId = cmd.Process.Pid
	resp.State = http.StatusOK

	json.NewEncoder(w).Encode(resp)
}

/**
 *
 */
func runningStressorsHandler(w http.ResponseWriter, r *http.Request) {
	runingStressorsJson, err := json.MarshalIndent(stressng.RunningStressor, "", "")
	if err == nil {
		stressorsJon := string(runingStressorsJson)
		log.Debug("Running stressors : %s", stressorsJon)
		fmt.Fprint(w, stressorsJon)
	} else {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/**
 *
 */
func killStressorHandler(w http.ResponseWriter, r *http.Request) {
	//pidStrs, ok := r.URL.Query()["pid"]
	//if !ok || len(pidStrs[0]) < 1 {
	//	fmt.Fprint(w, "You must specify a pid!")
	//	log.Error("Can not find a pid")
	//	return
	//}

	params := mux.Vars(r)
	pid, err := strconv.Atoi(params["processId"])
	if err == nil {
		cmd := stressng.RunningStressor[pid]
		if cmd.Process == nil {
			http.Error(w, fmt.Sprintf("The process with pid [%d] has not exist, it may be finished or be killed already.", pid), http.StatusNotFound)
			log.Info("The process with pid [%d] has not exist, it may be finished or be killed already.", pid)
			return
		}

		if err := cmd.Process.Kill(); err == nil {
			//delete(stressng.RunningStressor, pid)
			log.Info("Killing the stressor with key [%d] manually.", pid)
			fmt.Fprintf(w, "Process with pid [%d] has been killed successfully", pid)
		} else {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/**
 *
 */
func killAllStressorsHandler(w http.ResponseWriter, r *http.Request) {
	if len(stressng.RunningStressor) <= 0 {
		http.Error(w, "No running streesor here", http.StatusNotFound)
		log.Warn("No running streesor here")
		return
	}

	for pid, cmd := range stressng.RunningStressor {
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err == nil {
				log.Info("Killing the stressor with key [%d] manually.", pid)
				fmt.Fprintf(w, "Process with pid [%d] has been killed successfully\n", pid)
			} else {
				log.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
