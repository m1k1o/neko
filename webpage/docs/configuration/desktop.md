---
description: Configuration related to the Desktop Environment in Neko.
---

import { Def, Opt } from '@site/src/components/Anchor';
import { ConfigurationTab } from '@site/src/components/Configuration';
import configOptions from './help.json';

# Desktop Environment

This section describes how to configure the desktop environment inside neko.

Neko uses the [X Server](https://www.x.org/archive/X11R7.6/doc/man/man1/Xserver.1.xhtml) as the display server with [Openbox](http://openbox.org/wiki/Main_Page) as the default window manager. For audio, [PulseAudio](https://www.freedesktop.org/wiki/Software/PulseAudio/) is used.

<ConfigurationTab options={configOptions} filter={[
  'desktop.display',
  'desktop.screen'
]} comments={false} />

- <Def id="display" /> refers to the X server that is running on the system. If it is not specified, the environment variable `DISPLAY` is used. The same display is referred to in the [Capture](capture#video.display) configuration to capture the screen. In most cases, we want to use the same display for both.
- <Def id="screen" /> refers to the screen resolution and refresh rate. The format is `<width>x<height>@<refresh rate>`. If not specified, the default is `1280x720@30`.

:::tip
Admin can change the resolution in the GUI.
:::

## Input Devices {#input}

Neko uses the [XTEST Extension Library](https://www.x.org/releases/X11R7.7/doc/libXtst/xtestlib.html) to simulate keyboard and mouse events. However, for more advanced input devices like touchscreens, we need to use a custom driver that can be loaded as a plugin to the X server and then neko can connect to it.

:::note
Currently, only touchscreens are supported through the custom driver.
:::

<ConfigurationTab options={configOptions} filter={[
  'desktop.input.enabled',
  'desktop.input.socket'
]} comments={false} />

- <Def id="input.enabled" /> enables the input device support. If not specified, the default is `false`.
- <Def id="input.socket" /> refers to the socket file that the custom driver creates. If not specified, the default is `/tmp/xf86-input-neko.sock`.

:::info
When using Docker, the custom driver is already included in the image and the socket file is created at `/tmp/xf86-input-neko.sock`. Therefore, no additional configuration is needed.
:::

## Unminimize {#unminimize}

Most of the time, only a single application is used in the minimal desktop environment without any taskbar or desktop icons. It could happen that the user accidentally minimizes the application and then it is not possible to restore it. To prevent this, we can use the `unminimize` feature that simply listens for the minimize event and restores the window back to the original state.

<ConfigurationTab options={configOptions} filter={[
  'desktop.unminimize'
]} comments={false} />

## Upload Drop {#upload_drop}

The upload drop is a feature that allows the user to upload files to the application by dragging and dropping them into the application window. The files are then uploaded to the application and the application can process them.

The current approach is to catch the drag and drop events on the client side, upload them to the server along with the coordinates of the drop event, and then open an invisible overlay window on the server that has set the file path to the uploaded file and allows it to be dragged and dropped into the application. Then the mouse events are simulated to drag the file from the overlay window to the application window.

<ConfigurationTab options={configOptions} filter={[
  'desktop.upload_drop'
]} comments={false} />

## File Chooser Dialog {#file_chooser_dialog}

:::danger
This feature is experimental and may not work as expected.
:::

The file chooser dialog is a feature that allows handling the file chooser dialog in the application (for example, when uploading a file) externally. This means that the file chooser dialog is not displayed inside the neko desktop environment, but the neko client is requested to upload the file from the local filesystem.

The current approach is to put the file chooser dialog in the background as soon as it is displayed, prompt the user to upload the file, and then select this file in the file chooser dialog by simulating the keyboard events to navigate to the file and press the open button. **This is very error-prone and may not work as expected.**

<ConfigurationTab options={configOptions} filter={[
  'desktop.file_chooser_dialog'
]} comments={false} />
