# dev container

You need to run all dependencies with `deps` command before you start debugging.

Create `.env.development` in repository root. Make sure your local IP is correct.

```sh
NEKO_WEBRTC_NAT1TO1=10.0.0.8
```

# without container

- Make sure `pulseaudio` contains correct configuration.
- Specify `DISPLAY` that is being used by xorg.

```sh
DISPLAY=:0
NEKO_WEBRTC_NAT1TO1=10.0.0.8
NEKO_SERVER_BIND=:3000
```
