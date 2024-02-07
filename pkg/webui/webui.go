package webui

import (
	"errors"
	"net"
	"net/http"

	"github.com/christophwitzko/wg-hub/pkg/api"
	"github.com/christophwitzko/wg-hub/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

type Server struct {
	router *chi.Mux
	log    *logrus.Logger
	cfg    *config.Config
	api    *api.API
}

func newServer(log *logrus.Logger, dev *device.Device, cfg *config.Config) *Server {
	w := &Server{
		router: chi.NewRouter(),
		log:    log,
		cfg:    cfg,
		api:    api.NewAPIServer(log, dev, cfg),
	}
	w.router.Get("/*", getWebuiServer())
	w.router.Mount("/api", w.api)
	return w
}

func (a *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func StartServer(log *logrus.Logger, dev *device.Device, cfg *config.Config, tunNet *netstack.Net) error {
	listener, err := tunNet.ListenTCP(&net.TCPAddr{Port: 80})
	if err != nil {
		return err
	}
	server := &http.Server{Handler: newServer(log, dev, cfg)}
	go func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to start webui server: %v", err)
		}
	}()
	return nil
}
