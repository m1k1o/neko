---
sidebar_position: 4
---

# WebRTC

This page describes how to configure WebRTC settings inside neko.

Neko uses WebRTC with the [Pion](https://github.com/pion/webrtc) library to establish a peer-to-peer connection between the client and the server. This connection is used to stream audio, video, and data bidirectionally between the client and the server.

## ICE Trickle

ICE Trickle is a feature that allows ICE candidates to be sent as they are discovered, rather than waiting for all candidates to be discovered before sending them. It means that the ICE connection can be established faster as the server can start connecting to the client as soon as it has a few ICE candidates and doesn't have to wait for all of them to be discovered.

```yaml title="config.yaml"
webrtc:
  icetrickle: false
```

## ICE Servers

ICE servers are used to establish a connection between the client and the server. There are two types of ICE servers:

- [STUN](https://en.wikipedia.org/wiki/STUN): A STUN server is used to discover the public IP address of the client. This is used to establish a direct connection between the client and the server.
- [TURN](https://en.wikipedia.org/wiki/Traversal_Using_Relays_around_NAT): A TURN server is used to relay data between the client and the server if a direct connection cannot be established.

The configuration of a single [ICE server](https://developer.mozilla.org/en-US/docs/Web/API/RTCPeerConnection/RTCPeerConnection#iceservers) is defined by the following fields:

| Field                      | Description | Type |
|----------------------------|-------------|------|
| `urls`                     | List of URLs of the ICE server, if the same server is available on multiple URLs with the same credentials, they can be listed here. | `string[]` |
| `username`                 | Username used to authenticate with the ICE server, if the server requires authentication. | `string` |
| `credential`               | Credential used to authenticate with the ICE server, if the server requires authentication. | `string` |

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

- `frontend`: ICE servers that are sent to the client and used to establish a connection between the client and the server.
- `backend`: ICE servers that are used by the server to gather ICE candidates. They might contain private IP addresses or other sensitive information that should not be sent to the client.

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

## ICE Lite

ICE Lite is a minimal implementation of the ICE protocol intended for servers running on a public IP address. It is not enabled by default to allow more complex ICE configurations out of the box.

```yaml title="config.yaml"
webrtc:
  icelite: false
```

:::info
When using ICE Servers, ICE Lite must be disabled.
:::

