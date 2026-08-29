---
description: Example Docker Compose configurations for Neko.
---

# Examples

Here are some examples to get you started with Neko. You can use these examples as a reference to create your own configurations.

Every example below is also available as a self-contained, runnable folder in the [`examples/`](https://github.com/m1k1o/neko/tree/main/examples) directory of the repository, so you can browse, clone or download it directly. Each file is heavily commented, showing the available options and linking to the relevant documentation inline.

## Simple Browser {#simple-browser}

A basic Firefox setup. Includes commented volume mounts for persisting the browser profile or mounting your own pre-configured one, and a list of other browser images you can swap in.

Browse: [`examples/simple-browser`](https://github.com/m1k1o/neko/tree/main/examples/simple-browser)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/simple-browser/docker-compose.yaml
```

See also: [Persistent Browser Profile](/docs/v3/customization/browsers#persistent-profile) and [Browser Policy Files](/docs/v3/customization/browsers#policy-files).

## Kiosk Browser {#kiosk-browser}

Runs Firefox in kiosk mode (no address bar, tabs or browser chrome) and opens a fixed URL automatically on startup. This is useful for app-like setups such as streaming services or dashboards, and works by mounting a persistent copy of the Firefox supervisor config and overriding the command directly.

This setup is intentionally stateless - no browser profile is persisted, so logins are lost on every restart. See the comments in the example for how to add a persistent profile if you need to stay logged in.

Browse: [`examples/kiosk-browser`](https://github.com/m1k1o/neko/tree/main/examples/kiosk-browser)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/kiosk-browser/docker-compose.yaml
```

```ini title="firefox.conf" file=<rootDir>/examples/kiosk-browser/firefox.conf
```

For some workflows, passing the target URL directly to Firefox is more reliable than using homepage policies, because session restore can otherwise override the start page. See also: [Supervisord Configuration](/docs/v3/customization#supervisord).

## Nvidia Browser {#nvidia-browser}

Neko supports hardware acceleration using Nvidia GPUs. To use this feature, you need to have the Nvidia Container Toolkit installed on your system. You can find the installation instructions [here](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html). Check if your GPU supports hardware encoding with [this list](https://developer.nvidia.com/video-encode-decode-gpu-support-matrix).

This example accelerates both video encoding and browser rendering using the GPU. You can test if the GPU is used by running `nvtop` or `nvidia-smi`, which should show the GPU usage of both the browser and neko. In the browser, you can run the [WebGL Aquarium Demo](https://webglsamples.org/aquarium/aquarium.html) to test the GPU usage.

If you only want to accelerate the encoding, **not the browser rendering**, see the commented "ENCODE-ONLY OPTION" blocks in the example - they switch to the plain `firefox` image and a `videoconvert`-based pipeline instead of `nvidia-firefox` with `cudaupload`/`cudaconvert`.

Browse: [`examples/nvidia-browser`](https://github.com/m1k1o/neko/tree/main/examples/nvidia-browser)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/nvidia-browser/docker-compose.yaml
```

See available [Nvidia Docker Images](/docs/v3/installation/docker-images#nvidia).

## Intel Browser {#intel-browser}

Neko supports hardware acceleration using Intel GPUs via VAAPI. This requires the host to expose `/dev/dri` and have the Intel graphics driver installed.

This example accelerates both video encoding and browser rendering using the `intel-firefox` image. If you only want to accelerate the encoding, **not the browser rendering**, see the commented "ENCODE-ONLY OPTION" in the example, which switches to the plain `firefox` image.

Browse: [`examples/intel-browser`](https://github.com/m1k1o/neko/tree/main/examples/intel-browser)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/intel-browser/docker-compose.yaml
```

See available [Intel Docker Images](/docs/v3/installation/docker-images#intel).

## Raspberry Pi Browser {#raspberry-pi-browser}

Firefox tuned for a Raspberry Pi (or similar ARM SBC). Works out of the box with software rendering/encoding. See the commented "GPU ACCELERATION OPTION" in the example for enabling hardware-accelerated encoding via the Broadcom VideoCore V4L2 M2M encoder, available on Raspberry Pi 3/4 (not Pi 5).

Browse: [`examples/raspberry-pi-browser`](https://github.com/m1k1o/neko/tree/main/examples/raspberry-pi-browser)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/raspberry-pi-browser/docker-compose.yaml
```

## ARM64 Browser {#arm64-browser}

Firefox on a generic ARM64 host (e.g. Apple M1/M2 under virtualization, AWS Graviton, Oracle Cloud ARM free tier). The same multi-arch image used on amd64 works here - no special image tag is required.

DRM (Widevine) support is limited on ARM64 and needs extra setup for protected streaming content, see [DRM for ARM64](/docs/v3/customization/browsers#arm64-drm). If your device exposes a V4L2 M2M hardware encoder, you can reuse the pipeline from the [Raspberry Pi Browser](#raspberry-pi-browser) example.

Browse: [`examples/arm64-browser`](https://github.com/m1k1o/neko/tree/main/examples/arm64-browser)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/arm64-browser/docker-compose.yaml
```

See supported architectures and per-app availability in the [Availability Matrix](/docs/v3/installation/docker-images#availability).

## TURN Server {#turn-server}

WebRTC needs a direct connection between the client and the server. When that is not possible (e.g. both sides are behind restrictive NATs or firewalls), a [TURN server](/docs/v3/configuration/webrtc#iceservers) relays the traffic instead. This example runs a [Coturn](https://github.com/coturn/coturn) TURN server alongside Neko.

Browse: [`examples/turn-server`](https://github.com/m1k1o/neko/tree/main/examples/turn-server)

```yaml title="docker-compose.yaml" file=<rootDir>/examples/turn-server/docker-compose.yaml
```

Replace `<MY-COTURN-SERVER>` with your LAN or Public IP address, and allow ports `49160-49200/udp` and `3478/tcp`.
