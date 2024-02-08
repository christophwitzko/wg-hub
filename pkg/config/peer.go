package config

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/christophwitzko/wg-hub/pkg/ipc"
)

type Peer struct {
	PublicKey    string `yaml:"publicKey"`
	PublicKeyHex string `yaml:"-"`
	AllowedIP    string `yaml:"allowedIP"`
}

func NormalizeAllowedIP(ip string) (string, error) {
	// add subnet mask if not present
	if !strings.Contains(ip, "/") {
		ip = fmt.Sprintf("%s/32", ip)
	}

	// check if ip is valid
	ipPrefix, err := netip.ParsePrefix(ip)
	if err != nil {
		return "", fmt.Errorf("failed to parse allowed ip: %w", err)
	}
	return ipPrefix.String(), nil
}

func CheckIPOverlap(a, b string) (bool, error) {
	aNet, err := netip.ParsePrefix(a)
	if err != nil {
		return false, err
	}
	bNet, err := netip.ParsePrefix(b)
	if err != nil {
		return false, err
	}
	return aNet.Overlaps(bNet), nil
}

func NewPeer(peerConfig string) (*Peer, error) {
	publicKey, ip, ok := strings.Cut(peerConfig, ",")
	if !ok {
		return nil, fmt.Errorf("failed to parse peer config: %s", peerConfig)
	}
	p := &Peer{
		PublicKey: publicKey,
	}
	publicKeyHex, err := ipc.Base64ToHex(p.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode peer public key: %w", err)
	}
	p.PublicKeyHex = publicKeyHex

	ipPrefix, err := NormalizeAllowedIP(ip)
	if err != nil {
		return nil, err
	}
	p.AllowedIP = ipPrefix
	return p, nil
}

func (p *Peer) String() string {
	return fmt.Sprintf("peer(%s…%s): %s", p.PublicKeyHex[:4], p.PublicKeyHex[len(p.PublicKeyHex)-4:], p.AllowedIP)
}
