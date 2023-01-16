package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/christophwitzko/wireguard-hub/pkg/loopback"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/conn"
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
		Short: "wireguard userspace hub",
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

	rootCmd.PersistentFlags().String("private-key", "", "base64 encoded private key")
	rootCmd.PersistentFlags().Uint16("port", 9999, "port to listen on")
	rootCmd.PersistentFlags().StringArrayP("peer", "p", nil, "base64 encoded public key of peer and allowed ip (e.g. -p \"base64_encoded_public_key,allowed_ip\")")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func base64ToHex(b64 string) (string, error) {
	decKey, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decKey), nil
}

type Peer struct {
	PublicKey    string
	PublicKeyHex string
	AllowedIP    string
}

func NewPeer(peerConfig string) (*Peer, error) {
	publicKey, ip, ok := strings.Cut(peerConfig, ",")
	if !ok {
		return nil, fmt.Errorf("failed to parse peer config: %s", peerConfig)
	}
	p := &Peer{
		PublicKey: publicKey,
	}
	publicKeyHex, err := base64ToHex(p.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer public key: %w", err)
	}
	p.PublicKeyHex = publicKeyHex

	// add subnet mask if not present
	if !strings.Contains(ip, "/") {
		ip = fmt.Sprintf("%s/32", ip)
	}

	// check if ip is valid
	ipPrefix, err := netip.ParsePrefix(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowed ip: %w", err)
	}
	p.AllowedIP = ipPrefix.String()
	return p, nil
}

func (p *Peer) String() string {
	return fmt.Sprintf("peer(%s…%s): %s", p.PublicKeyHex[:4], p.PublicKeyHex[len(p.PublicKeyHex)-4:], p.AllowedIP)
}

type Config struct {
	PrivateKey string
	Port       uint16
	Peers      []*Peer
}

func (c *Config) GetPort() string {
	return strconv.FormatUint(uint64(c.Port), 10)
}

func mustGet[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func parseConfig(log *logrus.Logger, cmd *cobra.Command) (*Config, error) {
	privateKey := mustGet(cmd.Flags().GetString("private-key"))
	if privateKey == "" {
		return nil, fmt.Errorf("private-key is required")
	}
	privateKey, err := base64ToHex(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private-key: %w", err)
	}

	port := mustGet(cmd.Flags().GetUint16("port"))
	log.Infof("listening port: %d", port)

	inputPeers := mustGet(cmd.Flags().GetStringArray("peer"))
	if len(inputPeers) == 0 {
		return nil, fmt.Errorf("at least one peer is required")
	}
	peers := make([]*Peer, len(inputPeers))
	for i, peerConfig := range inputPeers {
		p, err := NewPeer(peerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse peer %d: %w", i, err)
		}
		peers[i] = p
		log.Infof("adding %s", p)
	}
	// TODO: check ip ranges overlap
	return &Config{
		PrivateKey: privateKey,
		Port:       port,
		Peers:      peers,
	}, nil
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
	dev := device.NewDevice(tunDev, conn.NewDefaultBind(), devLogger)

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

	<-ctx.Done()
	log.Println("shutting down...")
	stop()
	dev.Close()
	log.Println("hub stopped!")
	return nil
}
