# Server (WIP)
Server for n.eko, as of right now this will *only* work on Linux systems, only tested on Debian based distros.

## Configuration
------
```
--debug                                     // (bool) enable debug mode
--logs                                      // (bool) save logs to file
--config ""                                 // (string) configuration file path
--bind "127.0.0.1:8080"                     // (string) address/port/socket to serve neko
--cert ""                                   // (string) path to the SSL cert used to secure the neko server
--key ""                                    // (string) path to the SSL key used to secure the neko server
--static "./www"                            // (string) path to neko client files to serve
--device "auto_null.monitor"                // (string) audio device to capture
--display ":99.0"                           // (string) XDisplay to capture
--audio ""                                  // (string) audio codec parameters to use for streaming (unused)
--video ""                                  // (string) video codec parameters to use for streaming (unused)
--screen "1280x720@30"                      // (string) default screen resolution and framerate
--epr "59000-59100"                         // (string) limits the pool of ephemeral ports that ICE UDP connections can allocate from
--vp8                                       // (bool) use VP8 video codec
--vp9                                       // (bool) use VP9 video codec
--h264                                      // (bool) use H264 video codec
--opus                                      // (bool) use Opus audio codec
--g722                                      // (bool) use G722 audio codec
--pcmu                                      // (bool) use PCMU audio codec
--pcma                                      // (bool) use PCMA audio codec
--nat1to1                                   // ([]string) sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used
--icelite                                   // (bool) configures whether or not the ice agent should be a lite agent
--iceserver "stun:stun.l.google.com:19302"  // ([]string) describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer
--password "neko"                           // (string) password for connecting to stream
--password_admin "admin"                    // (string) admin password for connecting to stream
```

Config can be set via environment variables with the prefix `NEKO_` (I.E. NEKO_BIND="127.0.0.1:8080")

## Requirements
------
Runtime:
```
gstreamer1.0-plugins-base // opus plugin
gstreamer1.0-plugins-good // pulseaudio, ximagesrc & vpx plugins
gstreamer1.0-plugins-bad  // openh264 plugin (build from source)
gstreamer1.0-plugins-ugly // x264 plugin
gstreamer1.0-pulseaudio   // pulseaudio plugin
xclip                     // clipboard sync
libxtst6                  // keyboard and mouse input
```

Development:
```
libxtst-dev
```

## Testing
------
located in `.docker` folder
```
./test firefox  // creates an x server, puleseaudio server add firefox instance
./test chromium // creates an x server, puleseaudio server add chromium instance
```

## Building
------
located in `.docker` folder
```
./build gst     // builds the required gst packages in `.build/gst/`
./build docker  // builds the docker images
./build push    // pushes the images to docker hub
```
