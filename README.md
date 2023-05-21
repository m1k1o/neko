# neko
This app uses WebRTC to stream a desktop inside of a docker container. Client can be found here: [demodesk/neko-client](https://github.com/demodesk/neko-client).

For **community edition** neko with GUI and _plug & play_ deployment visit [m1k1o/neko](https://github.com/m1k1o/neko).

### **m1k1o/neko** vs **demodesk/neko**, why do we have two of them?

This project started as a fork of [m1k1o/neko](https://github.com/m1k1o/neko). But over time, development went way ahead of the original one in terms of features, updates and refactoring. The goal is to rebase [m1k1o/neko](https://github.com/m1k1o/neko) repository onto this one and move all extra features (such as chat and emotes) to a standalone plugin.

- This project is aimed to be the engine providing foundation for all applications that are streaming desktop environment using WebRTC to the browser.
- [m1k1o/neko](https://github.com/m1k1o/neko) is meant to be self-hosted replacement for [rabb.it](https://en.wikipedia.org/wiki/Rabb.it): Community edition with well-known GUI, all the social functions (such as chat and emotes) and easy deployment.

Notable differences to the [m1k1o/neko](https://github.com/m1k1o/neko) are:

- Go plugin support.
- Multiple encoding qualities simulcast.
   - Bandwidth estimation and adaptive quality.
- Custom screen size (with automatic sync).
- Single cursor for host - cursor image proxying.
- Custom cursor style/badge for participants.
- Inactive cursors (participants that are not hosting).
- Fallback mode and reconnection improvements:
  - Watching using screencasting.
  - Controlling using websockets.
- Members handling:
  - Access control (view, interactivity, clipboard).
  - Posibility to add external members providers.
  - Persistent login (using cookies).
- Drag and drop passthrough.
- File upload passthrough (experimental).
- Microphone passthrough.
- Webcam passthrough (experimental).
- Bi-directional text/html clipboard.
- Keyboard layouts/variants.
- Metrics and REST API.

## Docs

*TBD.*
