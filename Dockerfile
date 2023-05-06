#
# STAGE 1: SERVER
#
FROM golang:1.20-bullseye as server
WORKDIR /src

#
# install dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends git cmake make libx11-dev libxrandr-dev libxtst-dev \
    libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly; \
    #
    # install libclipboard
    set -eux; \
    cd /tmp; \
    git clone --depth=1 https://github.com/jtanx/libclipboard; \
    cd libclipboard; \
    cmake .; \
    make -j4; \
    make install; \
    rm -rf /tmp/libclipboard; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

COPY server/go.mod ./
COPY server/go.sum ./
RUN go mod download

#
# build server
COPY server/ .
RUN ./build

FROM ghcr.io/m1k1o/neko/intel-firefox:latest

RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends i965-va-driver-shaders; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

COPY --from=server /src/bin/neko /usr/bin/neko
