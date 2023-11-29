package api

import (
	"encoding/json"
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
		a.log.Infof("api request: %s - %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (a *apiServer) sendError(w http.ResponseWriter, err string, code int) {
	a.log.Warnf("api error (code=%d): %s", code, err)
	w.WriteHeader(code)
	a.writeJSON(w, map[string]string{"error": err})
}

func (a *apiServer) writeJSON(w http.ResponseWriter, d any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		a.log.Errorf("api json write error: %v", err)
	}
}

func (a *apiServer) initRoutes() {
	a.router.Use(a.loggerMiddleware)
	a.router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		a.sendError(w, "not found", http.StatusNotFound)
	})
	a.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		a.writeJSON(w, map[string]string{"status": "ok"})
	})
	a.router.Route("/peers", func(r chi.Router) {
		r.Get("/", a.listPeers)
		r.Post("/", a.addPeer)
		r.Delete("/{publicKey}", a.removePeer)
	})
}

func (a *apiServer) listPeers(w http.ResponseWriter, _ *http.Request) {
	// TODO
	a.sendError(w, "not implemented", http.StatusNotImplemented)
}

func (a *apiServer) addPeer(w http.ResponseWriter, _ *http.Request) {
	// TODO
	a.sendError(w, "not implemented", http.StatusNotImplemented)
}

func (a *apiServer) removePeer(w http.ResponseWriter, _ *http.Request) {
	// TODO
	a.sendError(w, "not implemented", http.StatusNotImplemented)
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
