package wgdebug

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/christophwitzko/wireguard-hub/pkg/wgconn"

	"github.com/christophwitzko/wireguard-hub/pkg/config"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func debugHandler(log *logrus.Logger, dev *device.Device) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("debug request: %s %s", r.Method, r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != "/get_ipc" {
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

func startDebugServer(log *logrus.Logger, dev *device.Device, listener net.Listener) func() {
	server := &http.Server{Handler: debugHandler(log, dev)}
	go func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("failed to start debug server: %v", err)
		}
	}()
	return func() {
		log.Info("stopping debug server...")
		err := server.Close()
		if err != nil {
			log.Errorf("failed to close debug server: %v", err)
		}
	}
}

func Init(log *logrus.Logger, dev *device.Device, cfg *config.Config) (func(), error) {
	pKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	closeFn, tunNet, err := createWgDevice(log, cfg, pKey)
	if err != nil {
		return nil, err
	}

	wgConf := &bytes.Buffer{}
	wgConf.WriteString("public_key=" + config.MustGet(config.Base64ToHex(pKey.PublicKey().String())) + "\n")
	wgConf.WriteString("allowed_ip=" + cfg.DebugAddress + "/32\n")
	err = dev.IpcSetOperation(wgConf)
	if err != nil {
		closeFn()
		return nil, err
	}

	listener, err := tunNet.ListenTCP(&net.TCPAddr{Port: 80})
	if err != nil {
		closeFn()
		return nil, err
	}
	stopFn := startDebugServer(log, dev, listener)
	return func() {
		stopFn()
		closeFn()
	}, nil
}

func createWgDevice(log *logrus.Logger, cfg *config.Config, pkey wgtypes.Key) (func(), *netstack.Net, error) {
	dbgIP, err := netip.ParseAddr(cfg.DebugAddress)
	if err != nil {
		return nil, nil, err
	}
	tunDev, tunNet, err := netstack.CreateNetTUN([]netip.Addr{dbgIP}, nil, device.DefaultMTU)
	if err != nil {
		return nil, nil, err
	}
	dev := device.NewDevice(tunDev, wgconn.NewStdNetBind(cfg.BindAddress), &device.Logger{
		Errorf:   log.Errorf,
		Verbosef: device.DiscardLogf,
	})

	hubPublicKey, err := config.Base64ToHex(cfg.PrivateKey.PublicKey().String())
	if err != nil {
		return nil, nil, err
	}

	epAddr := cfg.ResolvedBindAddr()
	if epAddr == "" {
		epAddr = "127.0.0.1"
	}
	wgConf := &bytes.Buffer{}
	wgConf.WriteString("private_key=" + hex.EncodeToString(pkey[:]) + "\n")
	wgConf.WriteString("public_key=" + hubPublicKey + "\n")
	wgConf.WriteString("endpoint=" + fmt.Sprintf("%s:%d", epAddr, cfg.Port) + "\n")
	wgConf.WriteString("allowed_ip=0.0.0.0/0\n")
	wgConf.WriteString("persistent_keepalive_interval=5\n")
	err = dev.IpcSetOperation(wgConf)
	if err != nil {
		return nil, nil, err
	}

	err = dev.Up()
	if err != nil {
		return nil, nil, err
	}

	return func() {
		dev.Close()
	}, tunNet, nil
}
