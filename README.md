<div align="center">
  <a href="https://github.com/m1k1o/neko" title="Neko's Github repository.">
    <img src="https://raw.githubusercontent.com/m1k1o/neko/master/docs/_media/logo.png" width="450" height="auto"/>
  </a>
  <p align="center">
    <img src="https://img.shields.io/github/v/release/m1k1o/neko" alt="release">
    <img src="https://img.shields.io/github/license/m1k1o/neko" alt="license">
    <img src="https://img.shields.io/docker/pulls/m1k1o/neko" alt="pulls">
    <img src="https://img.shields.io/github/issues/m1k1o/neko" alt="issues">
    <a href="https://discord.gg/3U6hWpC">
      <img src="https://discordapp.com/api/guilds/665851821906067466/widget.png" alt="Chat on discord">
    </a>
    <a href="https://github.com/m1k1o/neko/actions">
      <img src="https://github.com/m1k1o/neko/actions/workflows/build.yml/badge.svg" alt="build">
    </a>
  </p>
  <br/>
  <br/>
  <img src="https://i.imgur.com/ZSzbQr7.gif" width="650" height="auto"/>
  <br/>
  <br/>
</div>

# n.eko

This app uses Web RTC to stream a desktop inside of a docker container, original author made this because [rabb.it](https://en.wikipedia.org/wiki/Rabb.it) went under and his internet could not handle streaming and discord kept crashing when his friend attempted to. He just wanted to watch anime with his friends ლ(ಠ益ಠლ) so he started digging throughout the internet and found a few *kinda* clones, but none of them had the virtual browser, then he found [Turtus](https://github.com/Khauri/Turtus) and he was able to figure out the rest.

Then I found [this](https://github.com/nurdism/neko) project and started to dig into it. I really liked the idea of having collaborative browser browsing together with mutliple people, so I created a fork. Initially, I wanted to merge my changes to the upstream repository, but the original author did not have time for this project anymore and it got eventually archived.

### Features

  * Text Chat (With basic markdown support, discord flavor)
  * Admin users (Kick, Ban & Force Give/Release Controls)
  * Clipboard synchronization (on [supported browsers](https://developer.mozilla.org/en-US/docs/Web/API/Clipboard/readText))
  * Emote overlay
  * Ignore user (chat and emotes)
  * Persistent settings

### Why n.eko?

I like cats 🐱 (`Neko` is the Japanese word for cat), I'm a weeb/nerd.

***But why the cat butt?*** Because cats are *assholes*, but you love them anyways.

# Multiple rooms

For n.eko room management software, visit [neko-rooms](https://github.com/m1k1o/neko-rooms).

It also offers zero-knowledge [installation script](https://github.com/m1k1o/neko-rooms/#zero-knowledge-installation).

# Documentation

* [Getting Started](https://neko.m1k1o.net/#/getting-started/)
  * [Quick Start](https://neko.m1k1o.net/#/getting-started/quick-start)
  * [Examples](https://neko.m1k1o.net/#/getting-started/examples)
  * [Reverse Proxy](https://neko.m1k1o.net/#/getting-started/reverse-proxy)
  * [Configuration](https://neko.m1k1o.net/#/getting-started/configuration)
  * [Troubleshooting](https://neko.m1k1o.net/#/getting-started/troubleshooting)
* [Mobile Support](https://neko.m1k1o.net/#/mobile-support)
* [Contributing](https://neko.m1k1o.net/#/contributing)
  * [Non Goals](https://neko.m1k1o.net/#/non-goals)
  * [Technologies](https://neko.m1k1o.net/#/technologies)
* [Changelog](https://neko.m1k1o.net/#/changelog)

# How to contribute? How to build?

Navigate to [.docker](.docker) folder for further information.
