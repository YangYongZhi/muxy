package http

import (
	"fmt"

	"encoding/json"
	"github.com/YangYongZhi/muxy/log"
	m "github.com/YangYongZhi/muxy/run"
	"github.com/YangYongZhi/muxy/throttler"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/YangYongZhi/muxy/middleware"
	"github.com/YangYongZhi/muxy/symptom"
)

const (
	port = ":13003"
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
	log.Debug("request the %s api", log.Colorize(log.YELLOW, r.URL.Path[1:]))
	method := r.URL.Path[1:]
	fmt.Fprintf(w, "Hi, your resource is %s\n", method)

	switch method {
	case "check":
		middlewares := Muxy.MiddleWares()
		if len(middlewares) > 0 {
			fmt.Fprintf(w, "Muxy has been start, middleware count : %d\n", len(middlewares))
		}

		//e.g. body = {"Device":"ens33","Latency": 2000, "TargetBandWidth":20,"PacketLoss":70,"TargetPorts": ["5001","10090"], "TargetProtos":["tcp","icmp"]}
	case "reset":
		log.Info("Reset the middlewares, please wait.")

		//decode := json.NewDecoder(r.Body)

		body, _ := ioutil.ReadAll(r.Body)
		//    r.Body.Close()
		body_str := string(body)
		log.Debug("Body : %s", body_str)

		var config throttler.Config

		if err := json.Unmarshal(body, &config); err == nil {
			log.Debug("latency :%s", config.Latency)
			fmt.Fprint(w, body_str)
		} else {
			log.Error("Reset a network shape has an erro", err)
			fmt.Fprint(w, err)
		}

		log.Debug("The parameters for reset the work shape on current muxy: %s", body_str)

		for _, m := range Muxy.MiddleWares() {
			log.Debug("%s", reflect.TypeOf(m))

			switch v := m.(type) {
			case *middleware.LoggerMiddleware:
				log.Debug("Not support %v now.", v)
			case *symptom.HTTPDelaySymptom:
				log.Debug("Not support %v now.", v)
			case *symptom.NetworkShaperSymptom:
				v.Teardown()

				// We should sleep this current goroutine in order to make the ip tables change less frequently.
				log.Debug("We should sleep this current goroutine in order to make the ip tables change less frequently.")
				time.Sleep(5 * time.Second)

				v.Device = config.Device
				v.Latency = config.Latency
				v.PacketLoss = config.PacketLoss
				v.TargetBandwidth = config.TargetBandwidth
				v.TargetPorts = config.TargetPorts
				v.TargetProtos = config.TargetProtos

				v.Config.Device = config.Device
				v.Config.Latency = config.Latency
				v.Config.PacketLoss = config.PacketLoss
				v.Config.TargetBandwidth = config.TargetBandwidth
				v.Config.TargetPorts = config.TargetPorts
				v.Config.TargetProtos = config.TargetProtos

				v.Setup()
			}

		}

		log.Info("Create some middleware for current Muxy successfully")
		fmt.Fprint(w, "Reset the network shape successfully.")
	case "enable":
		log.Info("Enable the network shape of the current Muxy, please wait.")

		for _, m := range Muxy.MiddleWares() {
			log.Info("Enable type %s", reflect.TypeOf(m))
			m.Setup()
		}

		log.Info("Enable Muxy successfully")
	case "disable":
		log.Info("Disable all rules with current Muxy, please wait.")

		for _, m := range Muxy.MiddleWares() {
			log.Info("Disable type %s", reflect.TypeOf(m))
			m.Teardown()
		}

		log.Info("Disable the network shape of the current Muxy successfully")
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