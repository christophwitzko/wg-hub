package api

import "net/http"

type HubInfo struct {
	PublicKey string `json:"publicKey"`
	Port      uint16 `json:"port"`
}

func (a *API) getHubInfo(w http.ResponseWriter, _ *http.Request) {
	hubInfo := HubInfo{
		PublicKey: a.cfg.PrivateKey.PublicKey().String(),
		Port:      a.cfg.Port,
	}
	a.writeJSON(w, hubInfo)
}
