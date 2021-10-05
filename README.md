<div align="center">
  <a href="https://github.com/m1k1o/neko" title="Neko's Github repository.">
    <img src="https://raw.githubusercontent.com/m1k1o/neko/master/docs/_media/logo.png" width="450" height="auto"/>
  </a>
  <p align="center">
    <img src="https://img.shields.io/github/v/release/m1k1o/neko" alt="release">
    <img src="https://img.shields.io/github/license/m1k1o/neko" alt="license">
    <img src="https://img.shields.io/docker/pulls/m1k1o/neko" alt="pulls">
    <img src="https://img.shields.io/github/issues/m1k1o/neko" alt="issues">
    <a href="https://discord.gg/3U6hWpC">
      <img src="https://discordapp.com/api/guilds/665851821906067466/widget.png" alt="Chat on discord">
    </a>
    <a href="https://github.com/m1k1o/neko/actions">
      <img src="https://github.com/m1k1o/neko/actions/workflows/build.yml/badge.svg" alt="build">
    </a>
  </p>
  <br/>
  <br/>
  <img src="https://i.imgur.com/ZSzbQr7.gif" width="650" height="auto"/>
  <br/>
  <br/>
</div>

# n.eko (m1k1o fork)
This app uses Web-RTC to stream a desktop inside a docker container. This is a fork of https://github.com/nurdism/neko.

For n.eko room management software, visit https://github.com/m1k1o/neko-rooms.

## Differences to original repository.

### New Features
- Clipboard button with text area - for browsers, that don't support clipboard syncing (Firefox, what a shame...) or for HTTP.
- Keyboard modifier state synchronization (Num-Lock, Caps-Lock, Scroll-Lock) for each hosting.
- Added chromium ungoogled (with h265 support) an kept up to date (by @whalehub).
- Added Picture in Picture button (only for watching screen, controlling not possible).
- Added RTMP broadcast. Enables broadcasting n.eko screen to local RTMP server, YouTube or Twitch. Example: `rtmp://a.rtmp.youtube.com/live2/<your-streaming-key>`.
- Stereo sound (works properly only in Firefox host).
- Added limited support for some mobile browsers with `playsinline` attribute.
- Added `VIDEO_BITRATE` and `AUDIO_BITRATE` in kbit/s to control stream quality (in collaboration with @mbattista).
- Added `MAX_FPS`, where you can specify max WebRTC frame rate. When set to `0`, frame rate won't be capped, and you can enjoy your real `60fps` experience. Originally, it was constant at `25fps`.
- Invite links. You can invite people, and they don't need to enter passwords by themselves (and get confused about user accounts that do not exist). You can put your password in URL using `?pwd=<your-password>` and it will be automatically used when logging in.
- Added `/stats?pwd=<admin>` endpoint to get total active connections, host and members.
- Added `m1k1o/neko:vlc` tag, use VLC to watch local files together (by @mbattista).
- Added `m1k1o/neko:xfce` tag, as a non-video related showcase (by @mbattista).
- Added ARM-based images, for Raspberry Pi support (by @mbattista).
- Added simple language picker.
- Added `?usr=<display-name>` that will prefill username. This allows creating auto-join links.
- Added `?cast=1` that will hide all control and show only video.
- Shake keyboard icon if someone attempted to control when is nobody hosting.
- Support for password protected `NEKO_ICESERVERS` (by @mbattista).
- Added bunch of translations (🇸🇰, 🇪🇸, 🇸🇪, 🇳🇴, 🇫🇷) by various people.
- Added `m1k1o/neko:google-chrome` tag.
- Show red dot badge on sidebar toggle if there are new messages, and user can't see them.
- Added `m1k1o/neko:brave` tag.

### Bugs
- Fixed minor gst pipeline bug.
- Locked screen only for users, admins can still join.
- Fixed h264 pipelines bugs (by @mbattista).
- Fixed sessions manager thread safety by adding mutexes (caused panic in rare edge cases).
- Now when user gets kicked, he won't join as a ghost user again but will be logged out.
- **iOS compatibility!** Fixed a really strange CSS bug, which prevented iOS from loading the video.
- Proper disconnect only once with unsubscribing events. When Web-RTC fails, user won't be logged in without username again.
- Upgraded and fixed emojis to a new major version.
- Fixed bad `keymap -> keysym` translation to respect active modifiers (#45, with @mbattista).
- Respecting `NEKO_DEBUG` env variable.
- Full-screen support for iOS devices.
- Added `chrome-sandbox` to fix weird bug when chromium didn't start.
- Fixed keyboard mapping on macOS, when CMD could not be used for copy & paste.
- Fixed stop signal sent by supervisor to gracefully shut down neko server.

### Misc
- Custom docker workflow.
- Based on Debian buster instead of stretch.
- Versions bumped: Go 16, Node.js 14 (by @mbattista).
- Custom avatars without any 3rd party dependency.
- Ignore duplicate notify bars.
- No pointer events for notify bars.
- Disable debug mode by default.
- Remove HTML tags from username.
- Upgraded `pion/webrtc` to v3 (by @mbattista).
- Added `requestFullscreen` compatibility for older browsers and iOS devices.
- Fixed small lags in video and improved video UX (by @mbattista).
- Added `m1k1o/neko:vncviewer` tag, use `NEKO_VNC_URL` to specify VNC target and use n.eko as a bridge.
- Ability to include n.eko as a component in another Vue.Js project (by @gbrian).
- Added HEALTHCHECK to Dockerfile.
- Arguments in broadcast pipeline are optional, not positional and can be repeated `{url} {device} {display}`.
- Chat messages are dense, when repeated, they are joined together.
- While IP address fetching is now proxy ignored.
- Start unmuted on reconnects.
- Switched to the latest Firefox version instead of esr.
- Fixed very fast scroll speed on macOS.
- Broadcast pipeline errors are reported to the user.
- On stopping server all websocket connections are going to be gracefully disconnected.
- ARM-based images not bound to Raspberry Pi only.

# Getting started & FAQ

Use the following docker images:
- `m1k1o/neko:latest` - for Firefox.
- `m1k1o/neko:chromium` - for Chromium (needs `--cap-add=SYS_ADMIN`).
- `m1k1o/neko:google-chrome` - for Google Chrome (needs `--cap-add=SYS_ADMIN`).
- `m1k1o/neko:ungoogled-chromium` - for [Ungoogled Chromium](https://github.com/Eloston/ungoogled-chromium) (needs `--cap-add=SYS_ADMIN`) (by @whalehub).
- `m1k1o/neko:brave` - for [Brave Browser](https://brave.com) (needs `--cap-add=SYS_ADMIN`).
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

### Behind reverse proxy?

<details>
  <summary>Traefik2 configuration</summary>

  ```yaml
  labels:
    - "traefik.enable=true"
    - "traefik.http.services.neko-frontend.loadbalancer.server.port=8080"
    - "traefik.http.routers.neko.rule=${TRAEFIK_RULE}"
    - "traefik.http.routers.neko.entrypoints=${TRAEFIK_ENTRYPOINTS}"
    - "traefik.http.routers.neko.tls.certresolver=${TRAEFIK_CERTRESOLVER}"
  ```

  (by @m1k1o, [example](https://github.com/m1k1o/neko-vpn/blob/a1b934515dcf597992a515d61d307c2450a11002/docker-compose.yml#L38-L43))
</details>

<details>
  <summary>Nginx configuration</summary>

  ```conf
  server {
    listen 443 ssl http2;
    server_name example.com;

    location / {
      proxy_pass http://127.0.0.1:8080;
      proxy_http_version 1.1;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";
      proxy_read_timeout 86400;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $remote_addr;
      proxy_set_header X-Forwarded-Host $host;
      proxy_set_header X-Forwarded-Port $server_port;
      proxy_set_header X-Forwarded-Protocol $scheme;
    }
  }
  ```

  (by @GigaFyde, [source](https://github.com/nurdism/neko/issues/111#issuecomment-742656957))
</details>

<details>
  <summary>Apache configuration</summary>

  ```xml
  <VirtualHost *:80>
    # The ServerName directive sets the request scheme, hostname and port that
    # the server uses to identify itself. This is used when creating
    # redirection URLs. In the context of virtual hosts, the ServerName
    # specifies what hostname must appear in the request's Host: header to
    # match this virtual host. For the default virtual host (this file) this
    # value is not decisive as it is used as a last resort host regardless.
    # However, you must set it for any further virtual host explicitly.

    # Paths of those modules might vary across different distros.
    LoadModule proxy_module /usr/lib/apache2/modules/mod_proxy.so
    LoadModule proxy_http_module /usr/lib/apache2/modules/mod_proxy_http.so
    LoadModule proxy_wstunnel_module /usr/lib/apache2/modules/mod_proxy_wstunnel.so

    ServerName example.com
    ServerAlias www.example.com

    ProxyRequests Off
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/

    RewriteEngine on
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule /ws(.*) "ws://localhost:8080/ws$1" [P,L]

    # Available loglevels: trace8, ..., trace1, debug, info, notice, warn,
    # error, crit, alert, emerg.
    # It is also possible to configure the loglevel for particular
    # modules, e.g.
    #LogLevel info ssl:warn

    ErrorLog ${APACHE_LOG_DIR}/error.log
    CustomLog ${APACHE_LOG_DIR}/access.log combined

    # For most configuration files from conf-available/, which are
    # enabled or disabled at a global level, it is possible to
    # include a line for only one particular virtual host. For example the
    # following line enables the CGI configuration for this host only
    # after it has been globally disabled with "a2disconf".
    #Include conf-available/serve-cgi-bin.conf
  </VirtualHost>
  ```

  (by @DarkReaper231, [source](https://github.com/nurdism/neko/blob/cad98a62a5bd7f1daf2c11980631bb14ba81a1f6/docs/apache-proxypass-config.md#example-apache-config))
</details>

<details>
  <summary>Caddy configuration</summary>

  ```conf
  https://example.com {
    reverse_proxy localhost:8080 {
      header_up Host {host}
      header_up X-Real-IP {remote_host}
      header_up X-Forwarded-For {remote_host}
      header_up X-Forwarded-Proto {scheme}
    }
  }
  ```

  (by @ccallahan, [source](https://github.com/nurdism/neko/pull/125/commits/eb4ceda75423b0d960c8aea0240acf6d7a10fef4))
</details>

### Want to customize and install own add-ons, set custom bookmarks?
- You would need to modify the existing policy file and mount it to your container.
- For Firefox, copy [this](https://github.com/m1k1o/neko/blob/dev/.m1k1o/firefox/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/usr/share/firefox-esr/distribution/policies.json'`
- For Chromium, copy [this](https://github.com/m1k1o/neko/blob/dev/.m1k1o/chromium/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/etc/chromium/policies/managed/policies.json'`

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

## Firefox

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## Chromium

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:chromium"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## VLC

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:vlc"
    restart: "unless-stopped"
    shm_size: "2gb"
    volumes:
      - "<your-video-folder>:/video" 
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## Raspberry Pi

Note! Since HW accelerated pipeline is using H264, you are only able to connect from browsers supporting H264 for WebRTC. At the time of implementing, [Firefox does not support this](https://developer.mozilla.org/en-US/docs/Web/Media/Formats/WebRTC_codecs#supported-foot-1). When omitting `NEKO_VIDEO` and `NEKO_H264` parameters, you get default CPU encoding with VP8.

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:arm-chromium"
    restart: "unless-stopped"
    # increase on rpi's with more then 1gb ram.
    shm_size: "520mb"
    ports:
      - "8088:8080"
      - "52000-52100:52000-52100/udp"
    # note: this is important since we need a GPU for hardware acceleration alternatively
    #       mount the devices into the docker.
    privileged: true
    environment:
      NEKO_SCREEN: '1280x720@30'
      NEKO_PASSWORD: 'neko'
      NEKO_PASSWORD_ADMIN: 'admin'
      NEKO_EPR: 52000-52100
      # note: when setting NEKO_VIDEO, then variables NEKO_MAX_FPS and NEKO_VIDEO_BITRATE
      #       are not being used, you can adjust them in this variable.
      NEKO_VIDEO: |
        ximagesrc display-name=%s use-damage=0 show-pointer=true use-damage=false
          ! video/x-raw,framerate=30/1
          ! videoconvert
          ! queue
          ! video/x-raw,framerate=30/1,format=NV12
          ! v4l2h264enc extra-controls="controls,h264_profile=0,video_bitrate=1250000;"
          ! h264parse config-interval=3
          ! video/x-h264,profile=baseline,stream-format=byte-stream
      NEKO_H264: 1
```

## Not using docker?

You can execute `neko --help` to see available arguments. In [Dockerfile](https://github.com/m1k1o/neko/blob/dev/.m1k1o/base/Dockerfile) you can find required dependencies and install them manually.

## Mobile support

N.eko is now working on iOS and Android! Also, the UI screens have been fixed for small screens.

![mobile-screens](https://i.imgur.com/K9gfscU.png)

## Docker-Compose Environment Options

```code
NEKO_SCREEN:
  - Resolution after startup. Only Admins can change this later.
  - e.g. '1920x1080@30'
NEKO_PASSWORD:
  - Password for the user login
  - e.g. 'user_password'
NEKO_PASSWORD_ADMIN
  - Password for the admin login
  - e.g. 'admin_password'
NEKO_EPR:
  - For WebRTC needed range of ports
  - e.g. 52000-52100
NEKO_VP8:
  - If vp8 should be used as video encoder for the stream (default encoder)
  - e.g. 'true'
NEKO_VP9:
  - If vp9 should be used as video encoder for the stream (Parameter not optimized yet)
  - e.g. 'false'
NEKO_H264:
  - If h264 should be used as video encoder for the stream (second best option)
  - e.g. 'false'
NEKO_VIDEO_BITRATE:
  - Bitrate of the video stream in kb/s
  - e.g. 3500
NEKO_VIDEO:
  - Makes it possible to create custom gstreamer pipelines. With this you could find the best quality for your CPU
  - Installed are gstreamer1.0-plugins-base /  gstreamer1.0-plugins-good /  gstreamer1.0-plugins-bad /  gstreamer1.0-plugins-ugly
  - e.g. ' ximagesrc display-name=%s show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue ! video/x-raw,format=NV12 ! x264enc threads=4 bitrate=3500 key-int-max=60 vbv-buf-capacity=4000 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream '
NEKO_MAX_FPS:
  - The resulting stream frames per seconds should be capped (0 for uncapped)
  - e.g. 0
NEKO_OPUS:
  - If opus should be used as audio encoder for the stream (default encoder)
  - e.g. 'true'
NEKO_G722:
  - If g722 should be used as audio encoder for the stream
  - e.g. 'false'
NEKO_PCMU:
  - If pcmu should be used as audio encoder for the stream
  - e.g. 'false'
NEKO_PCMA:
  - If pcma should be used as audio encoder for the stream
  - e.g. 'false'
NEKO_AUDIO_BITRATE:
  - Bitrate of the audio stream in kb/s
  - e.g. 196
NEKO_CERT:
  - Path to the SSL-Certificate
  - e.g. '/certs/cert.pem'
NEKO_KEY:
  - Path to the SSL-Certificate private key
  - e.g. '/certs/key.pem'
NEKO_ICELITE:
  - Use the ice lite protocol
  - e.g. false
NEKO_ICESERVER:
  - Describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer (simple usage for server without authentication)
  - e.g. 'stun:stun.l.google.com:19302'
NEKO_ICESERVERS:
  - Describes multiple STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer
  - e.g. '[{"urls": ["turn:turn.example.com:19302", "stun:stun.example.com:19302"], "username": "name", "credential": "password"}, {"urls": ["stun:stun.example2.com:19302"]}]'
  - [More information](https://developer.mozilla.org/en-US/docs/Web/API/RTCIceServer)
```

# How to contribute? How to build?

Navigate to [.m1k1o/README.md](.m1k1o/README.md) for further information.
