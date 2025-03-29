---
description: Configuration related to the WebRTC and Networking in Neko.
---

import { Def, Opt } from '@site/src/components/Anchor';

# WebRTC Configuration

This page describes how to configure WebRTC settings inside neko.

Neko uses WebRTC with the [Pion](https://github.com/pion/webrtc) library to establish a peer-to-peer connection between the client and the server. This connection is used to stream audio, video, and data bidirectionally between the client and the server.

## ICE Setup {#ice}

ICE, which stands for Interactive Connectivity Establishment, is a protocol used to find the best path to connect peers, such as a client and a server. It helps discover the public IP addresses and ports of both parties to establish a direct connection. ICE candidates, which contain this information, are exchanged through a signaling server to facilitate the connection process.

### ICE Trickle {#icetrickle}

ICE Trickle is a feature that allows ICE candidates to be sent as they are discovered, rather than waiting for all candidates to be discovered before sending them. It means that the ICE connection can be established faster as the server can start connecting to the client as soon as it has a few ICE candidates and doesn't have to wait for all of them to be discovered.

```yaml title="config.yaml"
webrtc:
  icetrickle: false
```

### ICE Lite {#icelite}

ICE Lite is a minimal implementation of the ICE protocol intended for servers running on a public IP address. It is not enabled by default to allow more complex ICE configurations out of the box.

```yaml title="config.yaml"
webrtc:
  icelite: false
```

:::info
When using ICE Servers, ICE Lite must be disabled.
:::

### ICE Servers {#iceservers}

ICE servers are used to establish a connection between the client and the server. There are two types of ICE servers:

- [STUN](https://en.wikipedia.org/wiki/STUN): A STUN server is used to discover the public IP address of the client. This is used to establish a direct connection between the client and the server.
- [TURN](https://en.wikipedia.org/wiki/Traversal_Using_Relays_around_NAT): A TURN server is used to relay data between the client and the server if a direct connection cannot be established.

The configuration of a single [ICE server](https://developer.mozilla.org/en-US/docs/Web/API/RTCPeerConnection/RTCPeerConnection#iceservers) is defined by the following fields:

| Field                               | Description | Type |
|-------------------------------------|-------------|------|
| <Def id="iceservers.urls" />       | List of URLs of the ICE server, if the same server is available on multiple URLs with the same credentials, they can be listed here. | `string[]` |
| <Def id="iceservers.username" />   | Username used to authenticate with the ICE server, if the server requires authentication. | `string` |
| <Def id="iceservers.credential" /> | Credential used to authenticate with the ICE server, if the server requires authentication. | `string` |

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="yaml" label="YAML" default>
  
    ```yaml title="Example of multiple ICE servers in YAML"
    - urls: "turn:<MY-COTURN-SERVER>:3478"
      username: "neko"
      credential: "neko"
    - urls: "stun:stun.l.google.com:19302"
    ```

  </TabItem>
  <TabItem value="json" label="JSON">

    ```json title="Example of multiple ICE servers in JSON"
    [
      {
        "urls": "turn:<MY-COTURN-SERVER>:3478",
        "username": "neko",
        "credential": "neko"
      },
      {
        "urls": "stun:stun.l.google.com:19302"
      }
    ]
    ```

    :::tip
      You can specify the ICE servers as a JSON string in the `docker-compose.yaml` file using the `NEKO_WEBRTC_ICESERVERS_FRONTEND` and `NEKO_WEBRTC_ICESERVERS_BACKEND` environment variables.

      ```yaml title="docker-compose.yaml"
        NEKO_WEBRTC_ICESERVERS_FRONTEND: |
          [{
            "urls": [ "turn:<MY-COTURN-SERVER>:3478" ],
            "username": "neko",
            "credential": "neko"
          },{
            "urls": [ "stun:stun.nextcloud.com:3478" ]
          }]
      ```
    :::

  </TabItem>
</Tabs>

The ICE servers are divided into two groups:

```yaml title="config.yaml"
webrtc:
  iceservers:
    frontend:
    # List of ICE Server configurations as described above
    - urls: "stun:stun.l.google.com:19302"
    backend:
    # List of ICE Server configurations as described above
    - urls: "stun:stun.l.google.com:19302"
```

- <Def id="iceservers.frontend" /> - ICE servers that are sent to the client and used to establish a connection between the client and the server.
- <Def id="iceservers.backend" /> - ICE servers that are used by the server to gather ICE candidates. They might contain private IP addresses or other sensitive information that should not be sent to the client.

<details>
<summary>Example with Coturn server in Docker Compose</summary>

```yaml title="docker-compose.yaml"
services:
  coturn:
    image: 'coturn/coturn:latest'
    network_mode: "host"
    command: |
      -n
      --realm=localhost
      --fingerprint
      --listening-ip=0.0.0.0
      --external-ip=<MY-COTURN-SERVER>
      --listening-port=3478
      --min-port=49160
      --max-port=49200
      --log-file=stdout
      --user=neko:neko
      --lt-cred-mech
```

Replace `<MY-COTURN-SERVER>` with your LAN or Public IP address, and allow ports `49160-49200/udp` and `3478/tcp`. The `--user` flag is used to specify the username and password for the TURN server. The `--lt-cred-mech` flag is used to enable the long-term credentials mechanism for authentication. More information about the Coturn server can be found [here](https://github.com/coturn/coturn).

</details>

## Network Setup {#network}

Since WebRTC is a peer-to-peer protocol that requires a direct connection between the client and the server. This can be achieved by:

- Using a public IP address for the server (or at least reachable from the client if deployed on a private network).
- Using a [TURN server](#iceservers) to relay data between the client and the server if a direct connection cannot be established.

All specified ports along with the server's IP address will be sent to the client in ICE candidates to establish a connection. Therefore, it is important to ensure that the specified ports are open on the server's firewall, are not remapped to different ports, and are reachable from the client.

:::danger Remember
WebRTC does not use the HTTP protocol, therefore it is not possible to use nginx or other reverse proxies to forward the WebRTC traffic. If you only have exposed port `443` on your server, you must expose as well the WebRTC ports or use a TURN server.
:::

There exist two types of connections:

- [Ephemeral UDP port range](#epr): The range of UDP ports that the server uses to establish a connection with the client. Every time a new connection is established, a new port from this range is used. This range should be open on the server's firewall.
- [UDP/TCP multiplexing](#mux): The server can use a single port for multiple connections. This port should be open on the server's firewall.

### Ephemeral UDP port range {#epr}

The ephemeral UDP port range can be configured using the following configuration:

```yaml title="config.yaml"
webrtc:
  epr: "59000-59100"
```

The range `59000-59100` contains 101 ports, which should be open on the server's firewall. The server uses these ports to establish a connection with the client. You can specify a different range of ports if needed, with fewer or more ports, depending on the number of simultaneous connections you expect.

:::tip
You can specify the ephemeral UDP port range as an environment variable in the `docker-compose.yaml` file using the `NEKO_WEBRTC_EPR` environment variable. When using docker, make sure to expose the ports in the `docker-compose.yaml`.

```yaml title="docker-compose.yaml"
environment:
  NEKO_WEBRTC_EPR: "59000-59100"
ports:
  - "59000-59100:59000-59100/udp"
```

It is important to expose the same ports to the host machine, without any remapping e.g. `49000-49100:59000-59100/udp` instead of `59000-59100:59000-59100/udp`.
:::

### UDP/TCP multiplexing {#mux}

The UDP/TCP multiplexing port can be configured using the following configuration:

```yaml title="config.yaml"
webrtc:
  udpmux: 59000
  tcpmux: 59000
```

- <Def id="udpmux" /> - The port used for UDP connections.
- <Def id="tcpmux" /> - The port used for TCP connections.

The server uses only port `59000` for both UDP and TCP connections. This port should be open on the server's firewall. You can specify a different port if needed, or specify only one of the two protocols. UDP is generally better for latency, but some networks block UDP so it is good to have TCP available as a fallback.

:::tip
You can specify the UDP/TCP multiplexing port as an environment variable in the `docker-compose.yaml` file using the `NEKO_WEBRTC_TCPMUX` and `NEKO_WEBRTC_UDPMUX` environment variables. When using docker, make sure to expose the ports in the `docker-compose.yaml`.

```yaml title="docker-compose.yaml"
environment:
  NEKO_WEBRTC_UDPMUX: "59000"
  NEKO_WEBRTC_TCPMUX: "59000"
ports:
  - "59000:59000/udp"
  - "59000:59000/tcp"
```

It is important to expose the same ports to the host machine, without any remapping e.g. `49000:59000/udp` instead of `59000:59000/udp`.
:::

### Server IP Address {#ip}

The server IP address is sent to the client in ICE candidates so that the client can establish a connection with the server. By default, the server IP address is automatically resolved by the server to the public IP address of the server. If the server is behind a NAT, you want to specify a different IP address or use neko only in a local network, you can specify the server IP address manually.

#### NAT 1-to-1 {#nat1to1}

```yaml title="config.yaml"
webrtc:
  nat1to1:
  # IPv4 address of the server
  - 10.10.0.5
  # IPv6 address of the server
  - 2001:db8:85a3::8a2e:370:7334
```

Currently, only one IPv4 and one IPv6 address can be specified. Therefore if you want to access your instance from both local and public networks, your router must support [NAT loopback (hairpinning)](https://en.wikipedia.org/wiki/Network_address_translation#NAT_hairpinning).

:::tip
You can specify the server IP address as an environment variable in the `docker-compose.yaml` file using the `NEKO_WEBRTC_NAT1TO1` environment variable.

```yaml title="docker-compose.yaml"
environment:
  NEKO_WEBRTC_NAT1TO1: "10.10.0.5"
```

If you want to specify also an IPv6 address, use whitespace to separate the addresses.

```yaml title="docker-compose.yaml"
environment:
  NEKO_WEBRTC_NAT1TO1: "10.10.0.5 2001:db8:85a3::8a2e:370:7334"
```
:::

#### IP Retrieval URL {#ip_retrieval_url}

If you do not specify the server IP address, the server will try to resolve the public IP address of the server automatically.

```yaml title="config.yaml"
webrtc:
  ip_retrieval_url: "https://checkip.amazonaws.com"
```
The server will send an HTTP GET request to the specified URL to retrieve the public IP address of the server.

## Bandwidth Estimator {#estimator}

:::danger
The bandwidth estimator is an experimental feature and might not work as expected.
:::

The bandwidth estimator is a feature that allows the server to estimate the available bandwidth between the client and the server. It is used to switch between different video qualities based on the available bandwidth. The bandwidth estimator is disabled by default.

```yaml title="config.yaml"
webrtc:
  estimator:
    # Whether to enable the bandwidth estimator
    enabled: false
    # Whether the bandwidth estimator is passive - only used for logging and not for actual decisions
    passive: false
    # Enable debug logging for the bandwidth estimator (will print the current state and decisions)
    debug: false
    # Initial bitrate for the bandwidth estimator to start with (in bps)
    initial_bitrate: 1000000
    # How often to read and process bandwidth estimation reports
    read_interval: "2s"
    # How long to wait for a stable connection (upward or neutral trend) before upgrading
    stable_duration: "12s"
    # How long to wait for a stalled connection (neutral trend with low bandwidth) before downgrading
    unstable_duration: "6s"
    # How long to wait for stalled bandwidth estimation before downgrading
    stalled_duration: "24s"
    # How long to wait before downgrading again after the previous downgrade
    downgrade_backoff: "10s"
    # How long to wait before upgrading again after the previous upgrade
    upgrade_backoff: "5s"
    # How much bigger the difference between estimated and stream bitrate must be to trigger a change
    diff_threshold: 0.15
```
