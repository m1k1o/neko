# Getting started & FAQ

Use the following docker images:
- `m1k1o/neko:latest` or `m1k1o/neko:firefox` - for Firefox.
- `m1k1o/neko:chromium` - for Chromium (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:google-chrome` - for Google Chrome (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:ungoogled-chromium` - for [Ungoogled Chromium](https://github.com/Eloston/ungoogled-chromium) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)) (by @whalehub).
- `m1k1o/neko:microsoft-edge` - for Microsoft Edge (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:brave` - for [Brave Browser](https://brave.com) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:tor-browser` - for Tor Browser.
- `m1k1o/neko:vncviewer` - for simple VNC viewer (specify `NEKO_VNC_URL` to your VNC target).
- `m1k1o/neko:vlc` - for VLC Video player (needs volume mounted to `/media` with local video files, or setting `VLC_MEDIA=/media` path).
- `m1k1o/neko:xfce` - for a shared desktop / installing shared software.
- `m1k1o/neko:base` - for custom base.

For ARM-based devices (like Raspberry Pi, with GPU hardware acceleration):
- `m1k1o/neko:arm-firefox` - for Firefox.
- `m1k1o/neko:arm-chromium` - for Chromium.
- `m1k1o/neko:arm-base` - for custom arm based.

Images (except `arm-`) are built using GitHub actions on every push and on weekly basis to keep all browsers up-to-date,

### Networking:
- If you want to use n.eko in **external** network, you can omit `NEKO_NAT1TO1`. It will automatically get your Public IP.
- If you want to use n.eko in **internal** network, set `NEKO_NAT1TO1` to your local IP address (e.g. `NEKO_NAT1TO1: 192.168.1.20`)-
- Currently, it is not supported to supply multiple NAT addresses (see https://github.com/m1k1o/neko/issues/47).

### Why so many ports?
- WebRTC needs UDP ports in order to transfer Audio/Video towards user and Mouse/Keyboard events to the server in real time.
- If you don't set `NEKO_ICELITE=true`, every user will need 2 UDP ports.
- If you set `NEKO_ICELITE=true`, every user will need only 1 UDP port. It is **recommended** to use *ice-lite*.
- Do not forget, they are **UDP** ports, that configuration must be correct in your firewall/router/docker.
- You can freely limit number of UDP ports. But you can't map them to different ports.
  - This **WON'T** work: `32000-32100:52000-52100/udp`
- You can change API port (8080).
  - This **WILL** work: `3000:8080`

#### But there is a hope!
There has been an attempt to implement [single port ice using tcp and udp mux](https://github.com/m1k1o/neko/commit/c97b1fc4541caabf6b00331d081b02d2f9c58751) ([#106](https://github.com/m1k1o/neko/pull/106)), that allows using one port instead (each for TCP and/or UDP). This feature is not properly tested yet and only experimental.

We can use TCP mux and/or UDP mux, example:

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:firefox"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "8081:8081/tcp"
      - "8082:8082/udp"
    environment:
      NEKO_SCREEN: 1920x1080@30
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_TCPMUX: 8081
      NEKO_UDPMUX: 8082
      NEKO_ICELITE: 1
```

### Want to customize and install own add-ons, set custom bookmarks?
- You would need to modify the existing policy file and mount it to your container.
- For Firefox, copy [this](https://github.com/m1k1o/neko/blob/master/.docker/firefox/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/usr/lib/firefox/distribution/policies.json'`
- For Chromium, copy [this](https://github.com/m1k1o/neko/blob/master/.docker/chromium/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/etc/chromium/policies/managed/policies.json'`
- For others, see where existing `policies.json` is placed in their `Dockerfile`.

### Want to use VPN for your n.eko browsing?
- Check this out: https://github.com/m1k1o/neko-vpn

### Want to have multiple rooms on demand?
- Check this out: https://github.com/m1k1o/neko-rooms

### Want to use different Apps than Browser?
- Check this out: https://github.com/m1k1o/neko-apps

### Accounts:
- There are no accounts, display name (a.k.a. username) can be freely chosen. Only password needs to match. Depending on which password matches, the visitor gets its privilege:
  - Anyone, who enters with `NEKO_PASSWORD` will be **user**.
  - Anyone, who enters with `NEKO_PASSWORD_ADMIN` will be **admin**.

### Screen size
- Only admins can change screen size.
- You can set a default screen size, but this size **MUST** be one from the list, that your server supports.
- You will get this list in frontend, where you can choose from.

### Clipboard sharing
- Browsers have certain requirements to allow clipboard sharing.
  - Your instance must be HTTPS.
  - Firefox does not support clipboard sharing.
  - Use Chrome for the best experience.
- If your browser does not support clipboard sharing:
  - Clipboard icon in the bottom right corner will be displayed for host.
  - It opens text area that can share clipboard content bi-directionally.
  - Only plain-text is supported.
