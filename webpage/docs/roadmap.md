# Roadmap

The roadmap outlines the future development plans for Neko. It is divided into three phases, each focusing on a different aspect of the project.

## Phase 1 - Server migration to V3 {#phase-1}

This phase was successfully completed with the release of Neko V3.0.0. The [m1k1o/neko](https://github.com/m1k1o/neko) server was merged with the archived [demodesk/neko](https://github.com/demodesk/neko) server, and the new server was released as V3.0.0. A compatibility layer was added to support V2 clients.

## Phase 2 - Client rewrite to V3 {#phase-2}

The client rewrite is the next big step in the development of Neko. The V2 client uses Vue2, which reached [end of life](https://v2.vuejs.org/eol/) a long time ago. The new client will be based on Vue3 and will be more modular and easier to maintain.

While the V2 client focused on the user interface, the V3 client will focus on extensibility in the form of components. This means that the client will be able to be loaded seamlessly in any existing application, and the components will be able to be used in any other Vue3 application. For traditional users, the client will still be available as a standalone application with all the known features.

## Phase 3 - Modularization {#phase-3}

The V3 client and server will be modularized to allow for easier maintenance and extensibility.

- The client should be split into a library TypeScript component **that does not use Vue.js** or any library and can be imported by any project. It should be as easy to integrate into a custom project as embedding a video player. Similar to how [demodesk/neko-client](https://github.com/demodesk/neko-client) is built, but without Vue.js.
- The **connection**, **media streaming**, and **control** should be extracted as an interface so that it can be implemented by various protocols, not just WebSockets+WebRTC. That would elevate this project from just a shared virtual environment to basically a video streaming server with built-in tools for feedback and out-of-band communication (such as natively binding to RDP/VNC protocols, controlling drones/robots/PTZ cameras/industrial devices remotely). Since the controlling layer could be just a plugin, it does not need to rely on only keyboard and mouse but would allow plugging in gamepads, joysticks, or even Virtual Reality glasses (anything).

### Connection {#phase-3-connection}

Neko can connect to the backend using multiple channels. Therefore API users should not be exposed to WebSocket internals.

They should only care about the connection status:
- `connected` - user is connected to the server 
- `connecting` - currently the client is attempting to establish a connection to the server.
- `disconnected` - user is disconnected from the server and there are no attempts to connect to the server. This should always be notified with a reason why it has been disconnected.

And about connection type:
- `none` - no connection is currently used.
- `short_polling` - every X ms the client requests the server for updates.
- `long_polling` - HTTP request is kept open until the server has updates to send to the client. Then the client sends another request.
- `sse` - server sends updates to the client using Server-Sent Events.
- `websocket` - server sends updates to the client using WebSockets.
- ... others (e.g. MQTT...)

### Media streaming {#phase-3-media-streaming}

For media streaming, we implement a similar approach with the following streaming backends:
- `none` - no media streaming is currently streamed.
- `m3u8` - media is streamed using HLS.
- `webrtc` - media is streamed using WebRTC.
- `quic` - media is streamed using QUIC.
- ... others (e.g. RTSP, DASH...)

Various media streaming backends can have various features. For example, WebRTC can have a feature to send media to the server, while HTTP can only receive media from the server.
They can be selected based on the user's device capabilities, network conditions, and server capabilities.
There must be a single interface that all streaming backends must satisfy and it is their only communication channel with the rest of the system.

### Control (Human interface device) {#phase-3-control}

The user can control the target system using various human interface devices. The user can use a keyboard, mouse, gamepad, touch screen, or any other device that can be used to control the system. Custom or virtual devices can be used as well.

Normally in-band feedback should be provided to the user inside the media stream. But there can be cases where out-of-band feedback is required.
- When the user is using a gamepad, it can be used to control the system, but the gamepad can also have a vibration feature that can be used to provide feedback to the user.
- Cursors can be hidden on the screen, or for certain users, therefore they need to be transmitted out-of-band.
- Since there are multiple users, all of them can have custom cursors shown on the screen, but only one cursor can be controlled by the user. Therefore the cursor position must be transmitted out-of-band.
- Changing screen resolution, orientation, or other settings.
- Changing the keyboard layout or modifier keys.
- Setting a host - who is currently controlling the system based on priority and permissions.

Control can use both underlying connections or media streaming for transmitting and receiving control data. For example, WebRTC data channels can be used for transmitting control data in real-time.
