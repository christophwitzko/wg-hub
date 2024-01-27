package ipc

import (
	"encoding/base64"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

func Base64ToHex(b64 string) (string, error) {
	decKey, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decKey), nil
}

func HexToBase64(hexKey string) (string, error) {
	decKey, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(decKey), nil
}

type Peer struct {
	PublicKey     string `json:"publicKey"`
	AllowedIP     string `json:"allowedIP"`
	Endpoint      string `json:"endpoint"`
	LastHandshake uint64 `json:"lastHandshake"`
	TxBytes       uint64 `json:"txBytes"`
	RxBytes       uint64 `json:"rxBytes"`
}

func ParsePeers(config string) []*Peer {
	var currentPeer *Peer
	var peers []*Peer
	for _, line := range strings.Split(config, "\n") {
		if strings.HasPrefix(line, "public_key") {
			if currentPeer != nil {
				peers = append(peers, currentPeer)
			}
			currentPeer = &Peer{}
			_, publicKey, _ := strings.Cut(line, "=")
			currentPeer.PublicKey, _ = HexToBase64(publicKey)
			continue
		}
		if strings.HasPrefix(line, "allowed_ip") {
			_, allowedIP, _ := strings.Cut(line, "=")
			currentPeer.AllowedIP = allowedIP
			continue
		}
		if strings.HasPrefix(line, "endpoint") {
			_, endpoint, _ := strings.Cut(line, "=")
			currentPeer.Endpoint = endpoint
			continue
		}
		if strings.HasPrefix(line, "last_handshake_time_sec") {
			_, lastHandshake, _ := strings.Cut(line, "=")
			currentPeer.LastHandshake, _ = strconv.ParseUint(lastHandshake, 10, 64)
			continue
		}
		if strings.HasPrefix(line, "tx_bytes") {
			_, txBytes, _ := strings.Cut(line, "=")
			currentPeer.TxBytes, _ = strconv.ParseUint(txBytes, 10, 64)
			continue
		}
		if strings.HasPrefix(line, "rx_bytes") {
			_, rxBytes, _ := strings.Cut(line, "=")
			currentPeer.RxBytes, _ = strconv.ParseUint(rxBytes, 10, 64)
			continue
		}
	}
	if currentPeer != nil {
		peers = append(peers, currentPeer)
	}
	// sort by public key
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].PublicKey < peers[j].PublicKey
	})
	return peers
}
