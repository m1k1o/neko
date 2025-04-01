# Customization

When you want to customize the Neko virtual environment beyond the server configuration, you can do so by mounting your own files to the container that replace or extend the default files. This allows you to customize the desktop environment, browser settings, and more.

## Supervisord Configuration {#supervisord}

The Neko container uses [supervisord](https://supervisord.org/) to manage the processes inside the container. The search path for the configuration files is `/etc/neko/supervisord/<name>.conf`. You can mount your own `app.conf` file to the directory `/etc/neko/supervisord/` to add a new process to the container.

```config title="supervisord.conf"
[program:app]
environment=HOME="/home/%(ENV_USER)s",USER="%(ENV_USER)s",DISPLAY="%(ENV_DISPLAY)s"
command=/opt/path/to/my-app
stopsignal=INT
autorestart=true
priority=800
user=%(ENV_USER)s
stdout_logfile=/var/log/neko/app.log
stdout_logfile_maxbytes=100MB
stdout_logfile_backups=10
redirect_stderr=true
```

For example, with firefox, you can mount your own `firefox.conf` file to the directory `/etc/neko/supervisord/` to overwrite the default configuration file and modify the command that starts Firefox. Make sure to copy the default configuration file from the container to your local machine first:

```bash
# Create a container without starting it
docker create --name neko ghcr.io/m1k1o/neko/firefox:latest
# Copy the default configuration file to your local machine
docker cp neko:/etc/neko/supervisord/firefox.conf ./firefox.conf
# Remove the container
docker rm -f neko
```

Then, you can modify the configuration file to your liking and mount your new version to the container:

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
      - "./firefox.conf:/etc/neko/supervisord/firefox.conf"
    # highlight-end
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```

## Next Steps

import DocCardList from '@theme/DocCardList';
import {useCurrentSidebarCategory} from '@docusaurus/theme-common';

<DocCardList items={useCurrentSidebarCategory().items}/>
