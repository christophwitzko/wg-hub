package api

import (
	"errors"
	"net"
	"net/http"

	"github.com/christophwitzko/wg-hub/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

type apiServer struct {
	router *chi.Mux
	log    *logrus.Logger
	dev    *device.Device
	cfg    *config.Config
}

func newAPIServer(log *logrus.Logger, dev *device.Device, cfg *config.Config) *apiServer {
	a := &apiServer{
		router: chi.NewRouter(),
		log:    log,
		dev:    dev,
		cfg:    cfg,
	}
	a.initRoutes()
	return a
}

func (a *apiServer) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.log.Infof("[API REQUEST] %s - %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (a *apiServer) initRoutes() {
	a.router.Use(a.loggerMiddleware)
	a.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (a *apiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func StartServer(log *logrus.Logger, dev *device.Device, cfg *config.Config, tunNet *netstack.Net) error {
	listener, err := tunNet.ListenTCP(&net.TCPAddr{Port: 80})
	if err != nil {
		return err
	}
	server := &http.Server{Handler: newAPIServer(log, dev, cfg)}
	go func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to start api server: %v", err)
		}
	}()
	return nil
}
