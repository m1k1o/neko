---
description: Configuration related to the Neko plugins.
---

import { Def, Opt } from '@site/src/components/Anchor';
import { ConfigurationTab } from '@site/src/components/Configuration';
import configOptions from './help.json';

# Plugins Configuration

Neko allows you to extend its functionality by using [plugins](https://pkg.go.dev/plugin). Go plugins come with a lot of benefits as well as some limitations. The main advantage is that you can extend the functionality of the application without recompiling the main application. But the main limitation is that you need to use the same Go version and all dependencies with the same version as the main application.

<ConfigurationTab options={configOptions} filter={[
  'plugins.enabled',
  'plugins.required',
  'plugins.dir',
]} comments={false} />

- <Def id="enabled" /> enables the plugin support. If set to `false`, the plugins are not loaded.
- <Def id="required" /> makes the plugin loading mandatory, meaning that if a plugin fails to load, the application will not start.
- <Def id="dir" /> refers to the directory where the plugins are stored.

:::danger
External plugins are experimental and may not work as expected. They will be replaced with [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) in the future.
:::

There exist a few pre-loaded internal plugins that are shipped with Neko:

## Chat Plugin {#chat}

The chat plugin is a simple pre-loaded internal plugin that allows you to chat with other users in the same session. The chat messages are sent to the server and then broadcasted to all users in the same session.

<ConfigurationTab options={{
  'chat.enabled': true
}} />

- <Def id="chat.enabled" /> enables the chat support. If set to `false`, the chat is disabled.

The chat plugin extends user profile and room settings by adding the following fields:

```yaml
plugins:
  chat.can_send: true
  chat.can_receive: true
```

- `chat.can_send` in the room settings context controls whether the chat messages can be sent by any user in the room, and in the user's profile context controls whether the user can send chat messages.
- `chat.can_receive` in the room settings context controls whether the chat messages can be received by any user in the room, and in the user's profile context controls whether the user can receive chat messages.

## File Transfer Plugin {#filetransfer}

The file transfer plugin is a simple pre-loaded internal plugin that allows you to transfer files between the client and the server. The files are uploaded to the server and then downloaded by the client.


<ConfigurationTab options={{
  'filetransfer.enabled': true,
  'filetransfer.dir': './uploads',
  'filetransfer.refresh_interval': {
    type: 'duration',
    defaultValue: '30s',
  },
}} />

- <Def id="filetransfer.enabled" /> enables the file transfer support. If set to `false`, the file transfer is disabled.
- <Def id="filetransfer.dir" /> refers to the directory where the files are stored.
- <Def id="filetransfer.refresh_interval" /> refers to the interval at which the file list is refreshed.

The file transfer plugin extends user profile and room settings by adding the following fields:

```yaml
plugins:
  filetransfer.enabled: true
```

- `filetransfer.enabled` in the room settings context controls whether the file transfer is enabled for any user in the room, and in the user's profile context controls whether the user can transfer files.
