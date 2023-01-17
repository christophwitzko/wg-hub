package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	BindAddr   string
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
	bindAddr := viper.GetString("bindAddress")
	log.Infof("listening on %s:%d", bindAddr, port)

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
		BindAddr:   bindAddr,
		Peers:      peers,
	}, nil
}
