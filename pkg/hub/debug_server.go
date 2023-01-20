package hub

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func debugHandler(log *logrus.Logger, dev *device.Device) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("debug request from %s - %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		configStr, err := dev.IpcGet()
		if err != nil {
			log.Errorf("failed to get ipc operation: %v", err)
			http.Error(w, "failed to get ipc operation", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")

		for _, line := range strings.Split(configStr, "\n") {
			if strings.HasPrefix(line, "private_key") {
				_, _ = io.WriteString(w, "private_key=[...]\n")
				continue
			}
			_, _ = io.WriteString(w, line+"\n")
		}
	})
}

func StartDebugServer(log *logrus.Logger, dev *device.Device, tunNet *netstack.Net) error {
	listener, err := tunNet.ListenTCP(&net.TCPAddr{Port: 8080})
	if err != nil {
		return err
	}
	server := &http.Server{Handler: debugHandler(log, dev)}
	go func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to start debug server: %v", err)
		}
	}()
	return nil
}
