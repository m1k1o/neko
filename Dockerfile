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

FROM cave-xfce:202403162131

COPY --from=server /src/bin/neko /usr/bin/neko

RUN set -eux; apt-get update;

RUN cd /tmp; wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb; \
    dpkg -i google-chrome-stable_current_amd64.deb || true; \
    apt --fix-broken -y install;

# RUN cd /tmp; \
#     wget http://mirrors.kernel.org/ubuntu/pool/multiverse/f/fdk-aac/libfdk-aac1_0.1.6-1_amd64.deb; \
#     apt install ./libfdk-aac1_*_amd64.deb;

RUN cd /tmp; \
    wget https://launchpad.net/ubuntu/+archive/primary/+files/libfdk-aac2_2.0.2-1_amd64.deb; \
    sudo apt install ./libfdk-aac2_*_amd64.deb;

RUN cd /tmp; \
    wget https://dl.strem.io/shell-linux/v4.4.135/stremio_4.4.135-1_amd64.deb; \
    dpkg -i stremio_*.deb || true; \
    apt --fix-broken -y install; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*
