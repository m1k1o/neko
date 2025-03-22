# Frequently Asked Questions

### How to enable debug mode?

To see verbose information from the n.eko server, you can enable debug mode using `NEKO_DEBUG`.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
    - "8080:8080"
    - "52000-52100:52000-52100/udp"
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      # highlight-start
      NEKO_DEBUG: 1
      # highlight-end
```

And then view the logs using `docker logs -f neko`.

To see verbose information from the n.eko client, you need to visit the developer console in your browser. You can do this by pressing `F12` and then navigating to the `Console` tab.

### How to enable support for Chinese/Japanese/Korean input method?

There exists an extension [Google Input Tools](https://chrome.google.com/webstore/detail/mclkkofklkfljcocdinagocijmpgbhab) for Chrome that allows you to use Chinese input method.

### How can I embed the Neko desktop into web page without login prompt coming up for viewers?

You can use the following URL to embed the Neko desktop into a web page without login prompt coming up for viewers:

```
http://<your-neko-server-ip>:8080/?usr=neko&pwd=neko
```

https://stackoverflow.com/questions/15276929/how-to-make-a-video-fullscreen-when-it-is-placed-inside-an-iframe

Your iframe needs an attribute: `allowfullscreen="true" webkitallowfullscreen="true" mozallowfullscreen="true"` or more modern `allow="fullscreen *"`. For the second you can remove the star if your iframe has the same origin or replace it with your iframe origin.

### Can I use neko without docker?

Yes, you can, but it is not recommended. Neko is based on Debian and uses Xorg and Pulseaudio. Just follow the steps in the Dockerfile to install all dependencies.

However, it is recommend to start with existing system that has GUI with desktop manager, is based on Xorg and uses Pulseaudio (e.g. Ubuntu Desktop 24.04). For that matter you only need to install gstreamer dependencies, configure pulseaudio properly and run neko binary (you don't need to build it from scratch, you can copy it from docker image).

### Why does the clipboard button does not show up?

When you are using HTTPS connection and a compatible host browser (currently only Chromium-based browsers) which supports the Clipboard API, the clipboard button will not show up. Instead, you can use the native clipboard functionality of your host browser.

### Why am I unable to install extensions in the Neko browser?

The browser in Neko uses policies to restrict the installation of extensions. You can either add extensions to the policy file or disable the policy.

