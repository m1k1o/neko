# Release Notes

## master {#master}

No unreleased changes yet.

## [n.eko v3.1.5](https://github.com/m1k1o/neko/releases/tag/v3.1.5) {#v3.1.5}

### New Features {#v3.1.5-feats}
- Added H.265 (HEVC) video encoding support ([#648](https://github.com/m1k1o/neko/pull/648)).
- Added keyboard input via `XTestFakeDeviceKeyEvent` for improved XInput device compatibility ([#642](https://github.com/m1k1o/neko/pull/642)).
- Updated NVIDIA video encoding to use `nvautogpuh264enc` for NVIDIA driver 590+ ([#667](https://github.com/m1k1o/neko/pull/667)).
- Clipboard is now synced automatically when the browser window gains focus ([#662](https://github.com/m1k1o/neko/pull/662)).
- Display name input is now auto-focused when entering a room ([#654](https://github.com/m1k1o/neko/pull/654)).

### Fixes {#v3.1.5-fixes}
- Fixed incorrect `LegacyIsHost` boolean logic ([#675](https://github.com/m1k1o/neko/pull/675)).
- Fixed CORS handling: wildcard usage now emits a warning and CORS is disabled by default ([#674](https://github.com/m1k1o/neko/pull/674)).
- Fixed ungoogled-chromium Dockerfile to correctly fetch and extract the binary ([#676](https://github.com/m1k1o/neko/pull/676)).
- Reduced audio jitter buffer latency ([#673](https://github.com/m1k1o/neko/pull/673)).
- Fixed video pipelines being incorrectly overridden when using legacy settings ([#666](https://github.com/m1k1o/neko/pull/666)).
- Disabled `CommandLineFlagSecurityWarnings` in browser policies for Chromium-based browsers ([#668](https://github.com/m1k1o/neko/pull/668)).

### Misc {#v3.1.5-misc}
- Installed Widevine DRM support for ARM64 ([#660](https://github.com/m1k1o/neko/pull/660)).
- Upgraded Go dependencies ([#644](https://github.com/m1k1o/neko/pull/644)).

## [n.eko v3.1.4](https://github.com/m1k1o/neko/releases/tag/v3.1.4) {#v3.1.4}

### Fixes {#v3.1.4-fixes}
- Fixed regressions in GHCR workflow updates ([#639](https://github.com/m1k1o/neko/issues/639), [#638](https://github.com/m1k1o/neko/issues/638)).

### Misc {#v3.1.4-misc}
- Added latest-version check and automatic image update logic for GHCR workflows ([#637](https://github.com/m1k1o/neko/pull/637)).

## [n.eko v3.1.3](https://github.com/m1k1o/neko/releases/tag/v3.1.3) {#v3.1.3}

### Misc {#v3.1.3-misc}
- Upgraded Go to version 1.25.
- Updated Go module dependencies.
- Standardized internal constant names and improved goroutine lifecycle handling.

## [n.eko v3.1.2](https://github.com/m1k1o/neko/releases/tag/v3.1.2) {#v3.1.2}

### Fixes {#v3.1.2-fixes}
- Fixed profile API endpoint to only update the `name` field, preventing unintended changes to other profile attributes.

## [n.eko v3.1.1](https://github.com/m1k1o/neko/releases/tag/v3.1.1) {#v3.1.1}

### Misc {#v3.1.1-misc}
- Removed support for the `linux/arm/v7` platform. Supported platforms are now `linux/amd64` and `linux/arm64`.

## [n.eko v3.1.0](https://github.com/m1k1o/neko/releases/tag/v3.1.0) {#v3.1.0}

### New Features {#v3.1.0-feats}
- Upgraded base OS from Debian Bookworm (12) to Debian Trixie (13), bringing updated system packages and improved hardware support ([#581](https://github.com/m1k1o/neko/pull/581)).
- Added microphone passthrough button to the controls toolbar ([#620](https://github.com/m1k1o/neko/pull/620)).

### Fixes {#v3.1.0-fixes}
- Completed and standardized all locale translations ([#549](https://github.com/m1k1o/neko/pull/549)).

### Misc {#v3.1.0-misc}
- Added `vlc-plugin-skins2` to the VLC image ([#623](https://github.com/m1k1o/neko/pull/623)).

## [n.eko v3.0.11](https://github.com/m1k1o/neko/releases/tag/v3.0.11) {#v3.0.11}

Maintenance release for the v3.0.x series, backporting the new features from v3.1.0 while keeping Debian Bookworm (12) as the base OS.

### New Features {#v3.0.11-feats}
- Added microphone passthrough button to the controls toolbar ([#620](https://github.com/m1k1o/neko/pull/620)).

### Fixes {#v3.0.11-fixes}
- Fixed profile API endpoint to only update the `name` field, preventing unintended changes to other profile attributes.
- Completed and standardized all locale translations ([#549](https://github.com/m1k1o/neko/pull/549)).

### Misc {#v3.0.11-misc}
- Added `vlc-plugin-skins2` to the VLC image ([#623](https://github.com/m1k1o/neko/pull/623)).

## [n.eko v3.0.10](https://github.com/m1k1o/neko/releases/tag/v3.0.10) {#v3.0.10}

### Fixes {#v3.0.10-fixes}
- Fixed legacy API calls to correctly include the configured path prefix ([#615](https://github.com/m1k1o/neko/issues/615), [#618](https://github.com/m1k1o/neko/pull/618)).
- Fixed `--no-cache` flag in the build script.

## [n.eko v3.0.9](https://github.com/m1k1o/neko/releases/tag/v3.0.9) {#v3.0.9}

### Fixes {#v3.0.9-fixes}
- Fixed Opera browser Docker image build.
- Changed WebRTC heartbeat interval from 120 s to 10 s for more responsive disconnection detection ([#585](https://github.com/m1k1o/neko/issues/585)).
- Updated all Debian Bookworm-based images to the latest packages ([#580](https://github.com/m1k1o/neko/pull/580)).

### Misc {#v3.0.9-misc}
- Added Polish (`pl`) translation ([#582](https://github.com/m1k1o/neko/pull/582)).
- Removed additional Google telemetry from ungoogled-chromium.

## [n.eko v3.0.8](https://github.com/m1k1o/neko/releases/tag/v3.0.8) {#v3.0.8}

### New Features {#v3.0.8-feats}
- Added support for entering text indirectly through the virtual keyboard ([#577](https://github.com/m1k1o/neko/pull/577)).
- Hostname is now available as a broadcast template variable ([#576](https://github.com/m1k1o/neko/pull/576)).

### Fixes {#v3.0.8-fixes}
- Fixed WidevineCDM download for DRM-protected content ([#578](https://github.com/m1k1o/neko/pull/578)).
- Fixed unreliable CapsLock behavior under macOS Chrome (upstream fix from GUACAMOLE-1823).

## [n.eko v3.0.7](https://github.com/m1k1o/neko/releases/tag/v3.0.7) {#v3.0.7}

### New Features {#v3.0.7-feats}
- Implicit hosting: when a user interacts with the screen, control is automatically requested first ([#499](https://github.com/m1k1o/neko/issues/499), [#540](https://github.com/m1k1o/neko/pull/540)).

### Misc {#v3.0.7-misc}
- Added Indonesian (`id`) translation ([#552](https://github.com/m1k1o/neko/pull/552)).
- Upgraded Go to version 1.24 and updated all dependencies ([#564](https://github.com/m1k1o/neko/pull/564)).

## [n.eko v3.0.6](https://github.com/m1k1o/neko/releases/tag/v3.0.6) {#v3.0.6}

### New Features {#v3.0.6-feats}
- Added clipboard command replacement support on the desktop ([#539](https://github.com/m1k1o/neko/pull/539)).

### Fixes {#v3.0.6-fixes}
- Fixed Vivaldi browser installation in the Docker image.

### Misc {#v3.0.6-misc}
- Updated VirtualGL to a more recent version ([#538](https://github.com/m1k1o/neko/pull/538)).

## [n.eko v3.0.5](https://github.com/m1k1o/neko/releases/tag/v3.0.5) {#v3.0.5}

### Fixes {#v3.0.5-fixes}
- Fixed mobile keyboard behavior ([#522](https://github.com/m1k1o/neko/pull/522), [#523](https://github.com/m1k1o/neko/issues/523)).
- Fixed clipboard to use the `UTF8_STRING` target for better compatibility ([#517](https://github.com/m1k1o/neko/issues/517)).
- Fixed WebRTC pong message forwarding ([#510](https://github.com/m1k1o/neko/issues/510)).
- Fixed build script for Apple Silicon (macOS) ([#520](https://github.com/m1k1o/neko/pull/520)).
- Fixed Docker volume mount error in the build script ([#519](https://github.com/m1k1o/neko/pull/519)).
- Commented out Firefox `xpinstall` preferences that were causing extension issues ([#512](https://github.com/m1k1o/neko/issues/512)).

### Misc {#v3.0.5-misc}
- Added `SECURITY.md` with vulnerability reporting instructions.
- Temporarily disabled Waterfox build while the upstream download link is unavailable.

## [n.eko v3.0.4](https://github.com/m1k1o/neko/releases/tag/v3.0.4) {#v3.0.4}

### Fixes {#v3.0.4-fixes}
- Fixed HTTPS + legacy mode: a local HTTP server is now also started to handle both connection types ([#507](https://github.com/m1k1o/neko/pull/507)).
- Disabled HTTP proxy for local requests to prevent unintended request forwarding ([#509](https://github.com/m1k1o/neko/issues/509)).
- Updated Waterfox user agent to work around Cloudflare bot protection blocking downloads.

### Misc {#v3.0.4-misc}
- Added HTTPS condition to the Docker healthcheck ([#503](https://github.com/m1k1o/neko/pull/503), by [@Garrulousbrevity](https://github.com/Garrulousbrevity)).

## [n.eko v3.0.3](https://github.com/m1k1o/neko/releases/tag/v3.0.3) {#v3.0.3}

### Fixes {#v3.0.3-fixes}
- Fixed legacy WebSocket mode to correctly forward ping messages ([#506](https://github.com/m1k1o/neko/issues/506)).
- Fixed legacy handler logging and WebSocket error unwrapping.
- Fixed legacy pipeline generation when a specific codec is configured.

### Misc {#v3.0.3-misc}
- Configuration option names updated and scripts moved to a dedicated `scripts/` folder.
- Updated Docker image naming convention to include version information.
- Documentation improvements: legacy mode explanation, V2 migration guide, available encoders overview, NVIDIA GPU examples ([#502](https://github.com/m1k1o/neko/issues/502)), filetransfer migration guide.

## [n.eko v3.0.2](https://github.com/m1k1o/neko/releases/tag/v3.0.2) {#v3.0.2}

### Misc {#v3.0.2-misc}
- Added `net.m1k1o.neko.api-version` label to Docker images for easier API version detection.
- Legacy handler key/button action events are now only logged at trace level, reducing log verbosity.

## [n.eko v3.0.1](https://github.com/m1k1o/neko/releases/tag/v3.0.1) {#v3.0.1}

### New Features {#v3.0.1-feats}
- Added mobile keyboard icon for opening the on-screen keyboard on touch devices ([#497](https://github.com/m1k1o/neko/pull/497)).
- Chat now auto-scrolls to the latest message on mobile ([#496](https://github.com/m1k1o/neko/pull/496)).

### Fixes {#v3.0.1-fixes}
- Fixed ICE Lite being incorrectly enabled by default (it should be disabled).
- Fixed supervisord configuration to use the new `server.static` flag.

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
