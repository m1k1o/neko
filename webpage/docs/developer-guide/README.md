# Developer Guide

:::info
This guide is Work in Progress. It is not complete and will be updated over time.
:::

## Dependencies

- [node.js](https://nodejs.org/) and [npm](https://www.npmjs.com/) (for building the frontend).
- [go](https://golang.org/) (for building the server).
- [gstreamer](https://gstreamer.freedesktop.org/) (for video processing).
  ```shell
  sudo apt-get install libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
      gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
      gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
      gstreamer1.0-pulseaudio;
  ```
- [x.org](https://www.x.org/) (for X11 server).
  ```shell
  sudo apt-get install libx11-dev libxrandr-dev libxtst-dev libxcvt-dev xorg;
  ```
- [pulseaudio](https://www.freedesktop.org/wiki/Software/PulseAudio/) (for audio support).
  ```shell
  sudo apt-get install pulseaudio;
  ```
- other dependencies:
  ```shell
  sudo apt-get install xdotool xclip libgtk-3-0 libgtk-3-dev libopus0 libvpx6;
  ```

## Next Steps

import DocCardList from '@theme/DocCardList';
import {useCurrentSidebarCategory} from '@docusaurus/theme-common';

<DocCardList items={useCurrentSidebarCategory().items}/>
