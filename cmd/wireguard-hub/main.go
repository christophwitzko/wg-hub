package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustGet[T any](val T, err error) T {
	must(err)
	return val
}

func initViper(cmd *cobra.Command) error {
	configFile := mustGet(cmd.Flags().GetString("config"))
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("wireguard-hub.yaml")
	}
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var viperConfigNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &viperConfigNotFound) {
			return err
		}
	}
	return nil
}

func parseLogLevel(logLevel string) (logrus.Level, bool) {
	switch strings.ToLower(logLevel) {
	case "d", "debug":
		return logrus.DebugLevel, true
	case "i", "info":
		return logrus.InfoLevel, true
	case "w", "warn":
		return logrus.WarnLevel, true
	case "e", "error":
		return logrus.ErrorLevel, true
	case "f", "fatal":
		return logrus.FatalLevel, true
	}
	return logrus.DebugLevel, false
}

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
	rootCmd.PersistentFlags().StringArrayP("peer", "p", nil, "base64 encoded public key and allowed ips of peer (e.g. -p \"publicKey,allowedIPs\")")
	rootCmd.PersistentFlags().String("config", "", "config file (default is .wireguard-hub.yaml)")
	rootCmd.PersistentFlags().String("log-level", "debug", "log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().SortFlags = true

	must(viper.BindPFlag("privateKey", rootCmd.PersistentFlags().Lookup("private-key")))
	must(viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port")))
	must(viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("log-level")))

	cobra.OnInitialize(func() {
		if err := initViper(rootCmd); err != nil {
			log.Errorf("failed to load config: %v", err)
			os.Exit(1)
		}

		logLevel := viper.GetString("logLevel")
		parsedLogLevel, ok := parseLogLevel(logLevel)
		if !ok {
			log.Warnf("failed to parse log level: %s", logLevel)
		}
		log.SetLevel(parsedLogLevel)

		usedConfigFile := viper.ConfigFileUsed()
		if usedConfigFile != "" {
			log.Infof("using config: %s", usedConfigFile)
		}
	})

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

func parseConfig(log *logrus.Logger, cmd *cobra.Command) (*Config, error) {
	privateKey := viper.GetString("privateKey")
	if privateKey == "" {
		return nil, fmt.Errorf("private-key is required")
	}
	privateKey, err := base64ToHex(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private-key: %w", err)
	}

	port := viper.GetUint16("port")
	log.Infof("listening port: %d", port)

	inputPeers := mustGet(cmd.Flags().GetStringArray("peer"))
	var configPeers []map[string]string
	err = viper.UnmarshalKey("peers", &configPeers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peers from config: %w", err)
	}
	for _, peer := range configPeers {
		inputPeers = append(inputPeers, fmt.Sprintf("%s,%s", peer["publickey"], peer["allowedips"]))
	}
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
	log.Println("stopping...")
	stop()
	dev.Close()
	log.Println("stopped")
	return nil
}
