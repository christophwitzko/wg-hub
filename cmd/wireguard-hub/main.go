package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/christophwitzko/wireguard-hub/pkg/config"
	"github.com/christophwitzko/wireguard-hub/pkg/hub"
	"github.com/christophwitzko/wireguard-hub/pkg/loopback"
	"github.com/christophwitzko/wireguard-hub/pkg/wgconn"
	"github.com/christophwitzko/wireguard-hub/pkg/wgdebug"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.DebugLevel)

	rootCmd := &cobra.Command{
		Use:   "wireguard-hub",
		Short: "wireguard user space hub",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(log, cmd, args); err != nil {
				log.Errorf("ERROR: %v", err)
				os.Exit(1)
			}
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	config.SetFlags(rootCmd)

	cobra.OnInitialize(func() {
		config.OnInitialize(log, rootCmd)
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(log *logrus.Logger, cmd *cobra.Command, _ []string) error {
	cfg, err := config.ParseConfig(log, cmd)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tunDev := loopback.CreateTun(device.DefaultMTU)
	devLogger := &device.Logger{
		Verbosef: log.Debugf,
		Errorf:   log.Errorf,
	}
	dev := device.NewDevice(tunDev, wgconn.NewStdNetBind(cfg.BindAddress), devLogger)

	wgConf := &bytes.Buffer{}
	wgConf.WriteString("private_key=" + cfg.PrivateKeyHex + "\n")
	wgConf.WriteString("listen_port=" + cfg.GetPort() + "\n")
	for _, peer := range cfg.Peers {
		wgConf.WriteString("public_key=" + peer.PublicKeyHex + "\n")
		wgConf.WriteString("allowed_ip=" + peer.AllowedIP + "\n")
	}
	err = dev.IpcSetOperation(wgConf)
	if err != nil {
		return err
	}
	err = dev.Up()
	if err != nil {
		return err
	}

	stopHubInstance := func() {}
	var tunNet *netstack.Net
	if cfg.HubAddress != "" {
		log.Infof("starting hub instance on %s", cfg.HubAddress)
		stopHubInstance, tunNet, err = hub.Init(log, dev, cfg)
		if err != nil {
			return fmt.Errorf("failed to start debug server: %w", err)
		}
	}

	if cfg.DebugServer && tunNet != nil {
		log.Infof("starting debug server on http://%s:8080", cfg.HubAddress)
		err = wgdebug.StartDebugServer(log, dev, tunNet)
		if err != nil {
			return fmt.Errorf("failed to start debug server: %w", err)
		}
	}

	<-ctx.Done()
	log.Println("stopping...")
	stop()
	stopHubInstance()
	dev.Close()
	log.Println("stopped")
	return nil
}
