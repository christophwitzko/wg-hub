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

## Webui
To enable the Webui and dynamically manage peers the following config options need to be set.
```yaml
hubAddress: 192.168.0.254
webui: true
webuiJWTSecret: supersecure # random secret to sign JWT tokens
# use `caddy hash-password` or `htpasswd -nB admin` to generate the hash
webuiAdminPasswordHash: $2a$14$hTHK6KAynSb7tWknK4CvUum2eFVHIDSzbOuOlgDeP4bQW91ujnlli #admin
```
The Webui will be running on the `hubAddress` and port 80 (e.g. http://192.168.0.254).

## API

The API will be served under the same endpoint the Webui is running (`hubAddress`). All API request (except `POST /api/auth`) require a valid `Authorization` header with the value `Bearer <token>`. The token can be obtained by sending a `POST /api/auth` request with the `username` and `password` in the request body.

### POST /api/auth
<details>
<summary>Example request body</summary>

```json
{
  "username": "admin",
  "password": "admin"
}
```
</details>

<details>
<summary>Example response body</summary>

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3OTM3MTI2NTgsImlhdCI6MTcwNzMxMjY1OCwidXNlcm5hbWUiOiJhZG1pbiJ9.fVvahKZPJ2LUZE_dOIYQ6toYFN4x_r295jxINqlXY88"
}
```
</details>

### GET /api/auth
<details>
<summary>Example response body</summary>

```json
{
  "exp": "2026-11-03T13:30:58Z",
  "iat": "2024-02-07T13:30:58Z",
  "username": "admin"
}
```
</details>

### GET /api/peers
<details>
<summary>Example response body</summary>

```json
[
  {
    "publicKey": "ZbSHDrKwqmsQKpO5T6lOY/iipbcJpT4DPXTHGsLaGUU=",
    "allowedIP": "192.168.0.254/32",
    "endpoint": "127.0.0.1:64404",
    "lastHandshake": 1707312755,
    "txBytes": 4696,
    "rxBytes": 4968,
    "isHub": true,
    "isRequester": false
  },
  {
    "publicKey": "h1/wJ5KoQX1fQzQ25rlHb18wgAG80vkDLtn8B7pxOW0=",
    "allowedIP": "192.168.0.1/32",
    "endpoint": "127.0.0.1:58646",
    "lastHandshake": 1707312760,
    "txBytes": 4152,
    "rxBytes": 5640,
    "isHub": false,
    "isRequester": true
  },
  {
    "publicKey": "h2/PAmEgoIRLYBDDTL3dZKAOaLEhu4270vlNWXFMSys=",
    "allowedIP": "192.168.0.2/32",
    "endpoint": "",
    "lastHandshake": 0,
    "txBytes": 0,
    "rxBytes": 0,
    "isHub": false,
    "isRequester": false
  }
]
```
</details>

### GET /api/config
<details>
<summary>Example response body</summary>

```json
{
  "config": "privateKey: <redacted>\nport: 9999\nlogLevel: debug\nhubAddress: 192.168.0.254\ndebugServer: true\nwebui: true\nwebuiJWTSecret: <redacted>\nwebuiAdminPasswordHash: $2a$14$hTHK6KAynSb7tWknK4CvUum2eFVHIDSzbOuOlgDeP4bQW91ujnlli\npeers:\n    - publicKey: h1/wJ5KoQX1fQzQ25rlHb18wgAG80vkDLtn8B7pxOW0=\n      allowedIP: 192.168.0.1/32\n    - publicKey: h2/PAmEgoIRLYBDDTL3dZKAOaLEhu4270vlNWXFMSys=\n      allowedIP: 192.168.0.2/32\n"
}
```
</details>

### POST /api/peers
<details>
<summary>Example requeset body</summary>

```json
{
  "publicKey": "h2/PAmEgoIRLYBDDTL3dZKAOaLEhu4270vlNWXFMSys=",
  "allowedIP": "1.2.3.4/32"
}
```
</details>
<details>
<summary>Example response body</summary>

```json
{
  "status": "ok"
}
```
</details>

### DELETE /api/peers/:publicKey
<details>
<summary>Example response body</summary>

```json
{
  "status": "ok"
}
```
</details>

## Legal
[WireGuard](https://www.wireguard.com/) is a registered trademark of Jason A. Donenfeld.
