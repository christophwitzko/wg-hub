package config

import (
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/christophwitzko/wg-hub/pkg/ipc"
)

var (
	globalRand      = rand.New(rand.NewSource(time.Now().UnixNano()))
	globalRandMutex = sync.Mutex{}
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

func equalBytes(a, b []byte) int {
	cnt := 0
	for i := 0; i < 4; i++ {
		if a[i] == b[i] {
			cnt++
		} else {
			break
		}
	}
	return cnt
}

func findMinimalIPNet(ipRanges []string) (*net.IPNet, error) {
	if len(ipRanges) == 0 {
		return nil, fmt.Errorf("more than one ip range required")
	}
	_, minNet, err := net.ParseCIDR(ipRanges[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse ip range: %w", err)
	}
	if len(minNet.IP) != 4 {
		return nil, fmt.Errorf("only ipv4 supported")
	}
	for _, ipRange := range ipRanges[1:] {
		_, ipNet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ip range: %w", err)
		}
		if len(ipNet.IP) != 4 {
			return nil, fmt.Errorf("only ipv4 supported")
		}
		equalAddrParts := equalBytes(minNet.IP, ipNet.IP)
		equalMaskParts := equalBytes(minNet.Mask, ipNet.Mask)
		newMask := min(equalAddrParts, equalMaskParts)
		minNet.Mask = net.CIDRMask(newMask*8, 32)
	}
	minNet.IP = minNet.IP.Mask(minNet.Mask)
	return minNet, nil
}

func randomIP(rnd *rand.Rand, ipNet *net.IPNet) net.IP {
	ip := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		if ipNet.Mask[i] == 255 {
			ip[i] = ipNet.IP[i]
		} else {
			rndVal := byte(rnd.Intn(256))
			if i != 3 {
				ip[i] = rndVal
				continue
			}
			// do not use network or broadcast address
			if rndVal == 0 && (ipNet.Mask[2] == 255 || ip[2] == 0) {
				rndVal = 1
			}
			if rndVal == 255 && (ipNet.Mask[2] == 255 || ip[2] == 255) {
				rndVal = 254
			}
			ip[i] = rndVal
		}
	}
	return ip
}

func generateRandomIP(rnd *rand.Rand, minNet *net.IPNet, ipRanges []string) (string, string, error) {
	for c := 0; c < 10000; c++ {
		newIP := randomIP(rnd, minNet).String() + "/32"
		overlapCheck := func(ip string) bool {
			overlap, _ := CheckIPOverlap(ip, newIP)
			return overlap
		}
		if slices.ContainsFunc(ipRanges, overlapCheck) {
			// new ip overlaps with existing ip ranges
			continue
		}
		return newIP, minNet.String(), nil
	}
	return "", minNet.String(), nil
}

func FindMinimalNetwork(ipRanges []string) (string, error) {
	minNet, err := findMinimalIPNet(ipRanges)
	if err != nil {
		return "", err
	}
	return minNet.String(), nil
}

func GenerateRandomIP(ipRanges []string) (string, string, error) {
	globalRandMutex.Lock()
	defer globalRandMutex.Unlock()

	minNet, err := findMinimalIPNet(ipRanges)
	if err != nil {
		return "", "", err
	}
	return generateRandomIP(globalRand, minNet, ipRanges)
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
