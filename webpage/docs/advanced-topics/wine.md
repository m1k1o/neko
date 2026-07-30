---
sidebar_label: "Running Wine"
description: "Run Windows applications inside a Neko container using Wine."
---

# Running Wine inside Neko

[Wine](https://www.winehq.org/) is a compatibility layer that allows running Windows applications on Linux. Because Neko runs a full Xorg display server, Wine applications can render their windows inside the Neko virtual desktop — giving you browser-accessible remote Windows app sessions.

:::info
This guide covers building a custom Docker image that extends a Neko base image with Wine. The result is a containerized Windows application accessible from any browser, just like any other Neko app.
:::

## How it works

Neko provides a real Xorg display (not a virtual framebuffer), which is exactly what Wine needs to create windows. The `DISPLAY` and `PULSE_SERVER` environment variables are set automatically by the neko runtime, so Wine applications will use the correct display and audio output out of the box.

The most common error when Wine fails to start is:

```
err:winediag:nodrv_CreateWindow Application tried to create a window, but no driver could be loaded.
err:winediag:nodrv_CreateWindow The explorer process failed to start.
```

This happens when Wine cannot connect to an X11 display. Inside a Neko container this is handled automatically — ensure you are launching Wine **after** Xorg has started (i.e. from a supervisord program, not from a Docker `CMD`).

## Building a custom image

Extend any Neko base image (e.g. `xfce` for a full desktop or a browser image) with Wine:

```dockerfile title="Dockerfile"
FROM ghcr.io/m1k1o/neko/xfce:latest

# Enable 32-bit architecture required by many Windows apps
RUN dpkg --add-architecture i386

# Add WineHQ repository for the latest stable release
RUN apt-get update && apt-get install -y --no-install-recommends \
        software-properties-common \
        gnupg2 \
        wget \
    && wget -qO /etc/apt/keyrings/winehq-archive.key \
        https://dl.winehq.org/wine-builds/winehq.key \
    && wget -NP /etc/apt/sources.list.d/ \
        https://dl.winehq.org/wine-builds/debian/dists/bookworm/winehq-bookworm.sources \
    && apt-get update \
    && apt-get install -y --install-recommends winehq-stable \
    && rm -rf /var/lib/apt/lists/*

# Pre-initialize the Wine prefix as the neko user so it is ready at runtime
USER neko
RUN wine wineboot --init || true
USER root
```

:::tip
Most Windows applications require 32-bit libraries. Always run `dpkg --add-architecture i386` before installing Wine.
:::

## Running a specific application

Create a supervisord configuration file so neko starts your application automatically:

```ini title="app.conf"
[program:myapp]
environment=HOME="/home/%(ENV_USER)s",USER="%(ENV_USER)s",DISPLAY="%(ENV_DISPLAY)s"
command=wine /home/neko/.wine/drive_c/Program Files/MyApp/MyApp.exe
autorestart=true
priority=800
user=%(ENV_USER)s
stdout_logfile=/var/log/neko/myapp.log
stdout_logfile_maxbytes=100MB
stdout_logfile_backups=10
redirect_stderr=true
```

Mount it into the container at `/etc/neko/supervisord/app.conf`.

## docker-compose example

```yaml title="docker-compose.yaml"
services:
  neko:
    build: .           # uses the Dockerfile above
    restart: unless-stopped
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    volumes:
      - "./app.conf:/etc/neko/supervisord/app.conf"
      # persist the Wine prefix so installations survive restarts
      - "./wine-prefix:/home/neko/.wine"
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```

Fix permissions on the mounted wine prefix directory so the `neko` user (UID 1000) can write to it:

```bash
sudo chown -R 1000:1000 ./wine-prefix
```

## Audio support

Neko uses PulseAudio. Wine's audio backend will automatically pick up the `PULSE_SERVER` environment variable that Neko sets, so sound should work without any additional configuration. If you experience audio issues, ensure the `pulseaudio` package is available in your image (it is included in all official Neko base images).

## Using Winetricks

[Winetricks](https://github.com/Winetricks/winetricks) can be used to install additional Windows runtime dependencies (e.g. DirectX, .NET, Visual C++ redistributables) inside the Wine prefix:

```dockerfile title="Dockerfile (continued)"
RUN apt-get update && apt-get install -y --no-install-recommends \
        winetricks \
        cabextract \
        unzip \
    && rm -rf /var/lib/apt/lists/*

# Install common runtimes (run as neko user so they land in the correct prefix)
USER neko
RUN winetricks -q vcrun2019 dotnet48
USER root
```

## GPU acceleration with Wine

If you need GPU acceleration (e.g. for DirectX games or 3D applications), use a Neko NVIDIA image as your base and add the `--privileged` flag or the appropriate device mappings:

```dockerfile title="Dockerfile (NVIDIA)"
FROM ghcr.io/m1k1o/neko/nvidia-xfce:latest

RUN dpkg --add-architecture i386 \
    && apt-get update \
    && apt-get install -y --install-recommends winehq-stable \
    && rm -rf /var/lib/apt/lists/*
```

```yaml title="docker-compose.yaml (NVIDIA)"
services:
  neko:
    build: .
    restart: unless-stopped
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
      NVIDIA_VISIBLE_DEVICES: all
      NVIDIA_DRIVER_CAPABILITIES: all
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

## Troubleshooting

### `nodrv_CreateWindow` — no driver could be loaded

Wine requires an X11 display. Make sure your Wine application is started **from a supervisord program** (not from the Dockerfile `CMD` or `ENTRYPOINT`), so that Xorg is already running when Wine starts.

### Application crashes immediately

Run the container interactively and launch Wine manually to see the full error output:

```bash
docker exec -it neko bash
su - neko -c "DISPLAY=:99 wine /path/to/app.exe"
```

Check the Wine debug output by setting `WINEDEBUG`:

```bash
DISPLAY=:99 WINEDEBUG=+all wine /path/to/app.exe 2>&1 | head -100
```

### Missing DLLs or runtimes

Use `winetricks` to install required Windows runtimes. Common ones are:

| Runtime | Winetricks verb |
|---------|----------------|
| Visual C++ 2019 | `vcrun2019` |
| .NET Framework 4.8 | `dotnet48` |
| DirectX | `d3dx9` / `d3dx11_43` |
| Windows Media codecs | `wmp11` |

### 32-bit vs 64-bit Wine prefix

By default Wine creates a 64-bit prefix (`WINEARCH=win64`). Some older applications require a 32-bit prefix:

```bash
WINEARCH=win32 WINEPREFIX=/home/neko/.wine32 wine /path/to/app32.exe
```

Set `WINEARCH` and `WINEPREFIX` in the supervisord `environment=` line when needed.
