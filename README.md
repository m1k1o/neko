<div align="center">
  <a href="https://github.com/m1k1o/neko" title="Neko's Github repository.">
    <img src="https://neko.m1k1o.net/img/logo.png" width="400" height="auto"/>
  </a>
  <p align="center">
    <a href="https://github.com/m1k1o/neko/releases">
      <img src="https://img.shields.io/github/v/release/m1k1o/neko" alt="release">
    </a>
    <a href="https://github.com/m1k1o/neko/blob/master/LICENSE">
      <img src="https://img.shields.io/github/license/m1k1o/neko" alt="license">
    </a>
    <a href="https://hub.docker.com/r/m1k1o/neko">
      <img src="https://img.shields.io/docker/pulls/m1k1o/neko" alt="pulls">
    </a>
    <a href="https://github.com/m1k1o/neko/issues">
      <img src="https://img.shields.io/github/issues/m1k1o/neko" alt="issues">
    </a>
    <a href="https://github.com/sponsors/m1k1o">
      <img src="https://img.shields.io/badge/-sponsor-red" alt="issues">
    </a>
    <a href="https://discord.gg/HXQJmqNJMz">
      <img src="https://discordapp.com/api/guilds/665851821906067466/widget.png" alt="Chat on discord">
    </a>
    <a href="https://hellogithub.com/repository/4536d4546af24196af3f08a023dfa007" target="_blank">
      <img src="https://abroad.hellogithub.com/v1/widgets/recommend.svg?rid=4536d4546af24196af3f08a023dfa007&claim_uid=0x19e4dJwD83aW2&theme=small" alt="Featured｜HelloGitHub" />
    </a>
    <a href="https://github.com/m1k1o/neko/actions">
      <img src="https://github.com/m1k1o/neko/actions/workflows/ghcr.yml/badge.svg" alt="build">
    </a>
  </p>
  <img src="https://neko.m1k1o.net/img/intro.gif" width="650" height="auto"/>
</div>

# n.eko

Welcome to Neko, a self-hosted virtual browser that runs in Docker and uses WebRTC technology. Neko is a powerful tool that allows you to **run a fully-functional browser in a virtual environment**, giving you the ability to **access the internet securely and privately from anywhere**. With Neko, you can browse the web, **run applications**, and perform other tasks just as you would on a regular browser, all within a **secure and isolated environment**. Whether you are a developer looking to test web applications, a **privacy-conscious user seeking a secure browsing experience**, or simply someone who wants to take advantage of the **convenience and flexibility of a virtual browser**, Neko is the perfect solution.

In addition to its security and privacy features, Neko offers the **ability for multiple users to access it simultaneously**. This makes it an ideal solution for teams or organizations that need to share access to a browser, as well as for individuals who want to use **multiple devices to access the same virtual environment**. With Neko, you can **easily and securely share access to a browser with others**, without having to worry about maintaining separate configurations or settings. Whether you need to **collaborate on a project**, access shared resources, or simply want to **share access to a browser with friends or family**, Neko makes it easy to do so.

Neko is also a great tool for **hosting watch parties** and interactive presentations. With its virtual browser capabilities, Neko allows you to host watch parties and presentations that are **accessible from anywhere**, without the need for in-person gatherings. This makes it easy to **stay connected with friends and colleagues**, even when you are unable to meet in person. With Neko, you can easily host a watch party or give an **interactive presentation**, whether it's for leisure or work. Simply invite your guests to join the virtual environment, and you can share the screen and **interact with them in real-time**.

## About

This app uses WebRTC to stream a desktop inside of a docker container, original author made this because [rabb.it](https://en.wikipedia.org/wiki/Rabb.it) went under and his internet could not handle streaming and discord kept crashing when his friend attempted to. He just wanted to watch anime with his friends ლ(ಠ益ಠლ) so he started digging throughout the internet and found a few *kinda* clones, but none of them had the virtual browser, then he found [Turtus](https://github.com/Khauri/Turtus) and he was able to figure out the rest.

Then I found [this](https://github.com/nurdism/neko) project and started to dig into it. I really liked the idea of having collaborative browser browsing together with multiple people, so I created a fork. Initially, I wanted to merge my changes to the upstream repository, but the original author did not have time for this project anymore and it got eventually archived.

## Use-cases and comparison

M1 is a collaborative, containerized Chromium browser streamed to multiple users
with WebRTC. It intentionally supports one browser image so that performance,
hardware encoding, networking and outbound-proxy behavior have one testable
runtime target.

Primary use case is connecting with multiple people, leveraging real time synchronization and interactivity:
- **Watch party** - watching video content together with multiple people and reacting to it (chat, emotes) - open source alternative to [giggl.app](https://giggl.app/) or [hyperbeam](https://watch.hyperbeam.com).
- **Interactive presentation** - not only screen sharing, but others can control the screen.
- **Collaborative tool** - brainstorming ideas, cobrowsing, code debugging together.
- **Support/Teaching** - interactively guiding people in controlled environment.
- **Embed Chromium** - integrate a controlled remote browser in another product.
- **Persistent or temporary browser sessions** - keep profiles in a mounted volume
  or discard them with the container.
- **Support and teaching** - collaboratively guide users through a web workflow.

Compared to clientless remote desktop gateways such as [Apache Guacamole](https://guacamole.apache.org/) or [noVNC](https://novnc.com/), Neko provides:
- **Smooth video** because it uses WebRTC and not images sent over WebSockets.
- **Built in audio** support, what is not part of Apache Guacamole or noVNC.
- **Multi-participant control**, what is not natively supported by Apache Guacamole or noVNC.

### Supported runtime

<div align="center">
  <a href="https://neko.m1k1o.net/docs/v3/installation/docker-images#chromium">
    <img src="https://neko.m1k1o.net/img/icons/chromium.svg" title="ghcr.io/m1k1o/neko/chromium" width="60" height="auto"/>
  </a>
</div>

### Why neko?

I like cats 🐱 (`Neko` is the Japanese word for cat), I'm a weeb/nerd.

***But why the cat butt?*** Because cats are *assholes*, but you love them anyways.

## Multiple rooms

For neko room management software, visit [neko-rooms](https://github.com/m1k1o/neko-rooms).

It also offers [Zero-knowledge installation (with HTTPS)](https://github.com/m1k1o/neko-rooms/?tab=readme-ov-file#zero-knowledge-installation-with-https).

## Documentation

Full documentation is available at [neko.m1k1o.net](https://neko.m1k1o.net/). Key sections include:

- [M1 migration](https://neko.m1k1o.net/docs/v3/migration-from-v2)
- [Getting Started](https://neko.m1k1o.net/docs/v3/quick-start)
- [Installation](https://neko.m1k1o.net/docs/v3/installation)
- [Examples](https://neko.m1k1o.net/docs/v3/installation/examples)
- [Configuration](https://neko.m1k1o.net/docs/v3/configuration)
- [Frequently Asked Questions](https://neko.m1k1o.net/docs/v3/faq)
- [Troubleshooting](https://neko.m1k1o.net/docs/v3/troubleshooting)

## How to Contribute

Contributions are welcome! Check the [Contributing Guide](https://neko.m1k1o.net/contributing) for details.

## Support

If you find Neko useful, consider supporting the project via [GitHub Sponsors](https://github.com/sponsors/m1k1o).
