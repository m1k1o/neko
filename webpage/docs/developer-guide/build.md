---
description: Building the Neko frontend and backend from source
---

# Building from Source

This page covers building the frontend and backend binaries from source. The Dockerfiles in the repository are the authoritative reference for required dependencies.

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
