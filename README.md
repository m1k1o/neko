<div align="center">
  <a href="https://n.eko.moe/#/" ><img src="https://raw.githubusercontent.com/nurdism/neko/master/docs/_media/logo.png" width="450" height="auto"/></a>
  <p align="center">
    <img src="https://img.shields.io/github/v/release/nurdism/neko" alt="release">
    <img src="https://img.shields.io/github/license/nurdism/neko" alt="license">
    <img src="https://img.shields.io/docker/pulls/nurdism/neko" alt="pulls">
    <img src="https://img.shields.io/github/issues/nurdism/neko" alt="issues">
    <a href="https://discord.gg/3U6hWpC" ><img src="https://discordapp.com/api/guilds/665851821906067466/widget.png" alt="Chat on discord"><a/>
    <a href="https://github.com/nurdism/neko/actions" ><img src="https://github.com/nurdism/neko/workflows/deploy/badge.svg" alt="build"><a/>
  </p>
  <br/>
  <br/>
  <img src="https://i.imgur.com/ZSzbQr7.gif" width="650" height="auto"/>
  <br/>
  <br/>
</div>

# n.eko (m1k1o fork)
This app uses Web RTC to stream a desktop inside of a docker container. This is fork of https://github.com/nurdism/neko.

## Differences to original repository.

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
