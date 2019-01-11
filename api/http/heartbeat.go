package http

import (
	time2 "time"
	"log"
	"github.com/gocron"
	"net/http"
	"github.com/pquerna/ffjson/ffjson"
	"bytes"
	"gitlab.dmall.com/arch/datahub-agent/net"
	"fmt"
	"os"
)

const (
	serverURL   = `http://wonton.dmall.com:13004/wonton/heartbeat`
	Port        = 13003
	sendInteval = 30
	version     = `WONTON-AGENT-0.0.1`
)

type HeartBeatSender struct {
}

type HeartBeats struct {
	Host         string `json:"host"`
	Instancecode string `json:"instancecode"`
	Time         int64  `json:"time"`
	Version      string `json:"version"`
	Pid          int    `json:"pid"`
	StartTime    int64  `json:"startTime"`
}

func (*HeartBeatSender) Start() {
	startTime := time2.Now().UnixNano() / 1e6
	hostIp := net.GetOutboundIP().String()
	instanceCode := fmt.Sprintf("%s:%d", hostIp, Port)
	pid := os.Getpid()
	gocron.Every(sendInteval).Seconds().Do(sendHeartBeat, startTime, hostIp, instanceCode, pid)

	// function Start start all the pending jobs
	<-gocron.Start()
}

func sendHeartBeat(startTime int64, hostIp string, instanceCode string, pid int) {
	params := &HeartBeats{Host: hostIp, Instancecode: instanceCode, Time: time2.Now().UnixNano() / 1e6, Version: version, Pid: pid, StartTime: startTime}
	log.Printf("The heartbeat Object that send to wonton server : %v", params)
	postParams, err := ffjson.Marshal(params)
	if err != nil {
		log.Print(err.Error())
		return
	}

	//log.Printf("Send a heartbeat as String to dmc : \n%v", string(postParams))

	response, err := http.Post(serverURL, "application/json", bytes.NewBuffer(postParams))
	if err != nil {
		log.Print(err.Error())
		return
	}

	defer response.Body.Close()
}
