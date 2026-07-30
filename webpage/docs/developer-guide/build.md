---
description: Building the Neko frontend and backend from source
---

# Building from Source

This page covers building the frontend and backend binaries from source, and building a custom Docker image (e.g. for unsupported architectures such as `linux/arm64`). The Dockerfiles in the repository are the authoritative reference for required dependencies.

## Frontend

**Requires:** [Node.js](https://nodejs.org/) 18+

```bash
cd client
npm install
npm run build
```

The production build is written to `client/dist/`.

## Backend

**Requires:** [Go](https://golang.org/) 1.25+ and the following system libraries (Debian/Ubuntu):

```bash
apt-get install -y --no-install-recommends \
    libx11-dev libxrandr-dev libxtst-dev libgtk-3-dev libxcvt-dev \
    libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev
```

Then build:

```bash
cd server
./build
```

The binary is written to `server/bin/neko`, with any plugins in `server/bin/plugins/`.

Pass `core` to skip building plugins:

```bash
./build core
```

## Building a Custom Docker Image

This is useful when you need to target an architecture that is not covered by the published images (e.g. `linux/arm64` / armv8), or when you want to package a custom application.

The repository uses a [Dockerfile template](https://github.com/m1k1o/neko/blob/master/Dockerfile.tmpl) that is pre-processed by a small Go utility before being passed to `docker build`. The template stitches together four independent stages:

| Stage | Source | Description |
|---|---|---|
| `server` | `server/` | Go binary + plugins |
| `client` | `client/` | Vue.js frontend |
| `runtime` | `runtime/` | Xorg, PulseAudio, GStreamer base |
| `application` | `apps/<name>/` | The browser / desktop app |

### Requirements

- [Go](https://golang.org/) 1.25+ (for the template pre-processor)
- [Node.js](https://nodejs.org/) 18+ (for the frontend, unless you supply a pre-built `client/dist/`)
- [Docker](https://www.docker.com/) with BuildKit enabled

### Steps

**1. Build the frontend** (skip if you already have `client/dist/`):

```bash
cd client
npm install
npm run build
cd ..
```

**2. Generate the Dockerfile from the template:**

```bash
go run utils/docker/main.go \
  -i Dockerfile.tmpl \
  -o Dockerfile \
  -client client/dist
```

The `-client` flag tells the pre-processor to copy the already-built `client/dist/` directory instead of rebuilding the frontend inside Docker.

**3. Build the image:**

```bash
docker build -t local/neko -f Dockerfile .
```

To cross-compile for a different architecture, add `--platform`:

```bash
docker build --platform linux/arm64 -t local/neko -f Dockerfile .
```

:::note
The `runtime/Dockerfile` is the authoritative reference for Xorg/PulseAudio/GStreamer dependencies. Setting up the runtime outside of Docker is non-trivial; using Docker is strongly recommended.
:::

### Choosing an Application

The generated `Dockerfile` includes the runtime base only. To add an application (browser, desktop, etc.), append the relevant `apps/<name>/Dockerfile` contents, or build a multi-stage image that adds the app on top of your `local/neko` base:

```dockerfile
FROM local/neko
# Install your application here — follow the upstream docs for non-interactive installation.
RUN apt-get update && apt-get install -y --no-install-recommends firefox-esr
COPY apps/firefox/supervisord.conf /etc/supervisord.conf
```
