# Quick Start (WIP)

1. Deploy a server or VPS

    **Recommended Specs:**
    
    | Resolution  | Cores | Ram   | Recommendation   |
    |-------------|-------|-------|------------------|
    | 1024×576@30 | 2     | 2gb   | Not Recommended  |
    | 1280x720@30 | 4     | 4gb   | Good Performance |
    | 1280x720@30 | 6     | 4-6gb | Recommended      |
    | 1280x720@30 | 8     | 8gb+  | Best Performance |
  
    *Why are the specs so high?* : If you think about it, you have to run a full desktop, a browser (a resource hog on its own) *and* encode/transmit the desktop, there's a lot going on and so it demands some power.

    *Note:* changing the resolution will require additional setup 

2. [Login via SSH](https://www.digitalocean.com/docs/droplets/how-to/connect-with-ssh/)

3. Install Docker
    ```shell
    curl -sSL https://get.docker.com/ | CHANNEL=stable bash
    ```
4. Run these commands:
    ```shell
    sudo ufw allow 80/tcp # if you have ufw installed/enabled
    sudo ufw allow 59000:59100/udp
    wget https://raw.githubusercontent.com/nurdism/neko/master/docker-compose.yaml
    sudo docker-compose up -d
    ```
5. Visit the IP address server in your browser and login, the default password is `neko`

> 💡 **Protip**: Run `nano docker-compose.yaml` to edit the settings, then press `ctrl+x` to exit and save the file.