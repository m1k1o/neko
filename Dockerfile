FROM golang:1.15-buster as server
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
    git clone https://github.com/jtanx/libclipboard; \
    cd libclipboard; \
    cmake .; \
    make -j4; \
    make install; \
    rm -rf /tmp/libclipboard; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# build server
COPY . .
RUN go get -v -t -d . && go build -o bin/neko -i cmd/neko/main.go

ENTRYPOINT [ "bin/neko" ]
