# How to contribute to neko

If you want to contribute, but do not want to install anything on your host system, we got you covered. You only need docker. Technically, it could be done using vs code development in container, but this is more fun:).

## Running server (while developing)

Go to `../server/dev` and run:

- `./build` - Build server binary.
- `./start` - Start server.
- `./rebuild` - Rebuild server binary and restart server while it is running.

## Running client (while developing)

Go to `../client/dev` and run:

- `./npm install` - Install dependencies first.
- `./serve` - Start client with live reload.

## Building a new image after changes

You need to copy `.env.default` to `.env` and customize values.

- `./build` - You can use this command to build base image. It will be used for building other images.

If you want, you can build other tags. `base` tag needs to be build first:

- `./build base`
- `./build firefox`
- `./build chromium`
- `./build google-chrome`
- etc...
