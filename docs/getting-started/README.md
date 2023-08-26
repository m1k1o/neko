# Getting started & FAQ

<div align="center">
  <img src="../_media/icons/firefox.svg" title="m1k1o/neko:firefox" width="60" height="auto"/>
  <img src="../_media/icons/google-chrome.svg" title="m1k1o/neko:google-chrome" width="60" height="auto"/>
  <img src="../_media/icons/chromium.svg" title="m1k1o/neko:chromium" width="60" height="auto"/>
  <img src="../_media/icons/microsoft-edge.svg" title="m1k1o/neko:microsoft-edge" width="60" height="auto"/>
  <img src="../_media/icons/brave.svg" title="m1k1o/neko:brave" width="60" height="auto"/>
  <img src="../_media/icons/vivaldi.svg" title="m1k1o/neko:vivaldi" width="60" height="auto"/>
  <img src="../_media/icons/opera.svg" title="m1k1o/neko:opera" width="60" height="auto"/>
  <img src="../_media/icons/tor-browser.svg" title="m1k1o/neko:tor-browser" width="60" height="auto"/>
  <img src="../_media/icons/remmina.png" title="m1k1o/neko:remmina" width="60" height="auto"/>
  <img src="../_media/icons/vlc.svg" title="m1k1o/neko:vlc" width="60" height="auto"/>
  <img src="../_media/icons/xfce.svg" title="m1k1o/neko:xfce" width="60" height="auto"/>
  <img src="../_media/icons/kde.svg" title="m1k1o/neko:kde" width="60" height="auto"/>
</div>

Use the following docker images from [Docker Hub](https://hub.docker.com/r/m1k1o/neko) for x86_64:
- `m1k1o/neko:latest` or `m1k1o/neko:firefox` - for Firefox.
- `m1k1o/neko:chromium` - for Chromium (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:google-chrome` - for Google Chrome (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:ungoogled-chromium` - for [Ungoogled Chromium](https://github.com/Eloston/ungoogled-chromium) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)) (by @whalehub).
- `m1k1o/neko:microsoft-edge` - for Microsoft Edge (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:brave` - for [Brave Browser](https://brave.com) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:vivaldi` - for [Vivaldi Browser](https://vivaldi.com) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)) (by @Xeddius).
- `m1k1o/neko:opera` for [Opera Browser](https://opera.com) (requires extra steps to enable DRM, see instructions [here](https://www.reddit.com/r/operabrowser/wiki/opera/linux_widevine_config/). libffmpeg is already configured.) (by @prophetofxenu)
- `m1k1o/neko:tor-browser` - for Tor Browser.
- `m1k1o/neko:remmina` - for remote desktop connection (by @lowne).
  - Pass env var `REMMINA_URL=<proto>://[<username>[:<password>]@]server[:port]` (proto being `vnc`, `rdp` or `spice`).
  - Or create your custom configuration with remmina locally (it's saved in `~/.local/share/remmina/path_to_profile.remmina`) and bind-mount it, then pass env var `REMMINA_PROFILE=<path_to_profile.remmina>`.
- `m1k1o/neko:vlc` - for VLC Video player (needs volume mounted to `/media` with local video files, or setting `VLC_MEDIA=/media` path).
- `m1k1o/neko:xfce` or `m1k1o/neko:kde` - for a shared desktop / installing shared software.
- `m1k1o/neko:base` - for custom base.

Dockerhub images are built using GitHub actions on every push and on weekly basis to keep all browsers up-to-date.

All images are also available on [GitHub Container Registry](https://github.com/m1k1o?tab=packages&repo_name=neko) for faster pulls:

- `ghcr.io/m1k1o/neko/firefox:latest`
- `ghcr.io/m1k1o/neko/chromium:latest`
- `ghcr.io/m1k1o/neko/google-chrome:latest`
- `ghcr.io/m1k1o/neko/ungoogled-chromium:latest`
- `ghcr.io/m1k1o/neko/microsoft-edge:latest`
- `ghcr.io/m1k1o/neko/brave:latest`
- `ghcr.io/m1k1o/neko/vivaldi:latest`
- `ghcr.io/m1k1o/neko/opera:latest`
- `ghcr.io/m1k1o/neko/tor-browser:latest`
- `ghcr.io/m1k1o/neko/remmina:latest`
- `ghcr.io/m1k1o/neko/vlc:latest`
- `ghcr.io/m1k1o/neko/xfce:latest`
- `ghcr.io/m1k1o/neko/kde:latest`

For ARM-based images (like Raspberry Pi - with GPU hardware acceleration, Oracle Cloud ARM tier). Currently, not all images are available for ARM, because not all applications are available for ARM. Please note, that `m1k1o/neko:arm-*` images from dockerhub are currently not maintained and they can contain outdated software. Please use images below:

- `ghcr.io/m1k1o/neko/arm-firefox:latest`
- `ghcr.io/m1k1o/neko/arm-chromium:latest`
- `ghcr.io/m1k1o/neko/arm-ungoogled-chromium:latest`
- `ghcr.io/m1k1o/neko/arm-vlc:latest`
- `ghcr.io/m1k1o/neko/arm-xfce:latest`

For images with VAAPI GPU hardware acceleration using intel drivers use:

- `ghcr.io/m1k1o/neko/intel-firefox:latest`
- `ghcr.io/m1k1o/neko/intel-chromium:latest`
- `ghcr.io/m1k1o/neko/intel-google-chrome:latest`
- `ghcr.io/m1k1o/neko/intel-ungoogled-chromium:latest`
- `ghcr.io/m1k1o/neko/intel-microsoft-edge:latest`
- `ghcr.io/m1k1o/neko/intel-brave:latest`
- `ghcr.io/m1k1o/neko/intel-vivaldi:latest`
- `ghcr.io/m1k1o/neko/intel-opera:latest`
- `ghcr.io/m1k1o/neko/intel-tor-browser:latest`
- `ghcr.io/m1k1o/neko/intel-remmina:latest`
- `ghcr.io/m1k1o/neko/intel-vlc:latest`
- `ghcr.io/m1k1o/neko/intel-xfce:latest`
- `ghcr.io/m1k1o/neko/intel-kde:latest`

For images with Nvidia GPU hardware acceleration using EGL (see example below) use (please note, there is a known issue with EGL and Chromium-based browsers, see [here](https://github.com/m1k1o/neko/issues/279)):

- `ghcr.io/m1k1o/neko/nvidia-firefox:latest`
- `ghcr.io/m1k1o/neko/nvidia-chromium:latest`
- `ghcr.io/m1k1o/neko/nvidia-google-chrome:latest`
- `ghcr.io/m1k1o/neko/nvidia-microsoft-edge:latest`
- `ghcr.io/m1k1o/neko/nvidia-brave:latest`

GHCR images are built using GitHub actions for every tag.

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

### Using mux instead of epr

When using a mux, not so many ports are needed.

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

- When using mux, `NEKO_EPR` is ignored.
- Mux accepts only one port, not a range.
- You only need to expose maximum two ports for WebRTC on your router/firewall and have many users connected.
- It can even be the same port number, so e.g. `NEKO_TCPMUX: 8081` and `NEKO_UDPMUX: 8081`.
- The same port must be exposed from docker container, you can't map them to different ports. So `8082:8082` is OK, but `"5454:8082` will not work.
- You can use them alone (either TCP or UDP) when needed.
  - UDP is generally better for latency. But some networks block UDP so it is good to have TCP available as fallback.
- Still, using `NEKO_ICELITE=true` is recommended.

### Using turn servers instead of port forwarding

- If you don't want to use port forwarding, you can use turn servers.
- But you need to have your own turn server (e.g. [cotrun](https://github.com/coturn/coturn)) or have access to one.
- They are generally not free, because they require a lot of bandwidth.
- Please make sure that you correctly escape your turn server credentials in the environment variable or use aphostrophes.

```yaml
NEKO_ICESERVERS: '[{"urls": ["turn:<MY-COTURN-SERVER>:443?transport=udp", "turn:<MY-COTURN-SERVER>:443?transport=tcp", "turns:<MY-COTURN-SERVER>:443?transport=udp", "turns:<MY-COTURN-SERVER>:443?transport=tcp"], "credential": "<MY-COTURN-CREDENTIAL"}, {"urls": ["stun:stun.nextcloud.com:443"]}]'
```

### Want to customize and install own add-ons, set custom bookmarks?
- You would need to modify the existing policy file and mount it to your container.
- For Firefox, copy [this](https://github.com/m1k1o/neko/blob/master/.docker/firefox/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/usr/lib/firefox/distribution/policies.json'`
- For Chromium, copy [this](https://github.com/m1k1o/neko/blob/master/.docker/chromium/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/etc/chromium/policies/managed/policies.json'`
- For others, see where existing `policies.json` is placed in their `Dockerfile`.

#### Allow file uploading & downloading
- From security perspective, browser is not enabled to access local file data.
- If you want to enable this, you need to modify following policies:
```json
  "DownloadRestrictions": 0,
  "AllowFileSelectionDialogs": true,
  "URLAllowlist": [
      "file:///home/neko/Downloads"
  ],
```

### Want to preserve browser data between restarts?
- You need to mount browser profile as volume.
- For Firefox, that is this `/home/neko/.mozilla/firefox/profile.default` folder, mount it as: ` -v '${PWD}/data:/home/neko/.mozilla/firefox/profile.default'`
- For Chromium, that is this `/home/neko/.config/chromium` folder, mount it as: ` -v '${PWD}/data:/home/neko/.config/chromium'`
- For other chromium based browsers, see in `supervisord.conf` folder that is specified in `--user-data-dir`.

#### Allow persistent data in policies
- From security perspective, browser is set up to forget all cookies and browsing history when its closed.
- If you want to enable this, you need to modify following policies:
```json
  "DefaultCookiesSetting": 1,
  "RestoreOnStartup": 1,
```

### Nvidia GPU acceleration

You need to have [nvidia-docker](https://github.com/NVIDIA/nvidia-docker) installed, start the container with `--gpus all` flag and use images built for nvidia (see above).

```bash
docker run -d --gpus all \
  -p 8080:8080 \
  -p 56000-56100:56000-56100/udp \
  -e NEKO_SCREEN=1920x1080@30 \
  -e NEKO_PASSWORD=neko \
  -e NEKO_PASSWORD_ADMIN=admin \
  -e NEKO_EPR=56000-56100 \
  -e NEKO_NAT1TO1=192.168.1.10 \
  -e NEKO_ICELITE=1 \
  -e NEKO_VIDEO_CODEC=h264 \
  -e NEKO_HWENC=nvenc \
  --shm-size=2gb \
  --cap-add=SYS_ADMIN \
  --name neko \
  ghcr.io/m1k1o/neko/nvidia-google-chrome:latest
```

If you want to use docker-compose, you can use this example:

```yaml
version: "3.4"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/nvidia-google-chrome:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "56000-56100:56000-56100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 56000-56100
      NEKO_NAT1TO1: 192.168.1.10
      NEKO_VIDEO_CODEC: h264
      NEKO_HWENC: nvenc
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

- You can verify that GPU is available inside the container by running `docker exec -it neko nvidia-smi` command.
- You can verify that GPU is used for encoding by searching for `nvh264enc` in `docker logs neko` output.
- If you don'ŧ specify `NEKO_HWENC: nvenc` environment variable, CPU encoding will be used but GPU will still be available for browser rendering.

Broadcast pipeline is not hardware accelerated by default. You can use this pipeline created by [@evilalmus](https://github.com/m1k1o/neko/issues/276#issuecomment-1498362533).

```yaml
NEKO_BROADCAST_PIPELINE: "flvmux name=mux ! rtmpsink location={url} pulsesrc device={device} ! audio/x-raw,channels=2 ! audioconvert ! voaacenc ! mux. ximagesrc display-name={display} show-pointer=false use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue ! video/x-raw,format=NV12 ! nvh264enc name=encoder preset=low-latency-hq gop-size=25 spatial-aq=true temporal-aq=true bitrate=2800 vbv-buffer-size=2800 rc-mode=6 ! h264parse config-interval=-1 ! video/x-h264,stream-format=byte-stream,profile=high ! h264parse ! mux."
```

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
- Disabling passwords is not possible. However, you can use following query parameters to create auto-join links:
  - Adding `?pwd=<password>` will prefill password.
  - Adding `?usr=<display-name>` will prefill username.
  - Adding `?cast=1` will hide all control and show only video.
  - Adding `?embed=1` will hide most additional components and show only video.
  - Adding `?volume=<0-1>` will set volume to given value.
  - Adding `?lang=<language>` will set language to given value.
  - e.g. `http(s)://<URL:Port>/?pwd=neko&usr=guest&cast=1`

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
