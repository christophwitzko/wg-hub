package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/christophwitzko/wg-hub/pkg/ipc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustGet[T any](val T, err error) T {
	Must(err)
	return val
}

func initViper(cmd *cobra.Command) error {
	configFile := MustGet(cmd.Flags().GetString("config"))
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

func OnInitialize(log *logrus.Logger, cmd *cobra.Command) {
	if err := initViper(cmd); err != nil {
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

func SetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("private-key", "", "base64 encoded private key of the hub")
	cmd.PersistentFlags().Uint16("port", 9999, "port to listen on")
	cmd.PersistentFlags().String("bind-address", "", "address to bind on")
	cmd.PersistentFlags().StringArrayP("peer", "p", nil, "base64 encoded public key and allowed ips of a peer (e.g. -p \"<publicKey>,<allowedIPs>\")")
	cmd.PersistentFlags().String("config", "", "config file (default is .wireguard-hub.yaml)")
	cmd.PersistentFlags().String("log-level", "debug", "log level (debug, info, warn, error, fatal)")
	cmd.PersistentFlags().String("hub-address", "", "internal hub IP address")
	cmd.PersistentFlags().Bool("debug-server", false, "start on <hubIP>:8080 the debug server")
	cmd.PersistentFlags().Bool("webui", false, "start on <hubIP>:80 the webui and api")
	cmd.PersistentFlags().String("webui-jwt-secret", "", "secret for JWT authentication")
	cmd.PersistentFlags().String("webui-admin-password-hash", "", "bcrypt hash of the admin password")
	cmd.PersistentFlags().SortFlags = true

	Must(viper.BindPFlag("privateKey", cmd.PersistentFlags().Lookup("private-key")))
	viper.MustBindEnv("privateKey", "PRIVATE_KEY")
	Must(viper.BindPFlag("port", cmd.PersistentFlags().Lookup("port")))
	viper.MustBindEnv("port", "PORT")
	Must(viper.BindPFlag("bindAddress", cmd.PersistentFlags().Lookup("bind-address")))
	viper.MustBindEnv("bindAddress", "BIND_ADDRESS")
	Must(viper.BindPFlag("logLevel", cmd.PersistentFlags().Lookup("log-level")))
	viper.MustBindEnv("logLevel", "LOG_LEVEL")
	Must(viper.BindPFlag("hubAddress", cmd.PersistentFlags().Lookup("hub-address")))
	viper.MustBindEnv("hubAddress", "HUB_ADDRESS")
	Must(viper.BindPFlag("debugServer", cmd.PersistentFlags().Lookup("debug-server")))
	viper.MustBindEnv("debugServer", "DEBUG_SERVER")
	Must(viper.BindPFlag("webui", cmd.PersistentFlags().Lookup("webui")))
	viper.MustBindEnv("webui", "WEBUI")
	Must(viper.BindPFlag("webuiJWTSecret", cmd.PersistentFlags().Lookup("webui-jwt-secret")))
	viper.MustBindEnv("webui-jwt-secret", "WEBUI_JWT_SECRET")
	Must(viper.BindPFlag("webuiAdminPasswordHash", cmd.PersistentFlags().Lookup("webui-admin-password-hash")))
	viper.MustBindEnv("webui-admin-password-hash", "WEBUI_ADMIN_PASSWORD_HASH")
}

type Config struct {
	PrivateKeyHex          string      `yaml:"-"`
	PrivateKey             wgtypes.Key `yaml:"-"`
	Port                   uint16      `yaml:"port"`
	BindAddress            string      `yaml:"bindAddress,omitempty"`
	LogLevel               string      `yaml:"logLevel"`
	HubAddress             string      `yaml:"hubAddress,omitempty"`
	DebugServer            bool        `yaml:"debugServer,omitempty"`
	Webui                  bool        `yaml:"webui,omitempty"`
	WebuiJWTSecret         string      `yaml:"webuiJWTSecret,omitempty"`
	WebuiAdminPasswordHash string      `yaml:"webuiAdminPasswordHash,omitempty"`
	Peers                  []*Peer     `yaml:"peers"`
}

func (c *Config) GetPort() string {
	return strconv.FormatUint(uint64(c.Port), 10)
}

func (c *Config) ResolvedBindAddr() string {
	addr, err := net.ResolveIPAddr("ip", c.BindAddress)
	if err != nil {
		return c.BindAddress
	}
	return addr.String()
}

func (c *Config) GetHubAddress() string {
	return c.HubAddress + "/32"
}

//gocyclo:ignore
func ParseConfig(log *logrus.Logger, cmd *cobra.Command) (*Config, error) {
	privateKey := viper.GetString("privateKey")
	if privateKey == "" {
		return nil, fmt.Errorf("private-key is required")
	}
	wgPrivateKey, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	privateKeyHex, err := ipc.Base64ToHex(wgPrivateKey.String())
	if err != nil {
		// this should never happen
		panic(err)
	}

	port := viper.GetUint16("port")
	bindAddr := viper.GetString("bindAddress")
	log.Infof("listening on %s:%d", bindAddr, port)

	inputPeers := MustGet(cmd.Flags().GetStringArray("peer"))
	for _, s := range os.Environ() {
		if !strings.HasPrefix(s, "PEER_") {
			continue
		}
		_, peer, _ := strings.Cut(s, "=")
		inputPeers = append(inputPeers, peer)
	}
	var configPeers []map[string]string
	err = viper.UnmarshalKey("peers", &configPeers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peers from config: %w", err)
	}
	for _, peer := range configPeers {
		allowedIP := peer["allowedip"]
		if allowedIP == "" {
			allowedIP = peer["allowedips"]
		}
		inputPeers = append(inputPeers, fmt.Sprintf("%s,%s", peer["publickey"], allowedIP))
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
	}

	c := &Config{
		PrivateKeyHex:          privateKeyHex,
		PrivateKey:             wgPrivateKey,
		Port:                   port,
		BindAddress:            bindAddr,
		LogLevel:               viper.GetString("logLevel"),
		HubAddress:             viper.GetString("hubAddress"),
		DebugServer:            viper.GetBool("debugServer"),
		Webui:                  viper.GetBool("webui"),
		WebuiJWTSecret:         viper.GetString("webuiJWTSecret"),
		WebuiAdminPasswordHash: viper.GetString("webuiAdminPasswordHash"),
		Peers:                  peers,
	}

	for _, a := range peers {
		hubOverlap, err := CheckIPOverlap(a.AllowedIP, c.GetHubAddress())
		if err != nil {
			return nil, fmt.Errorf("failed to check ip overlap: %w", err)
		}
		if hubOverlap {
			return nil, fmt.Errorf("hub address overlaps with %s", a)
		}
		for _, b := range peers {
			if a == b {
				continue
			}
			overlap, err := CheckIPOverlap(a.AllowedIP, b.AllowedIP)
			if err != nil {
				return nil, fmt.Errorf("failed to check ip overlap: %w", err)
			}
			if overlap {
				return nil, fmt.Errorf("ip ranges overlap for %s and %s", a, b)
			}
		}
		log.Infof("adding %s", a)
	}
	return c, nil
}
