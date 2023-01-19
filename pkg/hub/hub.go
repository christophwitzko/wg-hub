package hub

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/netip"

	"github.com/christophwitzko/wireguard-hub/pkg/config"
	"github.com/christophwitzko/wireguard-hub/pkg/wgconn"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Init(log *logrus.Logger, dev *device.Device, cfg *config.Config) (func(), *netstack.Net, error) {
	pKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}
	closeFn, tunNet, err := createWgDevice(log, cfg, pKey)
	if err != nil {
		return nil, nil, err
	}

	wgConf := &bytes.Buffer{}
	wgConf.WriteString("public_key=" + config.MustGet(config.Base64ToHex(pKey.PublicKey().String())) + "\n")
	wgConf.WriteString("allowed_ip=" + cfg.HubAddress + "/32\n")
	err = dev.IpcSetOperation(wgConf)
	if err != nil {
		closeFn()
		return nil, nil, err
	}

	return closeFn, tunNet, nil
}

func createWgDevice(log *logrus.Logger, cfg *config.Config, pkey wgtypes.Key) (func(), *netstack.Net, error) {
	hubIP, err := netip.ParseAddr(cfg.HubAddress)
	if err != nil {
		return nil, nil, err
	}
	tunDev, tunNet, err := netstack.CreateNetTUN([]netip.Addr{hubIP}, nil, device.DefaultMTU)
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
		log.Infof("closing wg hub device...")
		dev.Close()
	}, tunNet, nil
}
