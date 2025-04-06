import { Def, Opt } from '@site/src/components/Anchor';
import { ConfigurationTab } from '@site/src/components/Configuration';
import configOptions from './help.json';

# Configuration

Neko uses the [Viper](https://github.com/spf13/viper) library to manage configuration. The configuration file is optional and is not required for Neko to run. If a configuration file is present, it will be read in and merged with the default configuration values.

The merge order is as follows:

- Default configuration values
- Configuration file
- Environment variables
- Command-line arguments

<details>
  <summary>Example merging order</summary>

```bash
# Default Value: 127.0.0.1:8080

# Config File
cat config.yaml <<EOF
server:
  bind: "127.0.0.1:8081"
EOF

# Environment Variable
export NEKO_SERVER_BIND=127.0.0.1:8082

# Command-line Argument
./neko -config=config.yaml -server.bind=127.0.0.1:8083
```

The final value of `server.bind` will be `127.0.0.1:8083`.

</details>

## Configuration File {#file}

You have multiple ways to specify the configuration file for the neko server:

- Command-line argument: `-config=/path/to/config.yaml`
- Environment variable: `NEKO_CONFIG=/path/to/config.yaml`
- Place the `neko.yaml` file in the same directory as the neko binary.
- Place the `neko.yaml` file to `/etc/neko/neko.yaml` *(ideal for Docker containers)*.

The configuration file can be specified in YAML, JSON, TOML, HCL, envfile, and Java properties format. Throughout the documentation, we will use the YAML format.

<details>
  <summary>Example configuration files</summary>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="yaml" label="YAML">

    ```yaml title="config.yaml"
    capture:
      screencast:
        enabled: false

    server:
      pprof: true

    desktop:
      screen: "1920x1080@60"

    member:
      provider: "multiuser"
      multiuser:
      admin_password: "admin"
      user_password: "neko"

    session:
      merciful_reconnect: true
      implicit_hosting: false
      inactive_cursors: true
      cookie:
        enabled: false

    webrtc:
      icelite: true
      iceservers:
        # Backend servers are ignored if icelite is true.
        backend:
          - urls: [ stun:stun.l.google.com:19302 ]
        frontend:
          - urls: [ stun:stun.l.google.com:19305 ]
    ```
  
  </TabItem>

  <TabItem value="json" label="JSON">

    ```json title="config.json"
    {
      "capture": {
        "screencast": {
          "enabled": false
        }
      },
      "server": {
        "pprof": true
      },
      "desktop": {
        "screen": "1920x1080@60"
      },
      "member": {
        "provider": "multiuser",
        "multiuser": {
          "admin_password": "admin",
          "user_password": "neko"
        }
      },
      "session": {
        "merciful_reconnect": true,
        "implicit_hosting": false,
        "inactive_cursors": true,
        "cookie": {
          "enabled": false
        }
      },
      "webrtc": {
        "icelite": true,
        "iceservers": {
          "backend": [
            {
              "urls": [ "stun:stun.l.google.com:19302" ]
            }
          ],
          "frontend": [
            {
              "urls": [ "stun:stun.l.google.com:19305" ]
            }
          ]
        }
      }
    }
    ```

  </TabItem>
  <TabItem value="toml" label="TOML">

    ```toml title="config.toml"
    [capture.screencast]
    enabled = false

    [server]
    pprof = true

    [desktop]
    screen = "1920x1080@60"

    [member]
    provider = "multiuser"

    [member.multiuser]
    admin_password = "admin"
    user_password = "neko"

    [session]
    merciful_reconnect = true
    implicit_hosting = false
    inactive_cursors = true

    [session.cookie]
    enabled = false

    [webrtc]
    icelite = true

    [[webrtc.iceservers.backend]]
    urls = [ "stun:stun.l.google.com:19302" ]

    [[webrtc.iceservers.frontend]]
    urls = [ "stun:stun.l.google.com:19305" ]
    ```

  </TabItem>

  <TabItem value="hcl" label="HCL">

    ```hcl title="config.hcl"
    capture {
      screencast {
        enabled = false
      }
    }

    server {
      pprof = true
    }

    desktop {
      screen = "1920x1080@60"
    }

    member {
      provider = "multiuser"

      multiuser {
        admin_password = "admin"
        user_password = "neko"
      }
    }

    session {
      merciful_reconnect = true
      implicit_hosting = false
      inactive_cursors = true

      cookie {
        enabled = false
      }
    }

    webrtc {
      icelite = true

      iceservers {
        backend {
          urls = [ "stun:stun.l.google.com:19302" ]
        }

        frontend {
          urls = [ "stun:stun.l.google.com:19305" ]
        }
      }
    }
    ```

  </TabItem>

  <TabItem value="envfile" label="Envfile">

    ```env title=".env"
    CAPTURE_SCREENCAST_ENABLED=false

    SERVER_PPROF=true

    DESKTOP_SCREEN=1920x1080@60

    MEMBER_PROVIDER=multiuser
    MEMBER_MULTIUSER_ADMIN_PASSWORD=admin
    MEMBER_MULTIUSER_USER_PASSWORD=neko

    SESSION_MERCIFUL_RECONNECT=true
    SESSION_IMPLICIT_HOSTING=false
    SESSION_INACTIVE_CURSORS=true
    SESSION_COOKIE_ENABLED=false

    WEBRTC_ICELITE=true

    WEBRTC_ICESERVERS_BACKEND="[{"urls":["stun:stun.l.google.com:19302"]}]"
    WEBRTC_ICESERVERS_FRONTEND="[{"urls":["stun:stun.l.google.com:19305"]}]"
    ```

  </TabItem>

  <TabItem value="properties" label="Java Properties">

    ```properties title="config.properties"
    capture.screencast.enabled = false

    server.pprof = true

    desktop.screen = 1920x1080@60

    member.provider = multiuser
    member.multiuser.admin_password = admin
    member.multiuser.user_password = neko

    session.merciful_reconnect = true
    session.implicit_hosting = false
    session.inactive_cursors = true
    session.cookie.enabled = false

    webrtc.icelite = true

    webrtc.iceservers.backend[0].urls[0] = stun:stun.l.google.com:19302
    webrtc.iceservers.frontend[0].urls[0] = stun:stun.l.google.com:19305
    ```

  </TabItem>

</Tabs>

</details>

## Room Configuration {#session}

This is the initial configuration of the room that can be modified by an admin in real-time.

<ConfigurationTab options={configOptions} filter={[
  'session.private_mode',
  'session.locked_logins',
  'session.locked_controls',
  'session.control_protection',
  'session.implicit_hosting',
  'session.inactive_cursors',
  'session.merciful_reconnect',
  'session.heartbeat_interval',
]} comments={false} />

- <Def id="session.private_mode" /> whether private mode is enabled, users do not receive the room video or audio.
- <Def id="session.locked_logins" /> whether logins are locked for users, admins can still login.
- <Def id="session.locked_controls" /> whether controls are locked for users, admins can still control.
- <Def id="session.control_protection" /> users can gain control only if at least one admin is in the room.
- <Def id="session.implicit_hosting" /> automatically grants control to a user when they click on the screen, unless an admin has locked the controls.
- <Def id="session.inactive_cursors" /> whether to show inactive cursors server-wide (only for users that have it enabled in their profile).
- <Def id="session.merciful_reconnect" /> whether to allow reconnecting to the websocket even if the previous connection was not closed. This means that a new login can kick out the previous one.
- <Def id="session.heartbeat_interval" /> interval in seconds for sending a heartbeat message to the server. This is used to keep the connection alive and to detect when the connection is lost.

## Server Configuration {#server}

This is the configuration of the neko server.

<ConfigurationTab options={configOptions} filter={[
  'server.bind',
  'server.cert',
  'server.key',
  'server.cors',
  'server.metrics',
  'server.path_prefix',
  'server.pprof',
  'server.proxy',
  'server.static',
]} comments={false} />

- <Def id="server.bind" /> address/port/socket to serve neko. For docker you might want to bind to `0.0.0.0` to allow connections from outside the container.
- <Def id="server.cert" /> and <Def id="server.key" /> paths to the SSL cert and key used to secure the neko server. If both are empty, the server will run in plain HTTP.
- <Def id="server.cors" /> is a list of allowed origins for CORS.
  - If empty, CORS is disabled, and only same-origin requests are allowed.
  - If `*` is present, all origins are allowed. Neko will respond always with the requested origin, not with `*` since [credentials are not allowed with wildcard](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS/Errors/CORSNotSupportingCredentials).
  - If a list of origins is present, only those origins are allowed for CORS.
- <Def id="server.metrics" /> when true, [prometheus](https://prometheus.io/docs/prometheus/latest/getting_started/) metrics are available at `/metrics`.
- <Def id="server.path_prefix" /> is the prefix for all HTTP requests. This is useful when running neko behind a reverse proxy and you want to serve neko under a subpath, e.g. `/neko`.
- <Def id="server.pprof" /> when true, the [pprof](https://golang.org/pkg/net/http/pprof/) endpoint is available at `/debug/pprof` for debugging and profiling. This should be disabled in production.
- <Def id="server.proxy" /> when true, neko will trust the `X-Forwarded-For` and `X-Real-IP` headers from the reverse proxy. Make sure your reverse proxy is configured to set these headers and never trust them when not behind a reverse proxy. See [Reverse Proxy Setup](/docs/v3/reverse-proxy-setup) for more information.
- <Def id="server.static" /> path to the directory containing the neko client files to serve. This is useful if you want to serve the client files on the same domain as the server.

## Logging Configuration {#log}

This is the configuration of the logging system.

<ConfigurationTab options={configOptions} filter={[
  'log.dir',
  'log.json',
  'log.level',
  'log.nocolor',
  'log.time',
]} comments={false} />

- <Def id="log.dir" /> directory to store logs. If empty, logs are written to stdout. This is useful when running neko in a container.
- <Def id="log.json" /> when true, logs are written in JSON format.
- <Def id="log.level" /> log level to set. Available levels are `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`, and `disabled`.
- <Def id="log.nocolor" /> when true, ANSI colors are disabled in non-JSON output. Accepts as well [`NO_COLOR` environment variable](https://no-color.org/).
- <Def id="log.time" /> time format used in logs. Available formats are `unix`, `unixms`, and `unixmicro`.

:::tip
Shortcut environment variable to enable DEBUG mode: `NEKO_DEBUG=true`
:::

## Full Configuration Reference {#full}

Here is a full configuration with default values as shown in the help command. Please refer to the sub-sections for more details.

<ConfigurationTab options={configOptions} heading={true} />

## Next Steps {#next}

import DocCardList from '@theme/DocCardList';

<DocCardList />
