<div align="center">
<img src="https://github.com/nurdism/neko/raw/master/.github/logo.png" width="650" height="auto"/>
</div>

<div align="center">
<img src="https://i.imgur.com/ZSzbQr7.gif" width="650" height="auto"/>
</div>

# **n**.eko
 This app uses Web RTC to stream a desktop inside of a docker container, I made this because [rabb.it](https://en.wikipedia.org/wiki/Rabb.it) went under and my internet can't handle streaming and discord keeps crashing when my friend attempts to. I just want to watch anime with my friends ლ(ಠ益ಠლ) so I started digging throughout the internet and found a few *kinda* clones, but none of them had the virtual browser, then I found [Turtus](https://github.com/Khauri/Turtus) and I was able to figure out the rest. This is by no means a fully featured clone of rabbit, it hs only *one* room. It's stateless, so no saved user names or passwords. 

### Features
  * Text Chat (with basic markdown support (discord flavor))
  * Admin users (Kick, Ban & Force Give/Release Controls)
  * Clipboard synchronization
  * Emote overlay
  * Ignore user (chat and emotes)
  * Settings are saved to local storage

### Why **n**.eko?
I like cats 🐱 (`Neko` is the Japanese word for cat), I'm a weeb/nerd

***But why the cat butt?*** Because cats are *assholes*, but you love them anyways.

### Super easy mode setup
1. Deploy a server or VPS

    *Recommended Specs:*
    
    | Resolution | Cores | Ram   | Recommendation   |
    |------------|-------|-------|------------------|
    | **576p**   | 2     | 2gb   | Not Recommended  |
    | **720p**   | 4     | 4gb   | Good Performance |
    | **720p**   | 6     | 4-6gb | Recommended      |
    | **720p+**  | 8     | 8gb+  | Best Performance |
  
    ***Why are the specs so high?*** : If you think about it, you have to run a full desktop, a browser (a resource hog on its own) *and* encode/transmit the desktop, there's a lot going on and so it demands some power.

2. [Login via SSH](https://www.digitalocean.com/docs/droplets/how-to/connect-with-ssh/)

3. Install Docker
    ```
    curl -sSL https://get.docker.com/ | CHANNEL=stable bash
    ```
4. Run these commands:
    ```
    sudo ufw allow 80/tcp // if you have ufw installed/enabled
    sudo ufw allow 59000:59100/udp
    wget https://raw.githubusercontent.com/nurdism/neko/master/docker-compose.yaml
    sudo docker-compose up -d
    ```
5. Visit the IP address of the droplet in your browser and login, the default password is `neko`

> 💡 **Protip**: Run `nano docker-compose.yaml` to edit the settings, then press *ctrl+x* to exit and save the file.

### Running the container:
```
sudo docker run -p 8080:8080 -p 59000-59100:59000-59100/udp -e NEKO_PASSWORD='secret' -e NEKO_ADMIN='secret' --shm-size=1gb nurdism/neko:latest 
```

*Note:* `--shm-size=1gb` is required, firefox tabs will crash, not sure what it does to be honest 😅

### Docker Basic Configuration
```
SCREEN_WIDTH=1280       // Display width
SCREEN_HEIGHT=720       // Display height
SCREEN_DEPTH=24         // Display bit depth
DISPLAY=:99.0           // Display number

NEKO_PASSWORD=neko      // Password
NEKO_ADMIN=neko         // Admin Password
NEKO_BIND=0.0.0.0:8080  // Bind
NEKO_KEY=               // (SSL)Key, needed for clipboard sync
NEKO_CERT=              // (SSL)Cert, needed for clipboard sync
```
for full documentation on configuring the server [go here](./server/README.md)

### Development
*Highly* recommend you use a [dev container](https://code.visualstudio.com/docs/remote/containers) for [vscode](https://code.visualstudio.com/), I've included the `.devcontainer` I've used to develop this app. To build **n**.eko docker container run:
```
cd .docker && ./build
```
the `.docker` folder also contains `./test` bash script which will launch a desktop with a browser for testing out any changes with the server.

To run the client with hot loading (for development of new client features)
```
cd ./client && npm run serve
```

### Non Goals
* Turning n.eko into a service that serves multiple rooms and browsers/desktops.
* Supporting multiple platforms
* Voice chat, use [Discord](https://discordapp.com/)