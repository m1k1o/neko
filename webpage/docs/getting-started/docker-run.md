---
sidebar_position: 3
---

# Running Neko in Docker

To start a basic Neko container, use the following command:

```sh
docker run -d --rm \
  -p 8080:8080 \
  -p 56000-56100:56000-56100/udp \
  -e NEKO_WEBRTC_EPR=56000-56100 \
  -e NEKO_WEBRTC_NAT1TO1=127.0.0.1 \
  -e NEKO_MEMBER_MULTIUSER_USER_PASSWORD=neko \
  -e NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD=admin \
  ghcr.io/m1k1o/neko/firefox:latest
```

### Explanation

- `-d` - Run the container in the background.
- `--rm` - Automatically remove the container when it exits.
- `-p 8080:8080` - Map the container's port `8080` to the host's port `8080`.
- `-p 56000-56100:56000-56100/udp` - Map the container's UDP ports `56000-56100` to the host's ports `56000-56100`.
- `-e NEKO_WEBRTC_EPR=56000-56100` - Set the WebRTC endpoint range, must match the mapped ports.
  - See [WebRTC Ephemeral Port Range](/docs/v3/getting-started/configuration/webrtc#ephemeral-udp-port-range) for more information about this setting.
- `-e NEKO_WEBRTC_NAT1TO1=127.0.0.1` - Set the NAT1TO1 IP address.
  - To test only on the local computer, use `127.0.0.1`.
  - To use it in a private network, use the host's IP address (e.g., `192.168.1.5`).
  - To use it in a public network, you need to correctly set up port forwarding on your router and remove this env variable.
  - See [WebRTC Server IP Address](/docs/v3/getting-started/configuration/webrtc#server-ip-address) for more information about this setting.
- `-e NEKO_MEMBER_MULTIUSER_USER_PASSWORD=neko` - Set the password for the user account.
- `-e NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD=admin` - Set the password for the admin account.
  - See [Multiuser Configuration](/docs/v3/getting-started/configuration/authentication#multi-user-provider) for more information about this setting.
- `ghcr.io/m1k1o/neko/firefox:latest` - The Neko Docker image to use.
  - See [Available Docker Images](/docs/v3/getting-started/docker-images) for more information about the available images.

Now, open your browser and go to: `http://localhost:8080`. You should see the Neko interface.

## Using Docker Compose

You can also use Docker Compose to run Neko. It is preferred to use Docker Compose for running Neko in production because you can easily manage the container, update it, and configure it.

Create a `docker-compose.yml` file with the following content:

```yaml title="docker-compose.yml"
services:
  neko:
    image: ghcr.io/m1k1o/neko/firefox:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "56000-56100:56000-56100/udp"
    environment:
      NEKO_WEBRTC_EPR: "56000-56100"
      NEKO_WEBRTC_NAT1TO1: "127.0.0.1"
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: "neko"
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: "admin"
```

Then, run the following command:

```sh
docker compose up -d
```

To stop Neko, run:

```sh
docker compose down
```

To update Neko, run:

```sh
docker compose pull
docker compose up -d
```

Learn more about [how compose works](https://docs.docker.com/compose/intro/compose-application-model/).

:::note
You need to be in the same directory as the `docker-compose.yml` file to run the `docker compose` commands.
:::
