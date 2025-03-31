---
sidebar_label: "User Interface"
description: "Customize the Neko user interface with your own UI files."
---

# Customizing the UI

Currently there is no configuration for customizing the UI of Neko. You need to modify the source code to change the UI.

```bash
# Clone the repository
git clone https://github.com/m1k1o/neko
# Change to the client directory
cd neko/client
# Install the dependencies
npm install
# Build the project
npm run build
```

You can mount your newly created UI files to the container to `/var/www` to overwrite the default files. The Neko web server will automatically reload the new files when they are changed. You can use the following command to mount your new UI files to the container:

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    # highlight-start
    volumes:
      - "./client/dist:/var/www"
    # highlight-end
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```
