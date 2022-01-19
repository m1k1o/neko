# How to contribute to neko

If you want to contribute, but do not want to install anything on your host system, we got you covered. You only need docker. Technically, it could be done using vs code development in container, but this is more fun:).

You need to copy `.env.default` to `.env` and customize values.

## Step 1: Building server

- `./build` - You can use this command to build your specified `SERVER_TAG` along with base image.

If you want, you can build other tags. `base` tag needs to be build first:

- `./build base`
- `./build firefox`
- `./build chromium`
- `./build google-chrome`
- etc...

## Step 2: Starting server

- `./start-server` - Starting server image you specified in `.env`.
- `./start-server -r` - Shortcut for rebuilding server binary and then starting.

If you are changing something in the server code, you do not want to rebuild container each time. You can just rebuild your binary:

- `./rebuild-server` - Rebuild only server binary.
- `./rebuild-server -f` - Force to rebuild whole Golang environment (you should do this only of you change some dependencies).

## Step 3: Serving client

- `./serve-client` - Serving vue.js client.
- `./serve-client -i` - Install all dependencies.

## Debug

You can navigate to `CLIENT_PORT` and see live client there. It will be connected to your local server on `SERVER_PORT`.

If you are leaving client as is and not changing it, you don't need to start `./serve-client` and you can access server's GUI directly on `SERVER_PORT`.

Feel free to open new PR.
