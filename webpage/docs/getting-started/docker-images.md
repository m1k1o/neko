---
sidebar_position: 4
---

# Available Docker Images

Neko as a standalone streaming server is available as a Docker image. But that is rarely interesting for general use. The real power of Neko is in its ability to accommodate custom applications in the virtual desktop environment. This is where the various flavors of Neko Docker images come in.

The base image is available at [`ghcr.io/m1k1o/neko/base`](https://ghcr.io/m1k1o/neko/base).

## Naming Convention

Neko Docker images are available on [GitHub Container Registry (GHCR)](https://github.com/m1k1o?tab=packages&repo_name=neko). The naming convention for Neko Docker images is as follows:

```
ghcr.io/m1k1o/neko/[<flavor>-]<application>:<version>
```

- `<flavor>` is the optional flavor of the image, see [Available Flavors](#available-flavors) for more information.
- `<application>` is the application name or base image, see [Available Applications](#available-applications) for more information.
- `<version>` is the [semantic version](https://semver.org/) of the image from the [GitHub tags](https://github.com/m1k1o/neko/tags). There is always a `latest` tag available.

An alternative registry is also available on [Docker Hub](https://hub.docker.com/r/m1k1o/neko), however, only images without flavor and with the latest version are available there.

```
m1k1o/neko:<application>
```

:::info
You should always prefer the GHCR registry with the ability to use flavors and specific versions.
:::

## Available Applications

The following applications are available as Neko Docker images:

### Firefox-based browsers

In comparison to Chromium-based browsers, Firefox-based browsers do not require additional capabilities or a bigger shared memory size to not crash.

| Icon | Name | Docker Image |
| ---- | ---- | ------------ |
| <img src="/img/icons/firefox.svg" width="60" height="60" /> | [Firefox](https://www.mozilla.org/firefox/) <br /> The open-source browser from Mozilla. | [`ghcr.io/m1k1o/neko/firefox`](https://ghcr.io/m1k1o/neko/firefox) |
| <img src="/img/icons/tor-browser.svg" width="60" height="60" /> | [Tor Browser](https://www.torproject.org/) <br /> A browser designed to access the Tor network for enhanced privacy. | [`ghcr.io/m1k1o/neko/tor-browser`](https://ghcr.io/m1k1o/neko/tor-browser) |
| <img src="/img/icons/waterfox.svg" width="60" height="60" /> | [Waterfox](https://www.waterfox.net/) <br /> A privacy-focused browser based on Firefox. | [`ghcr.io/m1k1o/neko/waterfox`](https://ghcr.io/m1k1o/neko/waterfox) |

### Chromium-based browsers

There are multiple flavors of Chromium-based browsers available as Neko Docker images.

They need `--cap-add=SYS_ADMIN` (see [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container) for more information) and extended shared memory size (`--shm-size=2g`) to work properly.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="docker-run" label="Docker run command">

    ```bash
    docker run \
      --cap-add=SYS_ADMIN \
      --shm-size=2g \
      ghcr.io/m1k1o/neko/chromium
    ```

  </TabItem>

  <TabItem value="docker-compose" label="Docker Compose configuration">

    ```yaml title="docker-compose.yaml"
    cap_add:
    - SYS_ADMIN
    shm_size: 2g
    ```

  </TabItem>
</Tabs>

| Icon | Name | Docker Image |
| ---- | ---- | ------------ |
| <img src="/img/icons/chromium.svg" width="60" height="60" /> | [Chromium](https://www.chromium.org/chromium-projects/) <br /> The open-source project behind Google Chrome. | [`ghcr.io/m1k1o/neko/chromium`](https://ghcr.io/m1k1o/neko/chromium) |
| <img src="/img/icons/google-chrome.svg" width="60" height="60" /> | [Google Chrome](https://www.google.com/chrome/) <br /> The most popular browser in the world. | [`ghcr.io/m1k1o/neko/google-chrome`](https://ghcr.io/m1k1o/neko/google-chrome) |
| <img src="/img/icons/ungoogled-chromium.svg" width="60" height="60" /> | [Ungoogled Chromium](https://ungoogled-software.github.io/) <br /> A fork of Chromium without Google integration. | [`ghcr.io/m1k1o/neko/ungoogled-chromium`](https://ghcr.io/m1k1o/neko/ungoogled-chromium) |
| <img src="/img/icons/microsoft-edge.svg" width="60" height="60" /> | [Microsoft Edge](https://www.microsoft.com/edge) <br/> The new Microsoft Edge is based on Chromium. | [`ghcr.io/m1k1o/neko/microsoft-edge`](https://ghcr.io/m1k1o/neko/microsoft-edge) |
| <img src="/img/icons/brave.svg" width="60" height="60" /> | [Brave](https://brave.com/) <br /> A privacy-focused browser. | [`ghcr.io/m1k1o/neko/brave`](https://ghcr.io/m1k1o/neko/brave) |
| <img src="/img/icons/vivaldi.svg" width="60" height="60" /> | [Vivaldi](https://vivaldi.com/) <br /> A highly customizable browser. | [`ghcr.io/m1k1o/neko/vivaldi`](https://ghcr.io/m1k1o/neko/vivaldi) |
| <img src="/img/icons/opera.svg" width="60" height="60" /> | [Opera](https://www.opera.com/)* <br /> A fast and secure browser. | [`ghcr.io/m1k1o/neko/opera`](https://ghcr.io/m1k1o/neko/opera) |

\* requires extra steps to enable DRM, see instructions [here](https://www.reddit.com/r/operabrowser/wiki/opera/linux_widevine_config/). `libffmpeg` is already configured.

### Desktop Environments

These images feature a full desktop environment where you can install and run multiple applications, use window management, and more. This is useful for people who want to run multiple applications in a single container.

| Icon | Name | Docker Image |
| ---- | ---- | ------------ |
| <img src="/img/icons/xfce.svg" width="60" height="60" /> | [Xfce](https://xfce.org/) <br /> A lightweight desktop environment. | [`ghcr.io/m1k1o/neko/xfce`](https://ghcr.io/m1k1o/neko/xfce) |
| <img src="/img/icons/kde.svg" width="60" height="60" /> | [KDE Plasma](https://kde.org/plasma-desktop) <br /> A feature-rich desktop environment. | [`ghcr.io/m1k1o/neko/kde`](https://ghcr.io/m1k1o/neko/kde) |

### Other Applications

As it would be impossible to include all possible applications in the repository, a couple of the most popular ones that work well with Neko have been chosen. Custom images can be created by using the base image and installing the desired application.

| Icon | Name | Docker Image |
| ---- | ---- | ------------ |
| <img src="/img/icons/remmina.svg" width="60" height="60" /> | [Remmina](https://remmina.org/) <br /> A remote desktop client. | [`ghcr.io/m1k1o/neko/remmina`](https://ghcr.io/m1k1o/neko/remmina) |
| <img src="/img/icons/vlc.svg" width="60" height="60" /> | [VLC](https://www.videolan.org/vlc/) <br /> A media player. | [`ghcr.io/m1k1o/neko/vlc`](https://ghcr.io/m1k1o/neko/vlc) |

#### Remmina Configuration

To use Remmina with Neko, you can either pass the `REMMINA_URL=<proto>://[<username>[:<password>]@]server[:port]` environment variable (proto being `vnc`, `rdp` or `spice`):

```bash
docker run \
  -e REMMINA_URL=vnc://server:5900 \
  ghcr.io/m1k1o/neko/remmina
```

Or bind-mount a custom configuration file to `~/.local/share/remmina/path_to_profile.remmina`. Then pass the `REMMINA_PROFILE=<path_to_profile.remmina>` environment variable:

```ini title="default.remmina"
[remmina]
name=Default
protocol=VNC
server=server.local
port=5900
```

```bash
docker run \
  -v /path/to/default.remmina:/root/.local/share/remmina/default.remmina \
  -e REMMINA_PROFILE=/root/.local/share/remmina/default.remmina \
  ghcr.io/m1k1o/neko/remmina
```

#### VLC Configuration

To use VLC with Neko, you can either pass the `VLC_MEDIA=<url>` environment variable:

```bash
docker run \
  -e VLC_MEDIA=http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4 \
  ghcr.io/m1k1o/neko/vlc
```

You can also bind-mount your local media files to the container, create a custom playlist, and pass the `VLC_MEDIA=<path_to_playlist>` environment variable:

```bash
docker run \
  -v /path/to/media:/media \
  -e VLC_MEDIA=/media/playlist.xspf \
  ghcr.io/m1k1o/neko/vlc
```

:::tip
See [neko-apps](https://github.com/m1k1o/neko-apps) repository for more applications.
:::

## Available Flavors

:::danger Keep in Mind
Currently the focus is on AMD64 & CPU image (wihout any flavor). So the flavor images might not work as expected.
:::


The following flavors are available for Neko Docker images:

- `arm` - ARM64 and ARMv7 architecture support.
- `nvidia` - NVIDIA GPU support.
- `intel` - Intel GPU support.

:::note
Not all flavors are available for all applications. Since not all applications support ARM architecture or GPU acceleration, the flavors are only available where they make sense.
:::

### ARM

For ARM-based images (like Raspberry Pi - with GPU hardware acceleration, [Oracle Cloud ARM free tier](https://www.oracle.com/cloud/free/)). Currently, not all images are available for ARM, because not all applications are available for ARM. Please use the images below:

- [`ghcr.io/m1k1o/neko/arm-firefox`](https://ghcr.io/m1k1o/neko/arm-firefox)
- [`ghcr.io/m1k1o/neko/arm-chromium`](https://ghcr.io/m1k1o/neko/arm-chromium)
- [`ghcr.io/m1k1o/neko/arm-ungoogled-chromium`](https://ghcr.io/m1k1o/neko/arm-ungoogled-chromium)
- [`ghcr.io/m1k1o/neko/arm-vlc`](https://ghcr.io/m1k1o/neko/arm-vlc)
- [`ghcr.io/m1k1o/neko/arm-xfce`](https://ghcr.io/m1k1o/neko/arm-xfce)

The base image is available at [`ghcr.io/m1k1o/neko/arm-base`](https://ghcr.io/m1k1o/neko/arm-base).

### Intel

For images with VAAPI GPU hardware acceleration using Intel drivers use:

- [`ghcr.io/m1k1o/neko/intel-firefox`](https://ghcr.io/m1k1o/neko/intel-firefox)
- [`ghcr.io/m1k1o/neko/intel-chromium`](https://ghcr.io/m1k1o/neko/intel-chromium)
- [`ghcr.io/m1k1o/neko/intel-google-chrome`](https://ghcr.io/m1k1o/neko/intel-google-chrome)
- [`ghcr.io/m1k1o/neko/intel-ungoogled-chromium`](https://ghcr.io/m1k1o/neko/intel-ungoogled-chromium)
- [`ghcr.io/m1k1o/neko/intel-microsoft-edge`](https://ghcr.io/m1k1o/neko/intel-microsoft-edge)
- [`ghcr.io/m1k1o/neko/intel-brave`](https://ghcr.io/m1k1o/neko/intel-brave)
- [`ghcr.io/m1k1o/neko/intel-vivaldi`](https://ghcr.io/m1k1o/neko/intel-vivaldi)
- [`ghcr.io/m1k1o/neko/intel-opera`](https://ghcr.io/m1k1o/neko/intel-opera)
- [`ghcr.io/m1k1o/neko/intel-tor-browser`](https://ghcr.io/m1k1o/neko/intel-tor-browser)
- [`ghcr.io/m1k1o/neko/intel-remmina`](https://ghcr.io/m1k1o/neko/intel-remmina)
- [`ghcr.io/m1k1o/neko/intel-vlc`](https://ghcr.io/m1k1o/neko/intel-vlc)
- [`ghcr.io/m1k1o/neko/intel-xfce`](https://ghcr.io/m1k1o/neko/intel-xfce)
- [`ghcr.io/m1k1o/neko/intel-kde`](https://ghcr.io/m1k1o/neko/intel-kde)

The base image is available at [`ghcr.io/m1k1o/neko/intel-base`](https://ghcr.io/m1k1o/neko/intel-base).

### Nvidia

For images with Nvidia GPU hardware acceleration using EGL use:

- [`ghcr.io/m1k1o/neko/nvidia-firefox`](https://ghcr.io/m1k1o/neko/nvidia-firefox)
- [`ghcr.io/m1k1o/neko/nvidia-chromium`](https://ghcr.io/m1k1o/neko/nvidia-chromium)
- [`ghcr.io/m1k1o/neko/nvidia-google-chrome`](https://ghcr.io/m1k1o/neko/nvidia-google-chrome)
- [`ghcr.io/m1k1o/neko/nvidia-microsoft-edge`](https://ghcr.io/m1k1o/neko/nvidia-microsoft-edge)
- [`ghcr.io/m1k1o/neko/nvidia-brave`](https://ghcr.io/m1k1o/neko/nvidia-brave)

The base image is available at [`ghcr.io/m1k1o/neko/nvidia-base`](https://ghcr.io/m1k1o/neko/nvidia-base).

:::danger
There is a known issue with EGL and Chromium-based browsers, see [m1k1o/neko #279](https://github.com/m1k1o/neko/issues/279).
:::

