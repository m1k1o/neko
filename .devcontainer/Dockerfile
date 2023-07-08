#
# Stage 0: Build xserver-xorg-video-dummy 0.3.8-2 with RandR support.
#
FROM debian:bullseye-slim as xserver-xorg-video-dummy

WORKDIR /usr/local/src

ENV DEBIAN_FRONTEND=noninteractive
RUN set -eux; \
    cp /etc/apt/sources.list /etc/apt/sources.list~; \
    sed -Ei 's/^deb /deb-src /' /etc/apt/sources.list; \
    cat /etc/apt/sources.list~ >> /etc/apt/sources.list; \
    apt-get update; \
    apt-get install -y dpkg-dev git; \
    apt-get build-dep -y xserver-xorg-video-dummy; \
    git clone --depth 1 --branch xserver-xorg-video-dummy-1_0.3.8-2 https://salsa.debian.org/xorg-team/driver/xserver-xorg-video-dummy; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

COPY runtime/xdummy-randr.patch /tmp/xdummy-randr.patch

RUN set -eux; \
    cd xserver-xorg-video-dummy; \
    patch -p1 < /tmp/xdummy-randr.patch; \
    bash ./autogen.sh; \
    make; \
    make install;

# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.166.0/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version: 1, 1.16, 1.15
ARG VARIANT="1"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Option] Install Node.js
ARG INSTALL_NODE="true"
ARG NODE_VERSION="lts/*"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# build dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        libx11-dev libxrandr-dev libxtst-dev libgtk-3-dev \
        libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev; \
    # install libxcvt-dev (not available in debian:bullseye)
    wget http://ftp.de.debian.org/debian/pool/main/libx/libxcvt/libxcvt-dev_0.1.2-1_amd64.deb; \
    wget http://ftp.de.debian.org/debian/pool/main/libx/libxcvt/libxcvt0_0.1.2-1_amd64.deb; \
    apt-get install  --no-install-recommends ./libxcvt0_0.1.2-1_amd64.deb ./libxcvt-dev_0.1.2-1_amd64.deb;

# runtime dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        wget ca-certificates supervisor \
        pulseaudio dbus-x11 xserver-xorg-video-dummy \
        libcairo2 libxcb1 libxrandr2 libxv1 libopus0 libvpx6 \
        #
        # needed for profile upload preStop hook
        zip curl \
        #
        # file chooser handler, clipboard, drop
        xdotool xclip libgtk-3-0 \
        #
        # gst
        gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
        gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
        gstreamer1.0-pulseaudio;
    # libxcvt already installed

# dev runtime dependencies
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        xfce4 xfce4-terminal firefox-esr sudo;

# configure runtime
ARG USERNAME=neko
ARG USER_UID=1001
ARG USER_GID=$USER_UID
RUN set -eux; \
    #
    # create a non-root user
    groupadd --gid $USER_GID $USERNAME; \
    useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME; \
    adduser $USERNAME audio; \
    adduser $USERNAME video; \
    adduser $USERNAME pulse; \
    #
    # add sudo support
    echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME; \
    chmod 0440 /etc/sudoers.d/$USERNAME; \
    #
    # workaround for an X11 problem: http://blog.tigerteufel.de/?p=476
    mkdir /tmp/.X11-unix; \
    chmod 1777 /tmp/.X11-unix; \
    chown $USERNAME /tmp/.X11-unix/; \
    #
    # make directories for neko
    mkdir -p /etc/neko /var/www; \
    chown -R $USERNAME:$USERNAME /home/$USERNAME; \
    #
    # install fonts
    apt-get install -y --no-install-recommends \
        # Emojis
        fonts-noto-color-emoji \
        # Chinese fonts
        fonts-arphic-ukai fonts-arphic-uming \
        # Japanese fonts
        fonts-ipafont-mincho fonts-ipafont-gothic \
        # Korean fonts
        fonts-unfonts-core \
        # Indian fonts
        fonts-indic;

# replace version
COPY --from=xserver-xorg-video-dummy /usr/local/lib/xorg/modules/drivers/dummy_drv.so /usr/lib/xorg/modules/drivers/dummy_drv.so

# copy runtime files
COPY runtime/dbus /usr/bin/dbus
COPY runtime/default.pa /etc/pulse/default.pa
COPY runtime/supervisord.conf /etc/neko/supervisord.conf
COPY runtime/xorg.conf /etc/neko/xorg.conf
COPY runtime/icon-theme /home/$USERNAME/.icons/default

# copy dev runtime files
COPY dev/runtime/config.yml /etc/neko/neko.yml
COPY dev/runtime/supervisord.conf /etc/neko/supervisord/dev.conf

# customized scripts
RUN chmod +x /usr/bin/dbus;\
    echo '#!/bin/sh\nsleep infinity' > /usr/bin/neko; \
    chmod +x /usr/bin/neko; \
    echo '#!/bin/sh\nsudo sh -c "export USER='$USERNAME'\nexport HOME=/home/'$USERNAME'\n/usr/bin/supervisord -c /etc/neko/supervisord.conf"' > /usr/bin/deps; \
    chmod +x /usr/bin/deps; \
    touch .env.development;

# set default envs
ENV USER=$USERNAME
ENV DISPLAY=:99.0
ENV PULSE_SERVER=unix:/tmp/pulseaudio.socket
ENV NEKO_SERVER_BIND=:3000
ENV NEKO_WEBRTC_EPR=3001-3004
