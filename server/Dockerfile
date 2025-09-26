ARG BASE_IMAGE=golang:1.24-bullseye
FROM $BASE_IMAGE AS server

WORKDIR /src

#
# install dependencies
ENV DEBIAN_FRONTEND=noninteractive
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        libx11-dev libxrandr-dev libxtst-dev libgtk-3-dev \
        libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev; \
    #
    # install libxcvt-dev (not available in debian:bullseye)
    ARCH=$(dpkg --print-architecture); \
    wget http://ftp.de.debian.org/debian/pool/main/libx/libxcvt/libxcvt-dev_0.1.2-1_${ARCH}.deb; \
    wget http://ftp.de.debian.org/debian/pool/main/libx/libxcvt/libxcvt0_0.1.2-1_${ARCH}.deb; \
    apt-get install --no-install-recommends ./libxcvt0_0.1.2-1_${ARCH}.deb ./libxcvt-dev_0.1.2-1_${ARCH}.deb; \
    rm ./libxcvt0_0.1.2-1_${ARCH}.deb ./libxcvt-dev_0.1.2-1_${ARCH}.deb; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

ARG GIT_COMMIT
ARG GIT_BRANCH
ARG GIT_TAG

#
# build server
COPY . .
RUN ./build

#
# artifacts from this stage
# COPY --from=server /src/bin/plugins/ $NEKO_PLUGINS_DIR
# COPY --from=server /src/bin/neko /usr/bin/neko
