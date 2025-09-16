ARG BASE_IMAGE=golang:1.24-trixie
FROM $BASE_IMAGE AS server

WORKDIR /src

#
# install dependencies
ENV DEBIAN_FRONTEND=noninteractive
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        libx11-dev libxrandr-dev libxtst-dev libgtk-3-dev libxcvt-dev \
        libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev; \
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
