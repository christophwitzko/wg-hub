package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/christophwitzko/wg-hub/pkg/config"
	"github.com/christophwitzko/wg-hub/pkg/ipc"
	"github.com/go-chi/chi/v5"
)

type Peer struct {
	*ipc.Peer
	IsHub       bool `json:"isHub"`
	IsRequester bool `json:"isRequester"`
}

func (a *API) listPeers(w http.ResponseWriter, r *http.Request) {
	devConfig, err := a.dev.IpcGet()
	if err != nil {
		a.sendError(w, "failed to get ipc operation", http.StatusInternalServerError)
		return
	}
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		a.sendError(w, "failed to parse remote address", http.StatusInternalServerError)
		return
	}
	hubIP := a.cfg.GetHubAddress()
	ipcPeers := ipc.ParsePeers(devConfig)
	peers := make([]*Peer, len(ipcPeers))
	for i, peer := range ipcPeers {
		peers[i] = &Peer{
			Peer:        peer,
			IsHub:       peer.AllowedIP == hubIP,
			IsRequester: peer.AllowedIP == remoteIP+"/32",
		}
	}

	a.writeJSON(w, peers)
}

type AddPeerRequest struct {
	PublicKey string `json:"publicKey"`
	AllowedIP string `json:"allowedIP"`
}

func (a *API) addPeer(w http.ResponseWriter, r *http.Request) {
	a.ipcMutex.Lock()
	defer a.ipcMutex.Unlock()

	defer r.Body.Close()
	var req AddPeerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		a.sendError(w, "failed to decode request", http.StatusBadRequest)
		return
	}
	if req.PublicKey == "" || req.AllowedIP == "" {
		a.sendError(w, "PublicKey and AllowedIP are required", http.StatusBadRequest)
		return
	}
	publicKeyHex, err := ipc.Base64ToHex(req.PublicKey)
	if err != nil {
		a.sendError(w, "failed to decode peer public key", http.StatusBadRequest)
		return
	}
	allowedIPPrefix, err := config.NormalizeAllowedIP(req.AllowedIP)
	if err != nil {
		a.sendError(w, "failed to parse allowed ip", http.StatusBadRequest)
		a.log.Errorf("failed to parse allowed ip: %v", err)
		return
	}
	// TODO: check ip overlap
	addInstruction := fmt.Sprintf(
		"public_key=%s\nreplace_allowed_ips=true\nallowed_ip=%s\n",
		publicKeyHex,
		allowedIPPrefix,
	)
	err = a.dev.IpcSet(addInstruction)
	if err != nil {
		a.sendError(w, "failed to add peer", http.StatusInternalServerError)
		a.log.Errorf("failed to add peer: %v", err)
		return
	}
	a.log.Infof("added peer %s (%s)", publicKeyHex, allowedIPPrefix)
	a.writeJSON(w, map[string]string{"status": "ok"})
}

func (a *API) removePeer(w http.ResponseWriter, r *http.Request) {
	a.ipcMutex.Lock()
	defer a.ipcMutex.Unlock()

	peerPublicKeyHex, err := ipc.Base64ToHex(chi.URLParam(r, "*"))
	if err != nil {
		a.sendError(w, "failed to decode peer public key", http.StatusBadRequest)
		return
	}
	deleteInstruction := fmt.Sprintf("public_key=%s\nremove=true\n", peerPublicKeyHex)
	err = a.dev.IpcSet(deleteInstruction)
	if err != nil {
		a.sendError(w, "failed to remove peer", http.StatusInternalServerError)
		a.log.Errorf("failed to remove peer: %v", err)
		return
	}
	a.log.Infof("removed peer %s", peerPublicKeyHex)
	a.writeJSON(w, map[string]string{"status": "ok"})
}
