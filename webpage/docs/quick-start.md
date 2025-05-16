---
description: A quick guide to get you started with Neko.
---

# Quick Start

Neko is easy to use and requires no technical expertise to get started. All you need to do is download the Docker image and you're ready to go:

1. Deploy a server or VPS with a public IP address.

    **Recommended Specs:**
    
    | Resolution  | Cores | Ram   | Recommendation   |
    |-------------|-------|-------|------------------|
    | 1024×576@30 | 2     | 2gb   | Not Recommended  |
    | 1280x720@30 | 4     | 3gb   | Good Performance |
    | 1280x720@30 | 6     | 4gb   | Recommended      |
    | 1280x720@30 | 8     | 4gb+  | Best Performance |
  

    :::danger[Why are the specs so high?]
    If you think about it, you have to run a full desktop, a browser (a resource hog on its own) *and* encode/transmit the desktop, there's a lot going on and so it demands some power.
    :::

    :::note
    The admin can change the resolution in the GUI.
    :::

2. [Login via SSH](https://www.digitalocean.com/docs/droplets/how-to/connect-with-ssh/).

3. Install [Docker](https://docs.docker.com/get-docker/):
    ```shell
    curl -sSL https://get.docker.com/ | CHANNEL=stable bash
    ```

4. Install [Docker Compose Plugin](https://docs.docker.com/compose/install/linux/):
    ```shell
    sudo apt-get update
    sudo apt-get install docker-compose-plugin
    ```

5. Download the docker compose file and start it:
    ```shell
    wget https://raw.githubusercontent.com/m1k1o/neko/master/docker-compose.yaml
    sudo docker compose up -d
    ```

    :::note
    If you want to run Neko on your local network, you have to add `NEKO_NAT1TO1: <your-local-ip>` to the `docker-compose.yaml` file.
    :::

6. Visit the server's IP address in your browser and log in, the default password is `neko`.

:::tip
Run `nano docker-compose.yaml` to edit the settings, then press `ctrl+x` to exit and save the file.
:::

## Well-known cloud providers {#providers}

* [Hetzner Cloud](https://www.hetzner.com/cloud)
* [Digital Ocean](https://www.digitalocean.com/)
* [Contabo](https://contabo.com/)
* [Linode](https://www.linode.com/)
* [Vultr](https://www.vultr.com/)
