package wgdebug

import (
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
)

func StartDebugServer(log *logrus.Logger, dev *device.Device, addr string) (func(), error) {
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infof("debug request: %s %s", r.Method, r.URL.Path)
			if r.Method != http.MethodGet || r.URL.Path != "/get_ipc" {
				http.NotFound(w, r)
				return
			}
			config, err := dev.IpcGet()
			if err != nil {
				log.Errorf("failed to get ipc operation: %v", err)
				http.Error(w, "failed to get ipc operation", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")

			for _, line := range strings.Split(config, "\n") {
				if strings.HasPrefix(line, "private_key") {
					_, _ = io.WriteString(w, "private_key=[...]\n")
					continue
				}
				_, _ = io.WriteString(w, line+"\n")
			}
		}),
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("failed to start debug server: %v", err)
		}
	}()
	stopFn := func() {
		log.Info("stopping debug server...")
		err := server.Close()
		if err != nil {
			log.Errorf("failed to close debug server: %v", err)
		}
	}
	return stopFn, nil
}
