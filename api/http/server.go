package http

import (
	"github.com/YangYongZhi/muxy/log"
	m "github.com/YangYongZhi/muxy/run"
	"github.com/mux"
	"net/http"
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

func (s *MuxyApiServer) Start() error {
	r := mux.NewRouter()
	r.HandleFunc("/muxy/networkshape/_enable", enableHanlder).Methods("PUT")
	r.HandleFunc("/muxy/networkshape/_disable", disableHanlder).Methods("PUT")
	r.HandleFunc("/muxy/networkshape/_reset", resetHanlder).Methods("PUT")
	r.HandleFunc("/muxy/networkshapes", networkshapeHanlder).Methods("GET")
	r.HandleFunc("/muxy/middlewares", middlewaresHanlder).Methods("GET")

	r.HandleFunc("/stressng/stressor", stressorHandler).Methods("POST")
	r.HandleFunc("/stressng/stressors", runningStressorsHandler).Methods("GET")
	r.HandleFunc("/stressng/stressor/{processId}", killStressorHandler).Methods("DELETE")
	r.HandleFunc("/stressng/stressors", killAllStressorsHandler).Methods("DELETE")

	//server := &http.Server{
	//	Addr:           port,
	//	Handler:        http.HandlerFunc(apiHandler),
	//	ReadTimeout:    10 * time.Second,
	//	WriteTimeout:   10 * time.Second,
	//	MaxHeaderBytes: 1 << 20,
	//}

	log.Fatal(http.ListenAndServe(port, r))

	return nil
}
