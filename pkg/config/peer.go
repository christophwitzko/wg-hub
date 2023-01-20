package config

import (
	"fmt"
	"net/netip"
	"strings"
)

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
	publicKeyHex, err := Base64ToHex(p.PublicKey)
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
