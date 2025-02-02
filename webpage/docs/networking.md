---
sidebar_position: 3
---

# Networking

- If you want to use n.eko in **external** network, you can omit `NEKO_NAT1TO1`. It will automatically get your Public IP.
- If you want to use n.eko in **internal** network, set `NEKO_NAT1TO1` to your local IP address (e.g. `NEKO_NAT1TO1: 192.168.1.20`)-

Currently, it is not supported to supply multiple NAT addresses directly to neko (see [#47](https://github.com/m1k1o/neko/issues/47)).

But it can be acheived by deploying own turn server alongside neko that is accessible from your LAN, see [Using turn servers instead of port forwarding](#using-turn-servers-instead-of-port-forwarding).

## Why so many ports?

- WebRTC needs UDP ports in order to transfer Audio/Video towards user and Mouse/Keyboard events to the server in real time.
- If you don't set `NEKO_ICELITE=true`, every user will need 2 UDP ports.
- If you set `NEKO_ICELITE=true`, every user will need only 1 UDP port. It is **recommended** to use *ice-lite*.
- Do not forget, they are **UDP** ports, that configuration must be correct in your firewall/router/docker.
- You can freely limit number of UDP ports. But you can't map them to different ports.
  - This **WON'T** work: `32000-32100:52000-52100/udp`
- You can change API port (8080).
  - This **WILL** work: `3000:8080`

## Using mux instead of epr

When using a mux, not so many ports are needed.

```yaml title="docker-compose.yml"
services:
  neko:
    image: "m1k1o/neko:firefox"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
    - "8080:8080"
    # highlight-start
    - "8081:8081/tcp"
    - "8082:8082/udp"
    # highlight-end
    environment:
      NEKO_SCREEN: 1920x1080@30
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      # highlight-start
      NEKO_TCPMUX: 8081
      NEKO_UDPMUX: 8082
      # highlight-end
      NEKO_ICELITE: 1
```

- When using mux, `NEKO_EPR` is ignored.
- Mux accepts only one port, not a range.
- You only need to expose maximum two ports for WebRTC on your router/firewall and have many users connected.
- It can even be the same port number, so e.g. `NEKO_TCPMUX: 8081` and `NEKO_UDPMUX: 8081`.
- The same port must be exposed from docker container, you can't map them to different ports. So `8082:8082` is OK, but `"5454:8082` will not work.
- You can use them alone (either TCP or UDP) when needed.
  - UDP is generally better for latency. But some networks block UDP so it is good to have TCP available as fallback.
- Still, using `NEKO_ICELITE=true` is recommended.

## Using turn servers instead of port forwarding

- If you don't want to use port forwarding, you can use turn servers.
- But you need to have your own turn server (e.g. [cotrun](https://github.com/coturn/coturn)) or have access to one.
- They are generally not free, because they require a lot of bandwidth.
- Please make sure that you correctly escape your turn server credentials in the environment variable or use aphostrophes.

```json title="NEKO_ICESERVERS"
[
  {
    "urls": [
      "turn:<MY-COTURN-SERVER>:443?transport=udp",
      "turn:<MY-COTURN-SERVER>:443?transport=tcp",
      "turns:<MY-COTURN-SERVER>:443?transport=udp",
      "turns:<MY-COTURN-SERVER>:443?transport=tcp"
    ],
    "credential": "<MY-COTURN-CREDENTIAL>"
  },
  {
    "urls": [
      "stun:stun.nextcloud.com:443"
    ]
  }
]
```

### Example with coturn

This setup adds local turn server to neko. It won't be reachable by your remote clients and your own IP won't be reachable from your lan. So it effectively just adds local candidate and allows connections from LAN.

```yaml title="docker-compose.yml"
services:
  neko:
    image: "m1k1o/neko:firefox"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
    - "8080:8080"
    - "52000-52100:52000-52100/udp"
    environment:
      NEKO_SCREEN: 1920x1080@30
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      # highlight-start
      NEKO_ICESERVERS: |
        [{
          "urls": [ "turn:<MY-COTURN-SERVER>:3478" ],
          "username": "neko",
          "credential": "neko"
        },{
          "urls": [ "stun:stun.nextcloud.com:3478" ]
        }]
      # highlight-end
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

- Replace `<MY-COTURN-SERVER>` with your LAN IP address, and allow ports `49160-49200/udp` and `3478/tcp` in your LAN.
- Make sure you don't use `NEKO_ICELITE: true` because ICE LITE does not support TURN servers.
