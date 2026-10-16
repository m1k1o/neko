# Migration from V2

Neko v3 accepts many V2 configuration options as deprecated aliases, which are mapped to their V3 equivalents. When both forms are set, the V2 option takes precedence. The V2 API and client compatibility layer have been removed; clients must use the V3 API.

## Docker Images {#docker}

Previously, neko was available primarily on Dockerhub as `m1k1o/neko`. While it stays as an option there, now the primary location is `ghcr.io/m1k1o/neko`.

ARM images were previously available as an flavor e.g. `m1k1o/neko:arm-firefox` or `ghcr.io/m1k1o/neko/arm-firefox`. Now, the ARM images are available as multi-arch images under the same tags as the amd64 images., e.g. `ghcr.io/m1k1o/neko/firefox`.

All applications available in the V2 images are also available in the V3 images. See the [Docker Images](/docs/v3/installation/docker-images) documentation for more details.

## Configuration {#config}

V3 continues to accept the V2 configuration aliases listed below. This does not enable or restore the V2 API.

The configuration in Neko V3 has been structured differently compared to V2. The V3 configuration is more modular and allows for more flexibility. The V3 configuration is split into multiple sections, each section is responsible for a specific part of the application. This allows for better organization and easier management of the configuration.

In order to migrate from V2 to V3, you need to update the configuration to the new format. The following table shows the mapping between the V2 and V3 configuration options.

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_LOGS=true`                      | `NEKO_LOG_DIR=/var/log/neko`, V3 allows specifying the log directory |
| `NEKO_CERT`                           | `NEKO_SERVER_CERT`                                        |
| `NEKO_KEY`                            | `NEKO_SERVER_KEY`                                         |
| `NEKO_BIND`                           | `NEKO_SERVER_BIND`                                        |
| `NEKO_PROXY`                          | `NEKO_SERVER_PROXY`                                       |
| `NEKO_STATIC`                         | `NEKO_SERVER_STATIC`                                      |
| `NEKO_PATH_PREFIX`                    | `NEKO_SERVER_PATH_PREFIX`                                 |
| `NEKO_CORS`                           | `NEKO_SERVER_CORS`                                        |
| `NEKO_LOCKS`                          | `NEKO_SESSION_LOCKED_CONTROLS` and `NEKO_SESSION_LOCKED_LOGINS`, <br /> V3 allows separate locks for controls and logins |
| `NEKO_IMPLICIT_CONTROL`               | `NEKO_SESSION_IMPLICIT_HOSTING`                           |
| `NEKO_CONTROL_PROTECTION`             | `NEKO_SESSION_CONTROL_PROTECTION`                         |
| `NEKO_HEARTBEAT_INTERVAL`             | `NEKO_SESSION_HEARTBEAT_INTERVAL`                         |
| `NEKO_FILE_TRANSFER_ENABLED`          | `NEKO_FILETRANSFER_ENABLED`                               |
| `NEKO_FILE_TRANSFER_PATH`             | `NEKO_FILETRANSFER_DIR`                                   |

See the V3 [configuration options](/docs/v3/configuration). For file transfer, see the [File Transfer Plugin](/docs/v3/configuration/plugins#filetransfer).

### WebRTC Video {#config.video}

See the V3 configuration options for the [WebRTC Video](/docs/v3/configuration/capture#video).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_DISPLAY`                        | `NEKO_CAPTURE_VIDEO_DISPLAY` and `NEKO_DESKTOP_DISPLAY`, <br /> consider using `DISPLAY` env variable if both should be the same |
| `NEKO_VIDEO_CODEC`                    | `NEKO_CAPTURE_VIDEO_CODEC`                                |
| `NEKO_AV1=true` *deprecated*          | `NEKO_CAPTURE_VIDEO_CODEC=av1`                            |
| `NEKO_H264=true` *deprecated*         | `NEKO_CAPTURE_VIDEO_CODEC=h264`                           |
| `NEKO_VP8=true` *deprecated*          | `NEKO_CAPTURE_VIDEO_CODEC=vp8`                            |
| `NEKO_VP9=true` *deprecated*          | `NEKO_CAPTURE_VIDEO_CODEC=vp9`                            |
| `NEKO_VIDEO`                          | `NEKO_CAPTURE_VIDEO_PIPELINE`, V3 allows multiple video pipelines |
| `NEKO_VIDEO_BITRATE`                  | **removed**, use [custom pipeline](/docs/v3/configuration/capture#video.gst_pipeline) instead |
| `NEKO_HWENC`                          | **removed**, use [custom pipeline](/docs/v3/configuration/capture#video.gst_pipeline) instead |
| `NEKO_MAX_FPS`                        | **removed**, use [custom pipeline](/docs/v3/configuration/capture#video.gst_pipeline) instead |


V2 video settings map to a single V3 video pipeline. The V3 client renders the cursor separately from the video stream.

### WebRTC Audio {#config.audio}

See the V3 configuration options for the [WebRTC Audio](/docs/v3/configuration/capture#audio).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_DEVICE`                         | `NEKO_CAPTURE_AUDIO_DEVICE`                               |
| `NEKO_AUDIO_CODEC`                    | `NEKO_CAPTURE_AUDIO_CODEC`                                |
| `NEKO_G722=true` *deprecated*         | `NEKO_CAPTURE_AUDIO_CODEC=g722`                           |
| `NEKO_OPUS=true` *deprecated*         | `NEKO_CAPTURE_AUDIO_CODEC=opus`                           |
| `NEKO_PCMA=true` *deprecated*         | `NEKO_CAPTURE_AUDIO_CODEC=pcma`                           |
| `NEKO_PCMU=true` *deprecated*         | `NEKO_CAPTURE_AUDIO_CODEC=pcmu`                           |
| `NEKO_AUDIO`                          | `NEKO_CAPTURE_AUDIO_PIPELINE`                             |
| `NEKO_AUDIO_BITRATE`                  | **removed**, use custom pipeline instead                  |

### Broadcast {#config.broadcast}

See the V3 configuration options for the [Broadcast](/docs/v3/configuration/capture#broadcast).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_BROADCAST_PIPELINE`             | `NEKO_CAPTURE_BROADCAST_PIPELINE`                         |
| `NEKO_BROADCAST_URL`                  | `NEKO_CAPTURE_BROADCAST_URL`                              |
| `NEKO_BROADCAST_AUTOSTART`            | `NEKO_CAPTURE_BROADCAST_AUTOSTART`                        |

### Desktop {#config.desktop}

See the V3 configuration options for the [Desktop](/docs/v3/configuration/desktop).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_SCREEN`                         | `NEKO_DESKTOP_SCREEN`                                     |

### Authentication {#config.auth}

See the V3 configuration options for the [Authentication](/docs/v3/configuration/authentication).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_PASSWORD`                       | `NEKO_MEMBER_MULTIUSER_USER_PASSWORD` with `NEKO_MEMBER_PROVIDER=multiuser` |
| `NEKO_PASSWORD_ADMIN`                 | `NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD` with `NEKO_MEMBER_PROVIDER=multiuser` |

The V2 password settings map to the [Multi-user](/docs/v3/configuration/authentication#member.multiuser) provider.

:::warning Limitation
V2 clients might not be compatible with any other authentication provider than the `multiuser`.
:::

### WebRTC {#config.webrtc}

See the V3 configuration options for the [WebRTC](/docs/v3/configuration/webrtc).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_NAT1TO1`                        | `NEKO_WEBRTC_NAT1TO1`                                     |
| `NEKO_TCPMUX`                         | `NEKO_WEBRTC_TCPMUX`                                      |
| `NEKO_UDPMUX`                         | `NEKO_WEBRTC_UDPMUX`                                      |
| `NEKO_ICELITE`                        | `NEKO_WEBRTC_ICELITE`                                     |
| `NEKO_ICESERVERS` or `NEKO_ICESERVER` | `NEKO_WEBRTC_ICESERVERS_FRONTEND` and `NEKO_WEBRTC_ICESERVERS_BACKEND`, <br /> V3 allows separate ICE servers for frontend and backend |
| `NEKO_IPFETCH`                        | `NEKO_WEBRTC_IP_RETRIEVAL_URL`                            |
| `NEKO_EPR`                            | `NEKO_WEBRTC_EPR`                                         |

### Full V2 Configuration Reference {#config.full}

Here is a full list of V2 configuration options that remain available as deprecated aliases in Neko V3.

import { ConfigurationTab } from '@site/src/components/Configuration';
import configOptions from './help.json';

<ConfigurationTab options={configOptions} heading={true} />

See the full [V3 configuration reference](/docs/v3/configuration/#full) for more details.

## API {#api}

The V2 API compatibility layer has been removed. Migrate clients to the [V3 API](/docs/v3/api); the V2 `/ws` endpoint and V2 HTTP endpoints are no longer available.

## Limitations {#limitations}

In v2, locks and muted users were managed using a simple map that tracked who set the lock and what was locked. In v3, locks are now implemented as setting options and no longer store the `session_id` of the user who applied the lock. As a result, if a client refreshes the page or reconnects, the lock information is lost, and the user who set the lock is displayed as `Somebody`.

Currently, the v3 has no native support for pipeline generation, meaning that the user has to manually specify the pipeline for video and audio if he wants to customize it. The v2 had a simple built-in support for setting the video bitrate, fps, audio bitrate, and hardware encoding. Since the introduction of multiple video pipelines in v3, this feature has been removed.
