# wg-hub

This application acts as a WireGuard® hub server to connect multiple clients (behind a NAT) with each other through a single hub. It runs entirely in the user space and can easily be deployed as a docker container or directly to [Fly.io](https://fly.io) (see [fly.toml](./fly.toml)).

For example, if `Host A` and `Host B` want to communicate with each other, they both connect to the `wg-hub` server.

![](./docs/wg-hub.svg)

`Host A` example WireGuard® config:

```
[Interface]
Address = 192.168.0.1/32
PrivateKey = ...

[Peer]
PublicKey = hub/...
Endpoint = 1.2.3.4:9999
AllowedIPs = 192.168.0.0/24
PersistentKeepalive = 30
```

`Host B` example WireGuard® config:
```
[Interface]
Address = 192.168.0.2/32
PrivateKey = ...

[Peer]
PublicKey = hub/...
Endpoint = 1.2.3.4:9999
AllowedIPs = 192.168.0.0/24
PersistentKeepalive = 30
```

`wireguard-hub.yaml` example config:
```yaml
privateKey: ...
port: 9999
peers:
  - publicKey: hostA/...
    allowedIPs: 192.168.0.1/32
  - publicKey: hostB/...
    allowedIPs: 192.168.0.2/32

```

Start the `wg-hub` instance:
```
$ ./wg-hub --log-level info
INFO[2023-01-20T20:15:10+01:00] using config: wireguard-hub.yaml
INFO[2023-01-20T20:15:10+01:00] listening on :9999
INFO[2023-01-20T20:15:10+01:00] adding peer(876f…29ed): 192.168.0.1/32
INFO[2023-01-20T20:15:10+01:00] adding peer(876f…92de): 192.168.0.2/32

```

Now `Host A` and `Host B` can communicate with each other through the `wg-hub` server.

## Installation

### Binary
```bash
curl -SL https://get-release.xyz/christophwitzko/wg-hub/linux/amd64 -o ./wg-hub && chmod +x ./wg-hub
```

### Docker
```bash
docker run -it --rm \
  -e PRIVATE_KEY="..."
  -e PORT=9999 \
  -e PEER_1="hostA/...,192.168.0.1/32" \
  -e PEER_2="hostB/...,192.168.0.2/32" \
  -p 9999:9999/udp \
  ghcr.io/christophwitzko/wg-hub
```

## Todo

- [ ] Allow to dynamically add or remove peers
- [ ] Web UI or dashboard?


## Legal
[WireGuard](https://www.wireguard.com/) is a registered trademark of Jason A. Donenfeld.
