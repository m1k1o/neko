# Troubleshooting

Neko UI loads, but you don't see the screen, and it gives you a `connection timeout` or `disconnected` error?

## Test your client {#client}

Some browsers may block WebRTC access by default. You can check if it is enabled by going to `about:webrtc` or `chrome://webrtc-internals` in your browser.

Check if your extensions are not blocking WebRTC access. The following extensions are known to block or not work properly with WebRTC:
- Privacy Badger
- Private Internet Access
- PIA VPN (even if disabled)

Test whether your client [supports](https://www.webrtc-experiment.com/DetectRTC/) and can [connect to WebRTC](https://www.webcasts.com/webrtc/).

## Networking {#networking}

If you are absolutely sure that your client is working correctly, then most likely your networking is not set up correctly.

### Check if your ports are correctly exposed in Docker {#exposed-ports}

Check that your ephemeral port range [`NEKO_WEBRTC_EPR`](/docs/v3/configuration/webrtc#epr) is correctly exposed as a `/udp` port range.

In the following example, the specified range `52000-52100` must also be exposed using Docker. You can't map it to a different range, e.g. `52000-52100:53000-53100/udp`. If you want to use a different range, you must change the range in [`NEKO_WEBRTC_EPR`](/docs/v3/configuration/webrtc#epr) too.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
    - "8080:8080"
    # highlight-start
    - "52000-52100:52000-52100/udp"
    # highlight-end
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      # highlight-start
      NEKO_WEBRTC_EPR: 52000-52100
      # highlight-end
      NEKO_WEBRTC_ICELITE: 1
```

### Validate UDP ports reachability {#reachable-ports}

Ensure that your ports are reachable through your external IP.

To validate the UDP connection the simplest way, run this on your server:

```shell
nc -ul 52101
```

And this on your local client:

```shell
nc -u [server ip] 52101
```
Then try to type on one end, you should see characters on the other side.

If it does not work for you, then most likely your port forwarding is not working correctly. Or your ISP is blocking traffic.

:::tip
If you get a [`command 'nc' not found`](https://command-not-found.com/nc) error, you can install the `netcat` package using:

```shell
sudo apt-get install netcat
```
:::

### Check if your external IP was determined correctly {#external-ip}

One of the first logs, when the server starts, writes down your external IP that will be sent to your clients to connect to.

```shell
docker compose logs neko | grep nat_ips
```

:::note
`docker-compose` was replaced with `docker compose` (no hyphen) recently. If you are using an older version of docker-compose, you should use `docker-compose` instead of `docker compose`.
:::

You should see this:

```
11:11AM INF webrtc starting ephemeral_port_range=52000-52100 ice_lite=true ice_servers="[{URLs:[stun:stun.l.google.com:19302] Username: Credential:<nil> CredentialType:password}]" module=webrtc nat_ips=<your-IP>
```

If your IP is not correct, you can specify your own IP resolver using [`NEKO_WEBRTC_IP_RETRIEVAL_URL`](/docs/v3/configuration/webrtc#ip_retrieval_url). It needs to return the IP address that will be used.

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
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
      # highlight-start
      NEKO_WEBRTC_IP_RETRIEVAL_URL: https://ifconfig.co/ip
      # highlight-end
```

Or you can specify your IP address manually using [`NEKO_WEBRTC_NAT1TO1`](/docs/v3/configuration/webrtc#nat1to1):

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
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
      # highlight-start
      NEKO_WEBRTC_NAT1TO1: <your-IP>
      # highlight-end
```

:::note
It's read as `NAT One to One`, so it's a capital letter `O`, not zero `0`, in `NAT1TO1`.
:::

If you want to use neko only locally, you must put your local IP address here, otherwise, the public address will be used.

### Neko works externally, but not locally {#works-externally-but-not-locally}

You are probably missing the NAT Loopback (NAT Hairpinning) setting on your router.

Example for pfsense with truecharts docker container:
- First, port forward the relevant ports `8080` and `52000-52100/udp` for the container.
- Then turn on `Pure NAT` in pfsense (under system > advanced > firewall and nat).
  - Make sure to check the two boxes so it works.
- Make sure [`NEKO_WEBRTC_NAT1TO1`](/docs/v3/configuration/webrtc#nat1to1) is blank and the [`NEKO_WEBRTC_IP_RETRIEVAL_URL`](/docs/v3/configuration/webrtc#ip_retrieval_url) address is working correctly (if unset, the default value is chosen).
- Test externally to confirm it works.
- Internally you have to access it using `<your-public-ip>:port`

If your router does not support NAT Loopback (NAT Hairpinning), you can use turn servers to overcome this issue. See [more details here](/docs/v3/configuration/webrtc#iceservers) on how to set up a local coturn instance.

### Neko works locally, but not externally {#works-locally-but-not-externally}

Make sure that you are exposing your ports correctly.

If you put a local IP as `NEKO_WEBRTC_NAT1TO1`, external clients try to connect to that IP. But it is unreachable for them because it is your local IP. You must use your public IP address with port forwarding.

## Frequently Encountered Errors {#frequently-encountered-errors}

### Getting a black screen with a cursor, but no browser for Chromium-based browsers {#black-screen-with-cursor}

Check if you did not forget to add `cap_add` to your `docker-compose.yaml` file. Make sure that the `shm_size` is set to `2gb` or higher.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/chromium:latest"
    # highlight-start
    cap_add:
    - SYS_ADMIN
    # highlight-end
    restart: "unless-stopped"
    # highlight-start
    shm_size: "2gb"
    # highlight-end
    ports:
    - "8080:8080"
    - "52000-52100:52000-52100/udp"
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
```

### No internet in the remote browser {#no-internet}

Try visiting `https://1.1.1.1` in the browser. If it works, then your internet is functioning, but DNS is not resolving.

You can specify a custom DNS server in the Docker file using the `--dns` flag or in your `docker-compose.yaml` file.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/chromium:latest"
    # highlight-start
    dns:
    - 1.1.1.1
    - 8.8.8.8
    # highlight-end
    cap_add:
    - SYS_ADMIN
    restart: "unless-stopped"
    shm_size: "2gb"
    # ...
```

If it still doesn't work, the issue is likely in the Docker/networking configuration. Check if your Docker network is not conflicting with your host network.

List all Docker networks:

```bash
$ for n in `docker network ls --format '{{ .ID }}'`; do docker network inspect --format '{{ .IPAM.Config }} {{ .Name }}' $n; done
[{172.16.0.0/24  172.16.0.1 map[]}] bridge
[] host
[{172.17.0.0/24  172.17.0.1 map[]}] neko1-net
[] none
[{172.18.0.0/24  172.18.0.1 map[]}] neko2-net
```

You can check your host network using:

```bash
$ ip route | grep default
default via 172.18.0.1 dev eth0 proto dhcp src 172.18.0.2 metric 100
```

In this case, the host subnet is the same as `neko2-net`, meaning that the internet stops working as soon as a second Docker network is created.

To fix this, you can either remove the conflicting Docker network or change the subnet of the Docker network by modifying the `daemon.json` file:

```json title="/etc/docker/daemon.json"
{
  "default-address-pools": [
    {
      "base" : "10.10.0.0/16",
      "size" : 24
    }
  ]
}
```

### Browser is not starting with persistent profile {#browser-profile-not-starting}

If you are using a persistent profile like `google-chrome` shown below, and the browser is not starting (you see a black screen), it may be because the profile is corrupted or not mounted correctly.

```yaml title="docker-compose.yaml"
volumes:
# For google-chrome
- /data:/home/neko/.config/google-chrome
```

Possible reasons are:
- The profile is corrupted, which can happen if the container is not stopped properly. Browsers should be able to recover from this, but it may not work in some cases.
- The profile is not mounted to the correct path. Make sure that you are mounting the profile to the correct path in your `docker-compose.yaml` file.
- The profile is not owned by the correct user. Make sure that the profile is owned by the `neko` user in the container. You can check this by running the following command:

```bash
# Check the owner of the profile
docker exec -it <container-id> ls -la /home/neko/.config/google-chrome
# To change the owner
docker exec -it <container-id> chown -R neko:neko /home/neko/.config/google-chrome
```

### Common server errors {#common-server-errors}

```
WRN session created with an error error="invalid 1:1 NAT IP mapping"
```

Check your [`NEKO_WEBRTC_NAT1TO1`](/docs/v3/configuration/webrtc#nat1to1) or ensure that [`NEKO_WEBRTC_IP_RETRIEVAL_URL`](/docs/v3/configuration/webrtc#ip_retrieval_url) returns the correct IP.

---

```
WRN could not get server reflexive address udp6 stun:stun.l.google.com:19302: write udp6 [::]:52042->[2607:f8b0:4001:c1a::7f]:19302: sendto: cannot assign requested address
```

Check if your DNS is set up correctly, and if your IPv6 connectivity is working properly, or is disabled.

---

```
WRN undeclaredMediaProcessor failed to open SrtcpSession: the DTLS transport has not started yet module=webrtc subsystem=
```

Check if your UDP ports are exposed correctly and reachable.

### Common client errors {#common-client-errors}

```
Firefox can’t establish a connection to the server at ws://<your-IP>/ws?password=neko.
```

Check if your TCP port is exposed correctly and your reverse proxy is correctly proxying websocket connections. And if your browser has not disabled websocket connections.

---

```
NotAllowedError: play() failed because the user didn't interact with the document first
```

This error occurs when the browser blocks the video from playing because the user has not interacted with the document. You just need to manually click on the play button to start the video.

### Unrelated server errors {#unrelated-server-errors}

```
[ERROR:bus.cc(393)] Failed to connect to the bus: Could not parse server address: Unknown address type (examples of valid types are "tcp" and on UNIX "unix")
```

This error originates from the browser, that it could not connect to dbus. This does not affect us and can be ignored.

---

```
I: [pulseaudio] client.c: Created 0 "Native client (UNIX socket client)"
I: [pulseaudio] protocol-native.c: Client authenticated anonymously.
I: [pulseaudio] source-output.c: Trying to change sample spec
I: [pulseaudio] sink.c: Reconfigured successfully
I: [pulseaudio] source.c: Reconfigured successfully
I: [pulseaudio] client.c: Freed 0 "neko"
I: [pulseaudio] protocol-native.c: Connection died.
```

These are just logs from pulseaudio. Unless you have audio issues, you can ignore them.

### Broadcast pipeline not working with some ingest servers {#broadcast-pipeline-not-working}

See [related issue](https://github.com/m1k1o/neko/issues/276).

```
Could not connect to RTMP stream "'rtmp://<ingest-url>/live/<stream-key-removed> live=1'" for writing
```

Some ingest servers require the `live=1` parameter in the URL (e.g. `nginx-rtmp-module`). Some do not and do not accept apostrophes (e.g. `owncast`). You can try to change the pipeline to:

```yaml
NEKO_CAPTURE_BROADCAST_PIPELINE: "flvmux name=mux ! rtmpsink location={url} pulsesrc device={device} ! audio/x-raw,channels=2 ! audioconvert ! voaacenc ! mux. ximagesrc display-name={display} show-pointer=false use-damage=false ! video/x-raw,framerate=28/1 ! videoconvert ! queue ! x264enc bframes=0 key-int-max=0 byte-stream=true tune=zerolatency speed-preset=veryfast ! mux."
```

See more details in broadcast pipeline [documentation](/docs/v3/configuration/capture#broadcast.pipeline).