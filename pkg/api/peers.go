package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/christophwitzko/wg-hub/pkg/config"
	"github.com/christophwitzko/wg-hub/pkg/ipc"
	"github.com/go-chi/chi/v5"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type AnnotatedPeer struct {
	*ipc.Peer
	IsHub       bool `json:"isHub"`
	IsRequester bool `json:"isRequester"`
}

type AnnotatedPeers []*AnnotatedPeer

func (p AnnotatedPeers) GetAllowedIPRanges() []string {
	ipRanges := make([]string, 0, len(p))
	for _, peer := range p {
		ipRanges = append(ipRanges, peer.AllowedIP)
	}
	return ipRanges
}

func (a *API) getPeers(r *http.Request) (AnnotatedPeers, error) {
	devConfig, err := a.dev.IpcGet()
	if err != nil {
		return nil, fmt.Errorf("failed to get ipc operation")
	}
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote address")
	}
	hubIP := a.cfg.GetHubAddress()
	ipcPeers := ipc.ParsePeers(devConfig)
	peers := make(AnnotatedPeers, len(ipcPeers))
	for i, peer := range ipcPeers {
		peers[i] = &AnnotatedPeer{
			Peer:        peer,
			IsHub:       peer.AllowedIP == hubIP,
			IsRequester: peer.AllowedIP == remoteIP+"/32",
		}
	}
	return peers, nil
}

func (a *API) listPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := a.getPeers(r)
	if err != nil {
		a.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.writeJSON(w, peers)
}

//gocyclo:ignore
func (a *API) validateAndAddPeer(w http.ResponseWriter, r *http.Request, publicKey, allowedIP string) (string, string, bool) {
	publicKeyHex, err := ipc.Base64ToHex(publicKey)
	if err != nil {
		a.sendError(w, "failed to decode peer public key", http.StatusBadRequest)
		return "", "", false
	}

	peers, err := a.getPeers(r)
	if err != nil {
		a.sendError(w, err.Error(), http.StatusInternalServerError)
		return "", "", false
	}

	var hubNetwork, allowedIPPrefix string
	if allowedIP == "" {
		allowedIPPrefix, hubNetwork, err = config.GenerateRandomIP(peers.GetAllowedIPRanges())
		if err != nil {
			a.sendError(w, "failed to generate random ip", http.StatusInternalServerError)
			return "", "", false
		}
	} else {
		allowedIPPrefix, err = config.NormalizeAllowedIP(allowedIP)
		if err != nil {
			a.sendError(w, "failed to parse allowed ip", http.StatusBadRequest)
			return "", "", false
		}
		hubNetwork, err = config.FindMinimalNetwork(append(peers.GetAllowedIPRanges(), allowedIPPrefix))
		if err != nil {
			a.sendError(w, "failed to find hub network", http.StatusInternalServerError)
			return "", "", false
		}
	}

	hubOverlap, err := config.CheckIPOverlap(allowedIPPrefix, a.cfg.GetHubAddress())
	if err != nil {
		a.sendError(w, "failed to check ip overlap", http.StatusInternalServerError)
		return "", "", false
	}
	if hubOverlap {
		a.sendError(w, "hub address overlaps with allowed ip", http.StatusBadRequest)
		return "", "", false
	}

	for _, peer := range peers {
		overlap, overlapErr := config.CheckIPOverlap(peer.AllowedIP, allowedIPPrefix)
		if overlapErr != nil {
			a.sendError(w, "failed to check ip overlap", http.StatusInternalServerError)
			return "", "", false
		}
		if overlap {
			a.sendError(w, "allowed ip already in use", http.StatusBadRequest)
			return "", "", false
		}
	}
	addInstruction := fmt.Sprintf(
		"public_key=%s\nreplace_allowed_ips=true\nallowed_ip=%s\n",
		publicKeyHex,
		allowedIPPrefix,
	)
	err = a.dev.IpcSet(addInstruction)
	if err != nil {
		a.sendError(w, "failed to add peer", http.StatusInternalServerError)
		a.log.Errorf("failed to add peer: %v", err)
		return "", "", false
	}
	a.log.Infof("added peer %s (%s)", publicKeyHex, allowedIPPrefix)
	return allowedIPPrefix, hubNetwork, true
}

type AddPeerRequest struct {
	AllowedIP string `json:"allowedIP"`
}

type AddPeerResponse struct {
	AllowedIP  string `json:"allowedIP"`
	HubNetwork string `json:"hubNetwork"`
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
	allowedIP, hubNetwork, ok := a.validateAndAddPeer(w, r, chi.URLParam(r, "*"), req.AllowedIP)
	if !ok {
		return
	}
	a.writeJSON(w, AddPeerResponse{
		AllowedIP:  allowedIP,
		HubNetwork: hubNetwork,
	})
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

type GeneratePeerRequest struct {
	AllowedIP string `json:"allowedIP"`
}

type GeneratePeerResponse struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	AllowedIP  string `json:"allowedIP"`
	HubNetwork string `json:"hubNetwork"`
}

func (a *API) generatePeer(w http.ResponseWriter, r *http.Request) {
	a.ipcMutex.Lock()
	defer a.ipcMutex.Unlock()

	defer r.Body.Close()
	var req GeneratePeerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		a.sendError(w, "failed to decode request", http.StatusBadRequest)
		return
	}
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		a.sendError(w, "failed to generate private key", http.StatusInternalServerError)
		return
	}
	allowedIP, hubNetwork, ok := a.validateAndAddPeer(w, r, privateKey.PublicKey().String(), req.AllowedIP)
	if !ok {
		return
	}
	a.writeJSON(w, GeneratePeerResponse{
		PrivateKey: privateKey.String(),
		PublicKey:  privateKey.PublicKey().String(),
		AllowedIP:  allowedIP,
		HubNetwork: hubNetwork,
	})
}
