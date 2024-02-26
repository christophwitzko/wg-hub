package api

import (
	"net/http"

	"github.com/christophwitzko/wg-hub/pkg/config"
)

type HubInfo struct {
	PublicKey    string `json:"publicKey"`
	Port         uint16 `json:"port"`
	HubNetwork   string `json:"hubNetwork"`
	RandomFreeIP string `json:"randomFreeIP"`
	ExternalIP   string `json:"externalIP"`
}

func (a *API) getHubInfo(w http.ResponseWriter, r *http.Request) {
	peers, err := a.getPeers(r)
	if err != nil {
		a.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	randomIP, hubNet, err := config.GenerateRandomIP(peers.GetAllowedIPRanges())
	if err != nil {
		a.sendError(w, "failed to find hub network", http.StatusInternalServerError)
		return
	}
	hubInfo := HubInfo{
		PublicKey:    a.cfg.PrivateKey.PublicKey().String(),
		Port:         a.cfg.Port,
		HubNetwork:   hubNet,
		RandomFreeIP: randomIP,
		ExternalIP:   a.cfg.GetExternalAddress(),
	}
	a.writeJSON(w, hubInfo)
}
