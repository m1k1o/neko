# Release Notes

## master {#master}

### New Features {#master-feats}
- Scroll to chat on mobile ([#496](https://github.com/m1k1o/neko/pull/496))
- Added mobile keyboard icon to open the keyboard on mobile devices ([#497](https://github.com/m1k1o/neko/pull/497))

### Fixes {#master-fixes}
- Fixed various bugs related to the legacy client and migration.
- Fixed long standing issue [#279](https://github.com/m1k1o/neko/pull/279) where Google Chrome GPU acceleration did not work with Nvidia GPUs, thanks to [@TobyColeman](https://github.com/TobyColeman), [@alexbakerdev](https://github.com/alexbakerdev) and [@samstefan](https://github.com/samstefan) from [@wearewildcards](https://github.com/wearewildcards).

### Misc {#master-misc}
- Added an https condition to the healthcheck ([#503](https://github.com/m1k1o/neko/pull/503), by @Garrulousbrevity).

## [n.eko v3.0.0](https://github.com/m1k1o/neko/releases/tag/v3.0.0) {#v3.0.0}

### Repository Changes {#v3.0.0-repo}
- The default registry is now `ghcr.io/m1k1o/neko` instead of `docker.io/m1k1o/neko`.
- Multiarch builds for `linux/amd64`, `linux/arm64`, and `linux/arm/v7` are now available instead of `arm-`based images.
- App folders from `.docker/` have been moved to `apps/`.
- Dev scripts from `.docker/` are now available in `client/dev/` and `server/dev/`.
- The `docs/` folder is now available at `webpage/docs/` with a completely new structure.
- The base `Dockerfile` has been split into `client/Dockerfile`, `server/Dockerfile`, and `runtime/Dockerfile`.
- The build process has been moved from `.docker/build` to `build/`, supporting more options. See `--help` for more information.
- Brave, Vivaldi, Remmina, and KDE are now also available as ARM-based images.
- Waterfox is now available as a new browser.

### Server Changes {#v3.0.0-server}
- A REST API with OpenAPI 3.0 documentation is now available.
- Prometheus metrics are now available.
- The server name is now `github.com/m1k1o/neko/server` and can be used as a library.
- Reusable components and types are available in the `server/pkg/` folder, such as `gst`, `xevent`, and `xorg`.
- A new authentication system with support for multiple authentication methods has been added.
- A new user management system with support for granular feature access control has been implemented.
- The structure for configuration options has been updated, with options now separated into modules.
- Compatibility with V2 configuration options is still available but deprecated.
- **Capture**: Added a screencast feature as a fallback for WebRTC.
- **Capture**: Added experimental webcam and microphone passthrough support.
- **Capture**: Added video simulcast support and a stream selector.
- **Capture**: When joining a pipeline, a keyframe is requested on demand and sent to the client.
- **Desktop**: Clipboard now uses `xclip` instead of `libclipboard`, allowing multiple targets.
- **Desktop**: Added drag-and-drop file upload support.
- **Desktop**: Added a file chooser dialog to prompt users for file uploads (experimental).
- **Desktop**: Added an unminimize feature to ensure that the window is not minimized when the user is trying to control it.
- **Desktop**: Created a custom input X11 driver to support touchscreen devices.
- **Desktop**: Added support for `xrandr` to set the virtual monitor resolution to any resolution, not just predefined ones.
- **Desktop**: Added a function to send events when the cursor changes, along with the cursor image.
- **HTTP**: Added batch mode to allow multiple requests in a single connection.
- **HTTP**: Added `pprof` support to enable server profiling.
- **HTTP**: Created a legacy driver to support the current neko client.
- **HTTP**: Refactored HTTP logging.
- **Plugins**: Added support for Go plugins to enable custom features to be added to the server.
- **Plugins**: Chat has been implemented as a plugin that can be disabled globally or per user (mute feature).
- **Plugins**: File upload has been implemented as a plugin that can be disabled globally or per user.
- **Session**: Added support to save session tokens as cookies to allow persistent login.
- **Session**: Added the ability to serialize and deserialize sessions to a file to survive restarts.
- **Session**: Added support for dynamic permissions with granular feature access control.
- **WebRTC**: Forwarded desktop cursor changes to the client.
- **WebRTC**: Forwarded cursor position to other clients that have enabled the inactive cursors option.
- **WebRTC**: Switched from LittleEndian to BigEndian for the video stream to improve browser compatibility.
- **WebRTC**: Created a legacy driver to support the current neko client.
- **WebRTC**: Added WebRTC ping to check if the connection is still alive and to determine latency.
- **WebRTC**: Added the ability to switch video pipelines on the fly.
- **WebRTC**: Implemented bandwidth estimation and adaptive quality (experimental).
- **WebSocket**: Added support for controlling the desktop using WebSockets as a fallback for WebRTC.
- **WebSocket**: Added support for sending unicast and broadcast messages to all clients.

Please note that in this version, only the server has been updated. The client is still in the old version; therefore, new features may not yet be available in the client.

## [n.eko v2.9.0](https://github.com/m1k1o/neko/releases/tag/v2.9.0) {#v2.9.0}

### New Features {#v2.9.0-feats}
- Added nvidia support for firefox.
- Added `?lang=<lang>` parameter to the URL, which will set the language of the interface (by @mbattista).
- Added `?show_side=1` and `?mute_chat=1` parameter to the URL, for chat mute and show side (by @mbattista).
- Added `NEKO_BROADCAST_AUTOSTART` to automatically start or do not start broadcasting when the room is created. By default, it is set to `true` because it was the previous behavior.
- Added new translations (🇹🇼,🇯🇵) by various people.

### Bugs {#v2.9.0-bugs}
- Fix incorrect version sorting for chromium, microsoft-edge, opera and ungoogledchromium.
- Fix buffer overflow in Gstreamer log function [#382](https://github.com/m1k1o/neko/pull/382) (by @tt2468).

### Misc {#v2.9.0-misc}
- Added RTMP broadcast support to nvidia docker image [#274](https://github.com/m1k1o/neko/issues/274).
- Ensured that paths are writable by neko user [#277](https://github.com/m1k1o/neko/issues/277).
- Git commit and tag are now included in the build when creating a docker image.
- Remove any temporary files associated with a Form after file upload, that would be otherwise never removed.
- Add check for volume parameter in URL before setting volume (by @FapFapDragon).
- Add glib main loop to capture manager [#383](https://github.com/m1k1o/neko/pull/383) (by @tt2468).
- Sync clipboard only if in focus.

## [n.eko v2.8.0](https://github.com/m1k1o/neko/releases/tag/v2.8.0) {#v2.8.0}

### New Features {#v2.8.0-feats}
- Added AV1 tag, metadata and pipeline. Unfortunately does not work yet, since the encoding is way too slow (by @mbattista).
- Added `m1k1o/neko:kde` tag as an alternative to `m1k1o/neko:xfce`.
- New VirtualGL version 3.1 was released, adding support for Chromium browsers to use Nvidia GPU acceleration!
- Added `?embed=1` parameter to the URL, which will hide the sidebar and the top bar, so that it can be embedded in other websites.
- Added `?volume=<0-1>` parameter to the URL, which will set the inital volume of the player (by @urbanekpj).
- Touch events are now supported on mobile devices (by @urbanekpj).
- Added NVENC support, hardware h264 encoding for Nvidia GPUs!
- Fixed an issue where `nvh264enc` did not send SPS and PPS NAL units (by @mbattista).

### Bugs {#v2.8.0-bugs}
- Fixed TCP mux occasional freeze by adding write buffer to it.
- Fixed stereo problem in chromium-based browsers, where it was only as mono by adding `stereo=1` to opus SDP to clients answer.
- Fixed keysym mapping for unknown keycodes, which was causing some key combinations to not work on some keyboards.
- Fixed a bug where `max_fps=0` would lead to an invalid pipeline.
- Fixed client side webrtc ICE gathering, so that neko can be used without exposed ports, only with STUN and TURN servers.
- Fixed play state synchronization, when autoplay is disabled.

### Misc {#v2.8.0-misc}
- Updated to go 1.19 and Node 18, removed go-events as dependency (by @mbattista).
- Added adaptive framerate which now streams in the framerate you selected from the dropdown.
- Improved chinese and korean characters support.
- Disabled autolock for kde, so that it does not lock the screen when you are not using it.
- Refactored autoplay, so that it will start playing audio, if it's allowed by the browser (by @urbanekpj).
- Renamed pulseaudio sink from `auto_null` to `audio_output`, because it was ignored by KDE.
- Pulseaudio is now configured using environment variables, so that users can mount `/home/neko` without losing audio configuration.

## [n.eko v2.7](https://github.com/m1k1o/neko/releases/tag/v2.7) {#v2.7}

### New Features {#v2.7-feats}
- Added `m1k1o/neko:vivaldi` tag (thanks @Xeddius).
- Added `m1k1o/neko:opera` tag (thanks @prophetofxenu).
- Added `NEKO_PATH_PREFIX`.
- Added screenshot function `/screenshot.jpg?pwd=<admin>`, works only for unlocked rooms.
- Added emoji support (by @yesBad).
- Added file transfer (by @prophetofxenu).

### Misc {#v2.7-misc}
- Server: Split `remote` to `desktop` and `capture`.
- Server: Refactored `xorg` - added `xevent` and clipboard is handled as event (no looped polling anymore).
- Introduced `NEKO_AUDIO_CODEC=` and `NEKO_VIDEO_CODEC=` as a new way of setting codecs.
- Added CORS.
- Opera versions are not hardcoded in Dockerfile anymore but automatically are fetch latest.

## [n.eko v2.6](https://github.com/m1k1o/neko/releases/tag/v2.6) {#v2.6}

### Bugs {#v2.6-bugs}
- Fixed fullscreen incompatibility for Safari [#121](https://github.com/m1k1o/neko/issues/121).
- Fixed bad emoji matching for e.g. `:+1:` and `:100:` with new regex `/^:([^:\s]+):/`.

### New Features {#v2.6-feats}
- Added `m1k1o/neko:microsoft-edge` tag.
- Fixed clipboard sync in chromium based browsers.
- Added support for implicit control (using `NEKO_IMPLICITCONTROL=1`). That means, users do not need to request control prior usage.
- Automatically start broadcasting using `NEKO_BROADCAST_URL=rtmp://your-rtmp-endpoint/live` (thanks @konsti).
- Added `m1k1o/neko:remmina` tag (by @lowne).

### Misc {#v2.6-misc}
- Automatic WebRTC SDP negotiation using onnegotiationneeded handlers. This allows adding/removing track on demand in a session.
- Added UDP and TCP mux for WebRTC connection. It should handle multiple peers.
- Broadcast status change is sent to all admins now.
- NordVPN replaced with Sponsorblock extension in default configuration #144.
- Removed `vncviewer` image, as its functionality is replaced and extended by remmina.
- Opus uses `useinbandfec=1` from now on, hopefully fixes minor audio loss issues.
- Font Awesome and Sweetalert2 upgraded to newest major version.
- Add chinese characters support.

## [n.eko v2.5](https://github.com/m1k1o/neko/releases/tag/v2.5) {#v2.5}

### Bugs {#v2.5-bugs}
- Fix ungoogled-chromium auto build bug.
- Audio on iOS works now! Apparently only for 15+ though [#62](https://github.com/m1k1o/neko/issues/62).

### New Features {#v2.5-feats}
- Lock controls for users, globally.
- Ability to set locks from config `NEKO_LOCKS=control login`.
- Added control protection - users can gain control only if at least one admin is in the room `NEKO_CONTROL_PROTECTION=true`.
- Emotes sending on mouse down holding.
- Include `banned`, `locked`, `server_started_at`, `last_admin_left_at`, `last_user_left_at`, `control_protection` data in stats.

### Misc {#v2.5-misc}
- ARM-based images not bound to Raspberry Pi only.
- Repository cleanup, renamed `.m1k1o` to `.docker`.
- Updated docs, now available at https://neko.m1k1o.net.
- Add japanese characters support.
- Sanitize display name and markdown codeblock input to prevent xss.
- Display unmute overlay when joined.
- Sync player play/pause/mute/umpute/volume state with store (beneficial for mobiles when using fullscreen mode).
- Automatic WebRTC SDP negotiation using `onnegotiationneeded` handlers. This allows adding/removing track on demand in a session.

## [n.eko v2.4](https://github.com/m1k1o/neko/releases/tag/v2.4) {#v2.4}

### New Features {#v2.4-feats}
- Show red dot badge on sidebar toggle if there are new messages, and user can't see them.
- Added `m1k1o/neko:brave` tag.

### Bugs {#v2.4-bugs}
- Fixed keyboard mapping on macOS, when CMD could not be used for copy & paste.
- Fixed stop signal sent by supervisor to gracefully shut down neko server.

### Misc {#v2.4-misc}
- Switched to the latest Firefox version instead of esr.
- Fixed very fast scroll speed on macOS.
- Broadcast pipeline errors are reported to the user.
- On stopping server all websocket connections are going to be gracefully disconnected.

### Other changes {#v2.4-other}
- Upgraded dependencies (server, client),
- Don't kill webrtc on temporary network issues #48.  
- Custom ipfetch #63.
- Build images using github actions #70.
- Refactored RTMP broadcast design #88.
- Based on Debian 11 #91.

## [n.eko v2.3](https://github.com/m1k1o/neko/releases/tag/v2.3) {#v2.3}

### New Features {#v2.3-feats}
- Added simple language picker.
- Added `?usr=<display-name>` that will prefill username. This allows creating auto-join links.
- Added `?cast=1` that will hide all control and show only video.
- Shake keyboard icon if someone attempted to control when is nobody hosting.
- Support for password protected `NEKO_ICESERVERS` (by @mbattista).
- Added bunch of translations (🇸🇰, 🇪🇸, 🇸🇪, 🇳🇴, 🇫🇷) by various people.
- Added `m1k1o/neko:google-chrome` tag.

### Bugs {#v2.3-bugs}
- Upgraded and fixed emojis to a new major version.
- Fixed bad `keymap -> keysym` translation to respect active modifiers (#45, with @mbattista).
- Respecting `NEKO_DEBUG` env variable.
- Fullscreen support for iOS devices.
- Added `chrome-sandbox` to fix weird bug when chromium didn't start.

### Misc {#v2.3-misc}
- Arguments in broadcast pipeline are optional, not positional and can be repeated `{url} {device} {display}`.
- Chat messages are dense, when repeated, they are joined together.
- While IP address fetching is now proxy ignored.
- Start unmuted on reconnects and auto unmute on any control attempt.

## [n.eko v2.2](https://github.com/m1k1o/neko/releases/tag/v2.2) {#v2.2}

### New Features {#v2.2-feats}
- Added limited support for some mobile browsers with `playsinline` attribute.
- Added `VIDEO_BITRATE` and `AUDIO_BITRATE` in kbit/s to control stream quality (in collaboration with @mbattista).
- Added `MAX_FPS`, where you can specify max WebRTC frame rate. When set to `0`, frame rate won't be capped and you can enjoy your real `60fps` experience. Originally, it was constant at `25fps`.
- Invite links. You can invite people and they don't need to enter passwords by themselves (and get confused about user accounts that do not exits). You can put your password in URL using `?pwd=<your-password>` and it will be automatically used when logging in.
- Added `/stats?pwd=<admin>` endpoint to get total active connections, host and members.
- Added `m1k1o/neko:vlc` tag, use VLC to watch local files together (by @mbattista).
- Added `m1k1o/neko:xfce` tag, as an non video related showcase (by @mbattista).
- Added ARM-based images, for Raspberry Pi support (by @mbattista).

### Bugs {#v2.2-bugs}
- Fixed h264 pipelines bugs (by @mbattista).
- Fixed sessions manager thread safety by adding mutexes (caused panic in rare edge cases).
- Now when user gets kicked, he won't join as a ghost user again but will be logged out.
- **iOS compatibility!** Fixed really strange CSS bug, which prevented iOS from loading the video.
- Proper disconnect only once with unsubscribing events. When webrtc fails, user won't be logged in without username again.

### Misc {#v2.2-misc}
- Versions bumped: Go 16, Node.js 14 (by @mbattista).
- Remove HTML tags from user name.
- Upgraded `pion/webrtc` to v3 (by @mbattista).
- Added `requestFullscreen` compatibility for older browsers.
- Fixed small lags in video and improved video UX (by @mbattista).
- Added `m1k1o/neko:vncviewer` tag, use `NEKO_VNC_URL` to specify VNC target and use n.eko as a bridge.
- Abiltiy to include neko as a component in another Vue.Js project (by @gbrian).
- Added HEALTHCHECK to Dockerfile.

## [n.eko v2.1](https://github.com/m1k1o/neko/releases/tag/v2.1) {#v2.1}

### New Features {#v2.1-feats}
- Clipboard button with text area - for browsers, that don't support clipboard syncing or for HTTP.
- Keyboard modifier state synchronization (Num Lock, Caps Lock, Scroll Lock) for each hosting.
- Added chromium ungoogled (with h265 support) an kept up to date by @whalehub.
- Added Picture in Picture button (only for watching screen, controlling not possible).
- Added RTMP broadcast. Enables broadcasting neko screen to local RTMP server, YouTube or Twitch.
- Stereo sound (works properly only in Firefox host).

### Bugs {#v2.1-bugs}
- Fixed minor gst pipeline bug.
- Locked screen only for users, admins can still join.

### Misc {#v2.1-misc}
- Custom docker workflow.
- Based on debian buster instead of stretch.
- Custom avatars without any 3rd party depenency.
- Ignore duplicate notify bars.
- No pointer events for notify bars.
- Disable debug mode by default.

## [n.eko v2.0](https://github.com/nurdism/neko/releases/tag/2.0.0) {#v2.0.0}

## [n.eko v1.1](https://github.com/nurdism/neko/releases/tag/1.1.0) {#v1.1.0}

## [n.eko v1.0](https://github.com/nurdism/neko/releases/tag/1.0.0) {#v1.0.0}
