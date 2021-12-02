# Contributing

1. Fork the [project](https://github.com/m1k1o/neko).

2. Navigate to [.docker/README.md](https://github.com/m1k1o/neko/tree/master/.docker) for further information.

3. Edit files in your branch.

4. Submit a pull request explaining the improvements.

## Server build dependencies

If you want to compile goalng code locally, you must install additional dependencies in order for it to compile.

```shell
apt-get install -y --no-install-recommends libx11-dev libxrandr-dev libxtst-dev libgstreamer1.0-dev
```

Libclipboard files can be retrieved from `neko_dev_server` container:

```shell
mkdir -p /usr/local/lib/pkgconfig/ /usr/local/include/
docker cp neko_dev_server:/usr/local/lib/libclipboard.a /usr/local/lib/
docker cp neko_dev_server:/usr/local/lib/pkgconfig/libclipboard.pc /usr/local/lib/pkgconfig/
docker cp neko_dev_server:/usr/local/include/libclipboard-config.h /usr/local/include/
docker cp neko_dev_server:/usr/local/include/libclipboard.h /usr/local/include/
```
