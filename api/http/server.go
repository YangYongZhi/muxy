package http

import (
	"github.com/YangYongZhi/muxy/log"
	"net/http"
	"time"
)

type MuxyApiServer struct {
	name   string
	server *http.Server
}

func New(name string) *MuxyApiServer {
	return &MuxyApiServer{name: name}
}

func ViewExisthandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("handler - ViewExisthandler")
}

func (s *MuxyApiServer) Start() error {
	s.server = &http.Server{
		Addr:           ":8080",
		Handler:        http.HandlerFunc(ViewExisthandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.server.ListenAndServe())

	return nil
}
