package api

import (
	"net/http"
	"strings"

	"github.com/christophwitzko/wg-hub/pkg/config"
	"github.com/christophwitzko/wg-hub/pkg/ipc"
	"gopkg.in/yaml.v3"
)

func (a *API) getConfig(w http.ResponseWriter, _ *http.Request) {
	devConfig, err := a.dev.IpcGet()
	if err != nil {
		a.sendError(w, "failed to get ipc operation", http.StatusInternalServerError)
		return
	}
	peers := ipc.ParsePeers(devConfig)
	currentPeers := make([]*config.Peer, 0)
	for _, peer := range peers {
		if peer.AllowedIP == a.cfg.HubAddress+"/32" {
			continue
		}
		currentPeers = append(currentPeers, &config.Peer{
			PublicKey: peer.PublicKey,
			AllowedIP: peer.AllowedIP,
		})
	}
	// create a new config with the current config and the peers
	cfgData, err := yaml.Marshal(config.Config{
		Port:                   a.cfg.Port,
		BindAddress:            a.cfg.BindAddress,
		LogLevel:               a.cfg.LogLevel,
		HubAddress:             a.cfg.HubAddress,
		DebugServer:            a.cfg.DebugServer,
		Webui:                  a.cfg.Webui,
		WebuiJWTSecret:         "<redacted>",
		WebuiAdminPasswordHash: a.cfg.WebuiAdminPasswordHash,
		Peers:                  currentPeers,
	})
	if err != nil {
		a.sendError(w, "failed to marshal config", http.StatusInternalServerError)
		return
	}
	var cfgStr strings.Builder
	cfgStr.WriteString("privateKey: <redacted>\n")
	cfgStr.Write(cfgData)
	a.writeJSON(w, map[string]string{"config": cfgStr.String()})
}
