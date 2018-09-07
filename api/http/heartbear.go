package http

import (
	"github.com/YangYongZhi/muxy/log"
	"github.com/gocron"
	"io/ioutil"
	"net/http"
)

const (
	SendDomain  = `http://localhost:13003/stressng/stressors`
	SendInteval = 30
)

type HeartBeatSender struct {
}

func (*HeartBeatSender) Start() {
	gocron.Every(SendInteval).Seconds().Do(sendHeartBeat)

	// function Start start all the pending jobs
	<-gocron.Start()
}

func sendHeartBeat() {
	response, err := http.Get(SendDomain)
	if err != nil {
		log.Error(err.Error())
		return
	} else {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error(err.Error())
			return
		} else {
			log.Debug("The content for a heartbeat:\n%s", string(content))
		}
	}

	_, time := gocron.NextRun()
	log.Debug("Send a heartbeat, next run time at %v", time)
}
