<div align="center">
<img src="https://github.com/nurdism/neko/raw/master/.github/logo.png" width="650" height="auto"/>
</div>

<div align="center">
<img src="https://github.com/nurdism/neko/raw/master/.github/demo.gif" width="650" height="auto"/>
</div>

# n.eko
This is a proof of concept project I threw together over the last few days, it's ugly, it's not perfect, but it looks nice. This uses web rtc to stream a desktop inside of a docker container, I made this because [rabb.it](https://en.wikipedia.org/wiki/Rabb.it) went under and my internet can't handle streaming and discord keeps crashing. I just want to watch anime with my friends ლ(ಠ益ಠლ) so I started digging throughout the net and found a few *kinda* clones, but non of them had the virtual browser, then I found [Turtus](https://github.com/Khauri/Turtus) and I was able to figure out the rest. 

This is by no means a fully featured clone of rabbit. The client has no concept of other peers. It has bugs, but for the most part it works. I'm not sure what the future holds for this. If I continue to use it and like it, I'll probably keep pushing updates to it. I'd be happy to accept PRs for any improvements. 

### Why n.eko?
I like cats, I'm a weeb and a nerd, I own the domain [n.eko.moe](https://n.eko.moe/) and I love that logo I came across, had to use it for something /shrug

### I need help setting this up!
Its a docker container, you need to have docker installed, you then need to build the image
```
cd .docker && ./build
```

Then run the container:
```
sudo docker run -p 8080:8080 --shm-size=2gb neko:latest 
```

*Note:* `--shm-size=2gb` is required, firefox-esr tabs will crash (not sure if 2gb is *really* needed)

### Config
```
NEKO_USER=$USERNAME     // User
NEKO_DISPLAY=0          // Display number
NEKO_WIDTH=1280         // Display width
NEKO_HEIGHT=720         // Display width
NEKO_PASSWORD=neko      // Password
NEKO_BIND=0.0.0.0:8080  // Bind
NEKO_KEY=               // Key (SSL)
NEKO_CERT=              // Cert (SSL)
```

### Development
*Highly* recommend you use a dev container for vscode, I've included the .devcontainer I've used to develop this app 