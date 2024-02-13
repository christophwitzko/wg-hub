package api

import (
	"net/http"

	"github.com/christophwitzko/wg-hub/pkg/config"
)

type HubInfo struct {
	PublicKey    string `json:"publicKey"`
	Port         uint16 `json:"port"`
	HubNetwork   string `json:"hubNetwork"`
	NextRandomIP string `json:"nextRandomIP"`
}

func (a *API) getHubInfo(w http.ResponseWriter, r *http.Request) {
	peers, err := a.getPeers(r)
	if err != nil {
		a.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ipRanges := make([]string, 0, len(peers))
	for _, peer := range peers {
		ipRanges = append(ipRanges, peer.AllowedIP)
	}
	randomIP, hubNet, err := config.GenerateRandomIP(ipRanges)
	if err != nil {
		a.sendError(w, "failed to find hub network", http.StatusInternalServerError)
		return
	}
	hubInfo := HubInfo{
		PublicKey:    a.cfg.PrivateKey.PublicKey().String(),
		Port:         a.cfg.Port,
		HubNetwork:   hubNet,
		NextRandomIP: randomIP,
	}
	a.writeJSON(w, hubInfo)
}
