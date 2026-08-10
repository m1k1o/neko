# Examples

Ready-to-use [Docker Compose](https://docs.docker.com/compose/) configurations for Neko. Each folder is self-contained and can be copied out and run on its own with `docker compose up`. All files are heavily commented, showing available options and linking to the relevant documentation inline.

These are the same examples shown in the [documentation](https://neko.m1k1o.net/docs/v3/installation/examples), kept here so they can be browsed, cloned or downloaded directly from the repository.

| Example | Description |
| --- | --- |
| [simple-browser](./simple-browser) | Firefox with commented options for persistent/custom profiles and other available browser images. |
| [kiosk-browser](./kiosk-browser) | Firefox in kiosk mode that opens a fixed URL on startup (e.g. a streaming service or dashboard). Stateless by default. |
| [nvidia-browser](./nvidia-browser) | Firefox with Nvidia GPU acceleration, with a commented "encode-only" option. |
| [intel-browser](./intel-browser) | Firefox with Intel (VAAPI) GPU acceleration, with a commented "encode-only" option. |
| [raspberry-pi-browser](./raspberry-pi-browser) | Firefox tuned for Raspberry Pi, with a commented hardware-encoding option. |
| [arm64-browser](./arm64-browser) | Firefox on generic ARM64 hosts, with links to DRM setup and GPU acceleration options. |
| [turn-server](./turn-server) | A Coturn TURN server running alongside Neko for WebRTC NAT traversal. |

