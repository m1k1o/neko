# Introduction

Neko is an open-source self-hosted virtual browser solution that allows multiple users to share a single web browser instance remotely. It is designed for use cases such as collaborative browsing, remote access to web-based applications, and private cloud-based browsing.

## Key Features {#features}

- **Multi-User Collaboration** – Multiple users can interact with the same browser session.
- **Audio & Video Streaming** – Real-time streaming of the browser’s output with low latency.
- **Secure & Self-Hosted** – You control the server, ensuring privacy and security.
- **GPU Acceleration** – Supports hardware acceleration for improved performance.
- **Customization Options** – Configure bookmarks, extensions, persistent data, and more.

## Use Cases {#use-cases}

- **Remote Browsing** – Access a web browser from any device without installing software.
- **Watch Parties** – Stream content together with friends and interact in real time.
- **Web Development & Testing** – Test websites in a controlled browser environment.
- **Cloud-Based Browsing** – Securely browse the web from a dedicated remote environment.

## Supported Platforms {#platforms}

Neko runs on various platforms, including:

- **Linux & Docker** – Easy deployment using Docker containers.
- **Cloud Services** – Deployable on AWS, Azure, Google Cloud, and other providers.
- **Raspberry Pi & ARM Devices** – Optimized versions for embedded and low-power hardware.

Explore the documentation to learn how to deploy, configure, and optimize Neko for your use case.

## Next Steps {#next}

import DocCardList from '@theme/DocCardList';
import {useCurrentSidebarCategory} from '@docusaurus/theme-common';

<DocCardList items={useCurrentSidebarCategory().items.filter((item) => item.docId !== 'introduction')} />

