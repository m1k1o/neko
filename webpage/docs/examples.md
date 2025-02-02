---
sidebar_position: 7
---

# Examples

## Firefox

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:firefox"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## Chromium

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:chromium"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## VLC

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:vlc"
    restart: "unless-stopped"
    shm_size: "2gb"
    volumes:
      - "<your-video-folder>:/video" 
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## Not using docker?

You can execute `neko --help` to see available arguments. In [Dockerfile](https://github.com/m1k1o/neko/blob/master/.docker/base/Dockerfile) you can find required dependencies and install them manually.
