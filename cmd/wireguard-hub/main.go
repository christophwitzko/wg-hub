package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/christophwitzko/wireguard-hub/pkg/loopback"
	"github.com/christophwitzko/wireguard-hub/pkg/wgconn"
	"github.com/christophwitzko/wireguard-hub/pkg/wgdebug"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/device"
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

	setFlags(rootCmd)

	cobra.OnInitialize(func() {
		onInitialize(log, rootCmd)
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(log *logrus.Logger, cmd *cobra.Command, _ []string) error {
	cfg, err := parseConfig(log, cmd)
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
	dev := device.NewDevice(tunDev, wgconn.NewStdNetBind(cfg.BindAddr), devLogger)

	wgConf := &bytes.Buffer{}
	wgConf.WriteString("private_key=" + cfg.PrivateKey + "\n")
	wgConf.WriteString("listen_port=" + cfg.GetPort() + "\n")
	for _, peer := range cfg.Peers {
		wgConf.WriteString("public_key=" + peer.PublicKeyHex + "\n")
		wgConf.WriteString("allowed_ip=" + peer.AllowedIP + "\n")
	}
	if err := dev.IpcSetOperation(wgConf); err != nil {
		return err
	}
	if err := dev.Up(); err != nil {
		return err
	}

	stopDebugServer := func() {}
	if cfg.DebugPort > 0 {
		log.Infof("starting debug server on port %d", cfg.DebugPort)
		stopDebugServer, err = wgdebug.StartDebugServer(log, dev, fmt.Sprintf(":%d", cfg.DebugPort))
		if err != nil {
			return fmt.Errorf("failed to start debug server: %w", err)
		}
	}

	<-ctx.Done()
	log.Println("stopping...")
	stop()
	stopDebugServer()
	dev.Close()
	log.Println("stopped")
	return nil
}
