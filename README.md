<div align="center">
<img src="https://github.com/nurdism/neko/raw/master/.github/logo.png" width="650" height="auto"/>
</div>

<div align="center">
<img src="https://github.com/nurdism/neko/raw/master/.github/demo.gif" width="650" height="auto"/>
</div>

# **n**.eko
This is a proof of concept project I threw together over the last few days, it's not perfect, but it looks nice. This uses web rtc to stream a desktop inside of a docker container, I made this because [rabb.it](https://en.wikipedia.org/wiki/Rabb.it) went under and my internet can't handle streaming and discord keeps crashing. I just want to watch anime with my friends ლ(ಠ益ಠლ) so I started digging throughout the net and found a few *kinda* clones, but non of them had the virtual browser, then I found [Turtus](https://github.com/Khauri/Turtus) and I was able to figure out the rest. 

This is by no means a fully featured clone of rabbit. The client has no concept of other peers. It has bugs, but for the most part it works. I'm not sure what the future holds for this. If I continue to use it and like it, I'll probably keep pushing updates to it. I'd be happy to accept PRs for any improvements. 

### Why n.eko?
I like cats (Neko is the Japanese word for cat), I'm a weeb/nerd, I own the domain [n.eko.moe](https://n.eko.moe/) and I love the logo /shrug

### Super easy mode setup
1. Deploy a Server/VPS

    *Recomended Specs:*
    
    | Resolution | Cores | Ram   | Recommendation   |
    |------------|-------|-------|------------------|
    | **576p**   | 2     | 2gb   | Not Recommended  |
    | **720p**   | 4     | 4gb   | Good Performance |
    | **720p**   | 6     | 4-6gb | Recommended      |
    | **720p+**  | 8     | 8gb+  | Best Performance |

2. [SSH into your box](https://www.digitalocean.com/docs/droplets/how-to/connect-with-ssh/)

3. Install Docker
    ```
    curl -sSL https://get.docker.com/ | CHANNEL=stable bash
    ```
4. Run these commands:
    ```
    ufw allow 80/tcp
    wget https://raw.githubusercontent.com/nurdism/neko/master/docker-compose.yaml
    docker-compose up -d
    ```
5. Visit the IP address of the droplet in your browser and login, the default password is `neko`

> *Protip*: Run `nano docker-compose.yaml` to edit the settings, then press *ctrl+x* to exit and save the file.

### Running the container:
```
sudo docker run -p 8080:8080 -e NEKO_PASSWORD='secret' --shm-size=1gb nurdism/neko:latest 
```

*Note:* `--shm-size=1gb` is required, firefox-esr tabs will crash

### Config
```
NEKO_USER=$USERNAME     // User
NEKO_DISPLAY=0          // Display number
NEKO_WIDTH=1280         // Display width
NEKO_HEIGHT=720         // Display height
NEKO_PASSWORD=neko      // Password
NEKO_BIND=0.0.0.0:8080  // Bind
NEKO_KEY=               // (SSL)Key 
NEKO_CERT=              // (SSL)Cert
```

### Development
*Highly* recommend you use a [dev container](https://code.visualstudio.com/docs/remote/containers) for [vscode](https://code.visualstudio.com/), I've included the `.devcontainer` I've used to develop this app. To build neko run:
`cd .docker && ./build`
