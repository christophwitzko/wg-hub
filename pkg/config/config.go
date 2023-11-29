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
	cmd.PersistentFlags().Bool("api-server", false, "start on <hubIP>:80 the api server")
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
	Must(viper.BindPFlag("apiServer", cmd.PersistentFlags().Lookup("api-server")))
	viper.MustBindEnv("apiServer", "API_SERVER")
}

type Config struct {
	PrivateKeyHex string
	PrivateKey    wgtypes.Key
	Port          uint16
	BindAddress   string
	Peers         []*Peer
	HubAddress    string
	DebugServer   bool
	APIServer     bool
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
		PrivateKeyHex: privateKeyHex,
		PrivateKey:    wgPrivateKey,
		Port:          port,
		BindAddress:   bindAddr,
		Peers:         peers,
		HubAddress:    viper.GetString("hubAddress"),
		DebugServer:   viper.GetBool("debugServer"),
		APIServer:     viper.GetBool("apiServer"),
	}, nil
}
