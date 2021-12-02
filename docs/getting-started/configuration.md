# Configuration

## Environment variables

```
NEKO_SCREEN:
  - Resolution after startup. Only Admins can change this later.
  - e.g. '1920x1080@30'
NEKO_PASSWORD:
  - Password for the user login
  - e.g. 'user_password'
NEKO_PASSWORD_ADMIN
  - Password for the admin login
  - e.g. 'admin_password'
NEKO_EPR:
  - For WebRTC needed range of ports
  - e.g. 52000-52100
NEKO_VP8:
  - If vp8 should be used as video encoder for the stream (default encoder)
  - e.g. 'true'
NEKO_VP9:
  - If vp9 should be used as video encoder for the stream (Parameter not optimized yet)
  - e.g. 'false'
NEKO_H264:
  - If h264 should be used as video encoder for the stream (second best option)
  - e.g. 'false'
NEKO_VIDEO_BITRATE:
  - Bitrate of the video stream in kb/s
  - e.g. 3500
NEKO_VIDEO:
  - Makes it possible to create custom gstreamer pipelines. With this you could find the best quality for your CPU
  - Installed are gstreamer1.0-plugins-base /  gstreamer1.0-plugins-good /  gstreamer1.0-plugins-bad /  gstreamer1.0-plugins-ugly
  - e.g. ' ximagesrc display-name=%s show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue ! video/x-raw,format=NV12 ! x264enc threads=4 bitrate=3500 key-int-max=60 vbv-buf-capacity=4000 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream '
NEKO_MAX_FPS:
  - The resulting stream frames per seconds should be capped (0 for uncapped)
  - e.g. 0
NEKO_OPUS:
  - If opus should be used as audio encoder for the stream (default encoder)
  - e.g. 'true'
NEKO_G722:
  - If g722 should be used as audio encoder for the stream
  - e.g. 'false'
NEKO_PCMU:
  - If pcmu should be used as audio encoder for the stream
  - e.g. 'false'
NEKO_PCMA:
  - If pcma should be used as audio encoder for the stream
  - e.g. 'false'
NEKO_AUDIO_BITRATE:
  - Bitrate of the audio stream in kb/s
  - e.g. 196
NEKO_CERT:
  - Path to the SSL-Certificate
  - e.g. '/certs/cert.pem'
NEKO_KEY:
  - Path to the SSL-Certificate private key
  - e.g. '/certs/key.pem'
NEKO_ICELITE:
  - Use the ice lite protocol
  - e.g. false
NEKO_ICESERVER:
  - Describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer (simple usage for server without authentication)
  - e.g. 'stun:stun.l.google.com:19302'
NEKO_ICESERVERS:
  - Describes multiple STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer
  - e.g. '[{"urls": ["turn:turn.example.com:19302", "stun:stun.example.com:19302"], "username": "name", "credential": "password"}, {"urls": ["stun:stun.example2.com:19302"]}]'
  - [More information](https://developer.mozilla.org/en-US/docs/Web/API/RTCIceServer)
NEKO_LOCKS:
  - Resources, that will be locked when starting, separated by whitespace.
  - Currently supported: control login
NEKO_CONTROL_PROTECTION:
  - Control protection means, users can gain control only if at least one admin is in the room.
  - e.g. false
```

## Agruments

You can execute `neko serve --help` to see available arguments.

```
Usage:
  neko serve [flags]

Flags:
      --audio string                audio codec parameters to use for streaming
      --audio_bitrate int           audio bitrate in kbit/s (default 128)
      --bind string                 address/port/socket to serve neko (default "127.0.0.1:8080")
      --broadcast_pipeline string   custom gst pipeline used for broadcasting, strings {url} {device} {display} will be replaced
      --cert string                 path to the SSL cert used to secure the neko server
      --device string               audio device to capture (default "auto_null.monitor")
      --display string              XDisplay to capture (default ":99.0")
      --epr string                  limits the pool of ephemeral ports that ICE UDP connections can allocate from (default "59000-59100")
      --g722                        use G722 audio codec
      --h264                        use H264 video codec
  -h, --help                        help for serve
      --icelite                     configures whether or not the ice agent should be a lite agent
      --iceserver strings           describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer (default [stun:stun.l.google.com:19302])
      --iceservers string           describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer
      --ipfetch string              automatically fetch IP address from given URL when nat1to1 is not present (default "http://checkip.amazonaws.com")
      --key string                  path to the SSL key used to secure the neko server
      --max_fps int                 maximum fps delivered via WebRTC, 0 is for no maximum (default 25)
      --nat1to1 strings             sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used
      --opus                        use Opus audio codec
      --password string             password for connecting to stream (default "neko")
      --password_admin string       admin password for connecting to stream (default "admin")
      --pcma                        use PCMA audio codec
      --pcmu                        use PCMU audio codec
      --proxy                       enable reverse proxy mode
      --screen string               default screen resolution and framerate (default "1280x720@30")
      --static string               path to neko client files to serve (default "./www")
      --video string                video codec parameters to use for streaming
      --video_bitrate int           video bitrate in kbit/s (default 3072)
      --vp8                         use VP8 video codec
      --vp9                         use VP9 video codec

Global Flags:
      --config string   configuration file path
  -d, --debug           enable debug mode
  -l, --logs            save logs to file
```

## Config file

You can mount YAML config file to docker container on this path `/etc/neko/neko.yaml` and store your configuration there.
