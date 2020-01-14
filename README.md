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
1. Head on to [Digital Ocean](https://digitalocean.com/) and create an account
2. Go [here](https://marketplace.digitalocean.com/apps/docker) and click on "Create Docker Droplet"
3. Configure the droplet:
    * **576p** [$15/mo] Not Recommended
    * **720p** [$40/mo] Good Performance
    * **720p** [$80/mo] Recommended
    * **720p+** [$160/mo] Best Performance
4. [Login to the droplet over ssh](https://www.digitalocean.com/docs/droplets/how-to/connect-with-ssh/)
5. Run these commands:
    ```
    ufw allow 80/tcp
    wget https://raw.githubusercontent.com/nurdism/neko/master/docker-compose.yaml
    docker-compose up -d
    ```
5. Visit the IP address of the droplet in your browser and login, the default password is `neko`

> *Protip*: Run `nano docker-compose.yaml` to edit the settings, then press *ctrl+x* to exit and save the file.

Heres the cool part, this will only cost you a little bit (maybe a few cents), *as long as you remember to delete the droplet after you are done!* Droplets are charged per hour, so when you want to share, just create a new droplet and start sharing.

### Running the container:
```
sudo docker run -p 8080:8080 -e NEKO_PASSWORD='secret' --shm-size=2gb nurdism/neko:latest 
```

*Note:* `--shm-size=2gb` is required, firefox-esr tabs will crash (not sure if 2gb is *really* needed)

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
