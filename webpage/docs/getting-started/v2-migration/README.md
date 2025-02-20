---
sidebar_position: 8
---

# V2 Migration Guide

Currently, Neko is in compatibility mode, meaning that as soon as a single V2 configuration option is set, the legacy mode is enabled. This approach allows for a smooth transition from V2 to V3, where it does not expose the V2 API for new users but still allows existing users who use the old configuration to continue using it as before.

The legacy mode can be explicitly enabled or disabled by setting the `NEKO_LEGACY` environment variable to `true` or `false`.

:::tip
You can migrate to a new configuration even if you are using a V2 client. Just make sure to set the `NEKO_LEGACY` environment variable to `true`.
:::

:::info Built-in Client
When using Neko in a container with a built-in client, the client will always be compatible with the server regardless of what configuration is used.
:::

## Configuration

V3 is compatible with V2 configuration options when legacy support is enabled. You should be able to run V3 with the V2 configuration without any issues.

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

See the V3 [configuration options](/docs/v3/getting-started/configuration).

### WebRTC Video

See the V3 configuration options for the [WebRTC Video](/docs/v3/getting-started/configuration/capture#webrtc-video).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_DISPLAY`                        | `NEKO_CAPTURE_VIDEO_DISPLAY` and `NEKO_DESKTOP_DISPLAY`, <br /> consider using `DISPLAY` env variable if both should be the same |
| `NEKO_VIDEO_CODEC`                    | `NEKO_CAPTURE_VIDEO_CODEC`                                |
| `NEKO_AV1=true` *deprecated*          | `NEKO_CAPTURE_VIDEO_CODEC=av1`                            |
| `NEKO_H264=true` *deprecated*         | `NEKO_CAPTURE_VIDEO_CODEC=h264`                           |
| `NEKO_VP8=true` *deprecated*          | `NEKO_CAPTURE_VIDEO_CODEC=vp8`                            |
| `NEKO_VP9=true` *deprecated*          | `NEKO_CAPTURE_VIDEO_CODEC=vp9`                            |
| `NEKO_VIDEO`                          | `NEKO_CAPTURE_VIDEO_PIPELINE`, V3 allows multiple video pipelines |
| `NEKO_VIDEO_BITRATE`                  | **removed**, use custom pipeline instead                  |
| `NEKO_HWENC`                          | **removed**, use custom pipeline instead                  |
| `NEKO_MAX_FPS`                        | **removed**, use custom pipeline instead                  |


:::warning
V2 did not have client-side cursor support, the cursor was always part of the video stream. In V3, the cursor is sent separately from the video stream. Therefore, when using legacy configuration, there will be two video streams created, one with the cursor (for V2 clients) and one without the cursor (for V3 clients). Please consider using new configuration options if this is not the desired behavior.
:::

### WebRTC Audio

See the V3 configuration options for the [WebRTC Audio](/docs/v3/getting-started/configuration/capture#webrtc-audio).

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

### Broadcast

See the V3 configuration options for the [Broadcast](/docs/v3/getting-started/configuration/capture#broadcast).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_BROADCAST_PIPELINE`             | `NEKO_CAPTURE_BROADCAST_PIPELINE`                         |
| `NEKO_BROADCAST_URL`                  | `NEKO_CAPTURE_BROADCAST_URL`                              |
| `NEKO_BROADCAST_AUTOSTART`            | `NEKO_CAPTURE_BROADCAST_AUTOSTART`                        |

### Desktop

See the V3 configuration options for the [Desktop](/docs/v3/getting-started/configuration/desktop).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_SCREEN`                         | `NEKO_DESKTOP_SCREEN`                                     |

### Authentication

See the V3 configuration options for the [Authentication](/docs/v3/getting-started/configuration/authentication).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_PASSWORD`                       | `NEKO_MEMBER_MULTIUSER_USER_PASSWORD` with `NEKO_MEMBER_PROVIDER=multiuser` |
| `NEKO_PASSWORD_ADMIN`                 | `NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD` with `NEKO_MEMBER_PROVIDER=multiuser` |

In order for the legacy authentication to work, you need to set [Multi-user](http://localhost:3000/docs/v3/getting-started/configuration/authentication#multi-user-provider).

:::warning
V2 clients might not be compatible with any other authentication provider than the `multiuser`.
:::

### WebRTC

See the V3 configuration options for the [WebRTC](/docs/v3/getting-started/configuration/webrtc).

| **V2 Configuration**                  | **V3 Configuration**                                      |
|---------------------------------------|-----------------------------------------------------------|
| `NEKO_NAT1TO1`                        | `NEKO_WEBRTC_NAT1TO1`                                     |
| `NEKO_TCPMUX`                         | `NEKO_WEBRTC_TCPMUX`                                      |
| `NEKO_UDPMUX`                         | `NEKO_WEBRTC_UDPMUX`                                      |
| `NEKO_ICELITE`                        | `NEKO_WEBRTC_ICELITE`                                     |
| `NEKO_ICESERVERS` or `NEKO_ICESERVER` | `NEKO_WEBRTC_ICESERVERS_FRONTEND` and `NEKO_WEBRTC_ICESERVERS_BACKEND`, <br /> V3 allows separate ICE servers for frontend and backend |
| `NEKO_IPFETCH`                        | `NEKO_WEBRTC_IP_RETRIEVAL_URL`                            |
| `NEKO_EPR`                            | `NEKO_WEBRTC_EPR`                                         |

### Full V2 Configuration Reference

Here is a full list of all the configuration options available in Neko V2 that are still available in Neko V3 with legacy support enabled.

import Configuration from '@site/src/components/Configuration';
import configOptions from './help.json';

<Configuration configOptions={configOptions} />

See the full [V3 configuration reference](/docs/v3/getting-started/configuration/#full-configuration-reference) for more details.

## API

V3 is compatible with the V2 API when legacy support is enabled. There was specifically created a compatibility layer (legacy API) that allows V2 clients to connect to V3. The legacy API is enabled by default, but it can be disabled if needed. In later versions, the legacy API will be removed.

### Authentication

In V2 there was only one authentication provider available, as in V3 called the `multiuser` provider. The API knew based on the provided password (as `?pwd=` query string) if the user is an admin or not.

Since V3 handles authentication differently (see [API documentation](/docs/v3/api/neko-api#authentication)), there has been added `?usr=` query string to the API to specify the username. The password is still provided as `?pwd=` query string. The `?usr=` query string is still optional, if not provided, the API will generate a random username.

For every request in the legacy API, a new user session is created based on the `?usr=&pwd=` query string. The session is destroyed after the API request is completed. So for HTTP API requests, the sessions are short-lived but for WebSocket API requests, the session is kept alive until the WebSocket connection is closed.

Only the `multiuser` provider (or the `noauth` provider) is supported without specifying the `?usr=` query string.

### WebSocket Messages

Since WebSocket messages are not user-facing API, there exists no migration guide for them. When the legacy API is enabled, the user connects to the `/ws` endpoint and is handled by the compatibility layer V2 API. The V3 API is available at the `/api/ws` endpoint.

### WebRTC API

Since the WebRTC API is not user-facing API, there exists no migration guide for it. It has been changed to Big Endian format (previously Little Endian) to allow easier manipulation on the client side. 
V2 created a new data channel on the client side, V3 creates a new data channel on the server side. That means, the server just listens for a new data channel from the client and accepts it with the legacy API handler. It overwrites the existing V3 data channel with the legacy one.

### HTTP API

The V2 version had a very limited HTTP API, the V3 API is much more powerful and flexible. See the [API documentation](/docs/v3/api/neko-api) for more details.

#### GET `/health`

Migrated to the [Health](/docs/v3/api/healthcheck) endpoint for server health checks.

Returns `200 OK` if the server is running.

#### GET `/stats`

Migrated to the [Stats](/docs/v3/api/stats) endpoint for server statistics and the [List Sessions](/docs/v3/api/sessions-get) endpoint for the session list.

Returns a JSON object with the following structure:

```json
{
  // How many connections are currently active
  "connections": 0,
  // Who is currently having a session (empty if no one)
  "host": "<session_id>",
  // List of currently connected users
  "members": [
    {
      "session_id": "<session_id>",
      "displayname": "Name",
      "admin": true,
      "muted": false,
    }
  ],
  // List of banned IPs and who banned them as a session_id
  "banned": {
    "<ip>": "<session_id>"
  },
  // List of locked resources and who locked them as a session_id
  "locked": {
    "<resource>": "<session_id>"
  },
  // Server uptime
  "server_started_at": "2021-01-01T00:00:00Z",
  // When was the last admin or user left the session
  "last_admin_left_at": "2021-01-01T00:00:00Z",
  "last_user_left_at": "2021-01-01T00:00:00Z",
  // Whether the control protection or implicit control is enabled
  "control_protection": false,
  "implicit_control": false,
}
```

#### GET `/screenshot.jpg`

Migrated to the [Screenshot](/docs/v3/api/screen-shot-image) endpoint for taking screenshots.

Returns a screenshot of the desktop as a JPEG image.

#### GET `/file`

The whole functionality of file transfer has been moved to a [File Transfer Plugin](/docs/v3/getting-started/configuration/plugins#file-transfer-plugin).
