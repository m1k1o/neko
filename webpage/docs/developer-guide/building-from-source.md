---
sidebar_position: 2
---

# Building from Source

This guide walks you through the process of setting up Neko on your local machine or server.

## Prerequisites {#prerequisites}

Before proceeding, ensure that you have the following installed on your system:

- [node.js](https://nodejs.org/) and [npm](https://www.npmjs.com/) (for building the frontend).
- [go](https://golang.org/) (for building the server).
- [gstreamer](https://gstreamer.freedesktop.org/) (for video processing).
  ```shell
  sudo apt-get install libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
      gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
      gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
      gstreamer1.0-pulseaudio;
  ```
- [x.org](https://www.x.org/) (for X11 server).
  ```shell
  sudo apt-get install libx11-dev libxrandr-dev libxtst-dev libxcvt-dev xorg;
  ```
- [pulseaudio](https://www.freedesktop.org/wiki/Software/PulseAudio/) (for audio support).
  ```shell
  sudo apt-get install pulseaudio;
  ```
- other dependencies:
  ```shell
  sudo apt-get install xdotool xclip libgtk-3-0 libgtk-3-dev libopus0 libvpx6;
  ```

## Step 1: Clone the Repository {#step-1}

Start by cloning the Neko Git repository to your machine:

```bash
git clone https://github.com/m1k1o/neko.git
cd neko
```

## Step 2: Build the Frontend {#step-2}

Navigate to the `client` directory and install the dependencies:

```shell
cd client;
npm install;
npm run build;
```

## Step 3: Build the Server {#step-3}

Navigate to the `server` directory and build the server:

```shell
cd server;
go build;
```

## Step 4: Run the Server {#step-4}

Finally, run the server:

```shell
./server/server;
```
