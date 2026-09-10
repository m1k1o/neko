---
sidebar_label: "Chromium"
description: "Customize the supported Chromium runtime."
---

# Chromium Customization

M1 supports only the Chromium image. Its profile directory is
`/home/neko/.config/chromium`; managed policies are read from
`/etc/chromium/policies/managed/policies.json`.

## Persistent profile

Mount a host directory when bookmarks, extensions and browser settings must
survive a container replacement:

```yaml title="docker-compose.yaml"
services:
  neko:
    image: ghcr.io/m1k1o/neko/chromium:latest
    shm_size: 2gb
    volumes:
      - ./profile:/home/neko/.config/chromium
```

The container user has UID/GID `1000`; grant that identity access to the host
directory before starting the container.

## Managed policies

To use your own Chromium policy, mount it read-only:

```yaml title="docker-compose.yaml"
services:
  neko:
    volumes:
      - ./policies.json:/etc/chromium/policies/managed/policies.json:ro
```

The image policy and preferences files in `apps/chromium/` are the source of
the defaults. Keep proxy credentials out of policies and command lines; use
the supported authenticated outbound-proxy configuration instead.
