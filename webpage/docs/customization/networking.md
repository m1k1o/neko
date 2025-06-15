---
sidebar_label: "Networking"
description: "Example networking configurations for Neko."
---

# Networking Customization

## Accessing Neko over the Internet {#internet}

If you want to access Neko over the internet, you need to expose the necessary ports on your router or firewall.

This is the default configuration for Neko so no additional configuration is needed.

## Accessing Neko over a VPN {#vpn}

If you want to access Neko over a VPN, you need to set NAT1TO1 to your server's IP address in the VPN network. This allows Neko to communicate with the server over the VPN.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
      NEKO_WEBRTC_NAT1TO1: <your-VPN-IP>
```

## Accessing Neko over SSH {#ssh}

If you do not want to expose Neko to the internet and want to access it securely over SSH, you can set up port forwarding using SSH. This allows you to access Neko from your local machine without exposing it to the internet.

Start neko with TCP multiplexing enabled and NAT1to1 set to loopback address. That way everytime you access Neko, it will use the loopback address to connect to the server.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/nvidia-firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000:52000"
    environment:
      NEKO_WEBRTC_TCPMUX: 52000
      NEKO_WEBRTC_ICELITE: 1
      NEKO_WEBRTC_NAT1TO1: 127.0.0.1
```

Set up your SSH configuration file (`~/.ssh/config`) to include the following port forwarding settings. This will forward the ports from the remote server to your local machine.

```shell title="~/.ssh/config"
Host PC-Work
    HostName work.example.com
    User xxx
    Port xyy
    RemoteForward 8080 localhost:8080
    RemoteForward 52000 localhost:52000
```
