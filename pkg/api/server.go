package api

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/christophwitzko/wg-hub/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
)

type API struct {
	router    *chi.Mux
	log       *logrus.Logger
	dev       *device.Device
	cfg       *config.Config
	tokenAuth *jwtauth.JWTAuth
	ipcMutex  sync.Mutex
}

func NewAPIServer(log *logrus.Logger, dev *device.Device, cfg *config.Config) *API {
	var jwtSecret bytes.Buffer
	if cfg.WebuiJWTSecret == "" {
		log.Warnf("using random jwt secret")
		_, err := jwtSecret.ReadFrom(io.LimitReader(rand.Reader, 32))
		if err != nil {
			panic(err)
		}
	} else {
		jwtSecret.WriteString(cfg.WebuiJWTSecret)
	}
	a := &API{
		router:    chi.NewRouter(),
		log:       log,
		dev:       dev,
		cfg:       cfg,
		tokenAuth: jwtauth.New("HS256", jwtSecret.Bytes(), nil),
	}
	a.initRoutes()
	return a
}

func (a *API) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.log.Infof("api request: %s - %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err != nil {
			a.sendError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if token == nil {
			a.sendError(w, "auth token missing", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) sendError(w http.ResponseWriter, err string, code int) {
	a.log.Warnf("api error (code=%d): %s", code, err)
	a.writeJSON(w, map[string]string{"error": err}, code)
}

func (a *API) writeJSON(w http.ResponseWriter, d any, code ...int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(code) > 0 {
		w.WriteHeader(code[0])
	}
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		a.log.Errorf("api json write error: %v", err)
	}
}

func (a *API) initRoutes() {
	a.router.Use(a.loggerMiddleware)
	a.router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		a.sendError(w, "not found", http.StatusNotFound)
	})

	// public routes
	a.router.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			a.writeJSON(w, map[string]string{"status": "ok"})
		})
		r.Post("/auth", a.createAuth)
	})

	// protected routes
	a.router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(a.tokenAuth))
		r.Use(a.authMiddleware)

		// auth routes
		r.Get("/auth", a.getAuth)

		// peers api
		r.Get("/peers", a.listPeers)
		r.Post("/peers", a.generatePeer)
		r.Put("/peers/*", a.addPeer)
		r.Delete("/peers/*", a.removePeer)

		// config api
		r.Get("/config", a.getConfig)

		// hub api
		r.Get("/hub", a.getHubInfo)
	})
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}
