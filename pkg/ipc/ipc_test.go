package ipc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const ipcGetTestData = `
private_key=xxxxx
listen_port=9999
public_key=875ff02792a8417d5f433436e6b9476f5f308001bcd2f9032ed9fc07ba71396d
preshared_key=0000000000000000000000000000000000000000000000000000000000000000
protocol_version=1
endpoint=127.0.0.1:49388
last_handshake_time_sec=1
last_handshake_time_nsec=1
tx_bytes=2828
rx_bytes=3092
persistent_keepalive_interval=0
allowed_ip=192.168.0.1/32
public_key=876fcf026120a0844b6010c34cbddd64a00e68b121bb8dbbd2f94d59714c4b2b
preshared_key=0000000000000000000000000000000000000000000000000000000000000000
protocol_version=1
last_handshake_time_sec=0
last_handshake_time_nsec=0
tx_bytes=0
rx_bytes=0
persistent_keepalive_interval=0
allowed_ip=192.168.0.2/32
public_key=3dc79647223bcb6cc8df16e2e5be5dd22a097430359e667d475c9df69c245964
preshared_key=0000000000000000000000000000000000000000000000000000000000000000
protocol_version=1
endpoint=127.0.0.1:55900
last_handshake_time_sec=3
last_handshake_time_nsec=3
tx_bytes=460
rx_bytes=372
persistent_keepalive_interval=0
allowed_ip=192.168.0.254/32
`

func TestParsePeers(t *testing.T) {
	peers := ParsePeers(ipcGetTestData)
	require.Equal(t, 3, len(peers))
	expectedPeers := []*Peer{
		{
			PublicKey:     "PceWRyI7y2zI3xbi5b5d0ioJdDA1nmZ9R1yd9pwkWWQ=",
			AllowedIP:     "192.168.0.254/32",
			Endpoint:      "127.0.0.1:55900",
			LastHandshake: 3,
			TxBytes:       460,
			RxBytes:       372,
		},
		{
			PublicKey:     "h1/wJ5KoQX1fQzQ25rlHb18wgAG80vkDLtn8B7pxOW0=",
			AllowedIP:     "192.168.0.1/32",
			Endpoint:      "127.0.0.1:49388",
			LastHandshake: 1,
			TxBytes:       2828,
			RxBytes:       3092,
		},
		{
			PublicKey: "h2/PAmEgoIRLYBDDTL3dZKAOaLEhu4270vlNWXFMSys=",
			AllowedIP: "192.168.0.2/32",
		},
	}
	require.Equal(t, expectedPeers, peers)
}
