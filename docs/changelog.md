# Changelog

## master branch

### New Features
- Added nvidia support for firefox.
- Added `?lang=<lang>` parameter to the URL, which will set the language of the interface (by @mbattista).

### Misc
- Git commit and tag are now included in the build when creating a docker image.

## [n.eko v2.8.0](https://github.com/m1k1o/neko/releases/tag/v2.8.0)

### New Features
- Added AV1 tag, metadata and pipeline. Unfortunately does not work yet, since the encoding is way too slow (by @mbattista).
- Added `m1k1o/neko:kde` tag as an alternative to `m1k1o/neko:xfce`.
- New VirtualGL version 3.1 was released, adding support for Chromium browsers to use Nvidia GPU acceleration!
- Added `?embed=1` parameter to the URL, which will hide the sidebar and the top bar, so that it can be embedded in other websites.
- Added `?volume=<0-1>` parameter to the URL, which will set the inital volume of the player (by @urbanekpj).
- Touch events are now supported on mobile devices (by @urbanekpj).
- Added NVENC support, hardware h264 encoding for Nvidia GPUs!
- Fixed an issue where `nvh264enc` did not send SPS and PPS NAL units (by @mbattista).

### Bugs
- Fixed TCP mux occasional freeze by adding write buffer to it.
- Fixed stereo problem in chromium-based browsers, where it was only as mono by adding `stereo=1` to opus SDP to clients answer.
- Fixed keysym mapping for unknown keycodes, which was causing some key combinations to not work on some keyboards.
- Fixed a bug where `max_fps=0` would lead to an invalid pipeline.
- Fixed client side webrtc ICE gathering, so that neko can be used without exposed ports, only with STUN and TURN servers.
- Fixed play state synchronization, when autoplay is disabled.

### Misc
- Updated to go 1.19 and Node 18, removed go-events as dependency (by @mbattista).
- Added adaptive framerate which now streams in the framerate you selected from the dropdown.
- Improved chinese and korean characters support.
- Disabled autolock for kde, so that it does not lock the screen when you are not using it.
- Refactored autoplay, so that it will start playing audio, if it's allowed by the browser (by @urbanekpj).
- Renamed pulseaudio sink from `auto_null` to `audio_output`, because it was ignored by KDE.
- Pulseaudio is now configured using environment variables, so that users can mount `/home/neko` without losing audio configuration.

## [n.eko v2.7](https://github.com/m1k1o/neko/releases/tag/v2.7)

### New Features
- Added `m1k1o/neko:vivaldi` tag (thanks @Xeddius).
- Added `m1k1o/neko:opera` tag (thanks @prophetofxenu).
- Added `NEKO_PATH_PREFIX`.
- Added screenshot function `/screenshot.jpg?pwd=<admin>`, works only for unlocked rooms.
- Added emoji support (by @yesBad).
- Added file transfer (by @prophetofxenu).

### Misc
- Server: Split `remote` to `desktop` and `capture`.
- Server: Refactored `xorg` - added `xevent` and clipboard is handled as event (no looped polling anymore).
- Introduced `NEKO_AUDIO_CODEC=` and `NEKO_VIDEO_CODEC=` as a new way of setting codecs.
- Added CORS.
- Opera versions are not hardcoded in Dockerfile anymore but automatically are fetch latest.

## [n.eko v2.6](https://github.com/m1k1o/neko/releases/tag/v2.6)

### Bugs
- Fixed fullscreen incompatibility for Safari [#121](https://github.com/m1k1o/neko/issues/121).
- Fixed bad emoji matching for e.g. `:+1:` and `:100:` with new regex `/^:([^:\s]+):/`.

### New Features
- Added `m1k1o/neko:microsoft-edge` tag.
- Fixed clipboard sync in chromium based browsers.
- Added support for implicit control (using `NEKO_IMPLICITCONTROL=1`). That means, users do not need to request control prior usage.
- Automatically start broadcasting using `NEKO_BROADCAST_URL=rtmp://your-rtmp-endpoint/live` (thanks @konsti).
- Added `m1k1o/neko:remmina` tag (by @lowne).

### Misc
- Automatic WebRTC SDP negotiation using onnegotiationneeded handlers. This allows adding/removing track on demand in a session.
- Added UDP and TCP mux for WebRTC connection. It should handle multiple peers.
- Broadcast status change is sent to all admins now.
- NordVPN replaced with Sponsorblock extension in default configuration #144.
- Removed `vncviewer` image, as its functionality is replaced and extended by remmina.
- Opus uses `useinbandfec=1` from now on, hopefully fixes minor audio loss issues.
- Font Awesome and Sweetalert2 upgraded to newest major version.
- Add chinese characters support.

## [n.eko v2.5](https://github.com/m1k1o/neko/releases/tag/v2.5)

### Bugs
- Fix ungoogled-chromium auto build bug.
- Audio on iOS works now! Apparently only for 15+ though [#62](https://github.com/m1k1o/neko/issues/62).

### New Features
- Lock controls for users, globally.
- Ability to set locks from config `NEKO_LOCKS=control login`.
- Added control protection - users can gain control only if at least one admin is in the room `NEKO_CONTROL_PROTECTION=true`.
- Emotes sending on mouse down holding.
- Include `banned`, `locked`, `server_started_at`, `last_admin_left_at`, `last_user_left_at`, `control_protection` data in stats.

### Misc
- ARM-based images not bound to Raspberry Pi only.
- Repository cleanup, renamed `.m1k1o` to `.docker`.
- Updated docs, now available at https://neko.m1k1o.net.
- Add japanese characters support.
- Sanitize display name and markdown codeblock input to prevent xss.
- Display unmute overlay when joined.
- Sync player play/pause/mute/umpute/volume state with store (beneficial for mobiles when using fullscreen mode).
- Automatic WebRTC SDP negotiation using `onnegotiationneeded` handlers. This allows adding/removing track on demand in a session.

## [n.eko v2.4](https://github.com/m1k1o/neko/releases/tag/v2.4)

### New Features
- Show red dot badge on sidebar toggle if there are new messages, and user can't see them.
- Added `m1k1o/neko:brave` tag.

### Bugs
- Fixed keyboard mapping on macOS, when CMD could not be used for copy & paste.
- Fixed stop signal sent by supervisor to gracefully shut down neko server.

### Misc
- Switched to the latest Firefox version instead of esr.
- Fixed very fast scroll speed on macOS.
- Broadcast pipeline errors are reported to the user.
- On stopping server all websocket connections are going to be gracefully disconnected.

### Other changes
- Upgraded dependencies (server, client),
- Don't kill webrtc on temporary network issues #48.  
- Custom ipfetch #63.
- Build images using github actions #70.
- Refactored RTMP broadcast design #88.
- Based on Debian 11 #91.

## [n.eko v2.3](https://github.com/m1k1o/neko/releases/tag/v2.3)

### New Features
- Added simple language picker.
- Added `?usr=<display-name>` that will prefill username. This allows creating auto-join links.
- Added `?cast=1` that will hide all control and show only video.
- Shake keyboard icon if someone attempted to control when is nobody hosting.
- Support for password protected `NEKO_ICESERVERS` (by @mbattista).
- Added bunch of translations (🇸🇰, 🇪🇸, 🇸🇪, 🇳🇴, 🇫🇷) by various people.
- Added `m1k1o/neko:google-chrome` tag.

### Bugs
- Upgraded and fixed emojis to a new major version.
- Fixed bad `keymap -> keysym` translation to respect active modifiers (#45, with @mbattista).
- Respecting `NEKO_DEBUG` env variable.
- Fullscreen support for iOS devices.
- Added `chrome-sandbox` to fix weird bug when chromium didn't start.

### Misc
- Arguments in broadcast pipeline are optional, not positional and can be repeated `{url} {device} {display}`.
- Chat messages are dense, when repeated, they are joined together.
- While IP address fetching is now proxy ignored.
- Start unmuted on reconnects and auto unmute on any control attempt.

## [n.eko v2.2](https://github.com/m1k1o/neko/releases/tag/v2.2)

### New Features
- Added limited support for some mobile browsers with `playsinline` attribute.
- Added `VIDEO_BITRATE` and `AUDIO_BITRATE` in kbit/s to control stream quality (in collaboration with @mbattista).
- Added `MAX_FPS`, where you can specify max WebRTC frame rate. When set to `0`, frame rate won't be capped and you can enjoy your real `60fps` experience. Originally, it was constant at `25fps`.
- Invite links. You can invite people and they don't need to enter passwords by themselves (and get confused about user accounts that do not exits). You can put your password in URL using `?pwd=<your-password>` and it will be automatically used when logging in.
- Added `/stats?pwd=<admin>` endpoint to get total active connections, host and members.
- Added `m1k1o/neko:vlc` tag, use VLC to watch local files together (by @mbattista).
- Added `m1k1o/neko:xfce` tag, as an non video related showcase (by @mbattista).
- Added ARM-based images, for Raspberry Pi support (by @mbattista).

### Bugs
- Fixed h264 pipelines bugs (by @mbattista).
- Fixed sessions manager thread safety by adding mutexes (caused panic in rare edge cases).
- Now when user gets kicked, he won't join as a ghost user again but will be logged out.
- **iOS compatibility!** Fixed really strange CSS bug, which prevented iOS from loading the video.
- Proper disconnect only once with unsubscribing events. When webrtc fails, user won't be logged in without username again.

### Misc
- Versions bumped: Go 16, Node.js 14 (by @mbattista).
- Remove HTML tags from user name.
- Upgraded `pion/webrtc` to v3 (by @mbattista).
- Added `requestFullscreen` compatibility for older browsers.
- Fixed small lags in video and improved video UX (by @mbattista).
- Added `m1k1o/neko:vncviewer` tag, use `NEKO_VNC_URL` to specify VNC target and use n.eko as a bridge.
- Abiltiy to include neko as a component in another Vue.Js project (by @gbrian).
- Added HEALTHCHECK to Dockerfile.

## [n.eko v2.1](https://github.com/m1k1o/neko/releases/tag/v2.1)

### New Features
- Clipboard button with text area - for browsers, that don't support clipboard syncing or for HTTP.
- Keyboard modifier state synchronization (Num Lock, Caps Lock, Scroll Lock) for each hosting.
- Added chromium ungoogled (with h265 support) an kept up to date by @whalehub.
- Added Picture in Picture button (only for watching screen, controlling not possible).
- Added RTMP broadcast. Enables broadcasting neko screen to local RTMP server, YouTube or Twitch.
- Stereo sound (works properly only in Firefox host).

### Bugs
- Fixed minor gst pipeline bug.
- Locked screen only for users, admins can still join.

### Misc
- Custom docker workflow.
- Based on debian buster instead of stretch.
- Custom avatars without any 3rd party depenency.
- Ignore duplicate notify bars.
- No pointer events for notify bars.
- Disable debug mode by default.

## [n.eko v2.0](https://github.com/nurdism/neko/releases/tag/2.0.0)

## [n.eko v1.1](https://github.com/nurdism/neko/releases/tag/1.1.0)

## [n.eko v1.0](https://github.com/nurdism/neko/releases/tag/1.0.0)
