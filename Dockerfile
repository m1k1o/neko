#
# Stage 1: Build.
#
FROM golang:1.18-bullseye as build
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
# Stage 2: Runtime.
#
FROM debian:bullseye-slim as runtime

#
# set custom user
ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# install dependencies
ENV DEBIAN_FRONTEND=noninteractive
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
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
        gstreamer1.0-pulseaudio; \
    #
    # create a non-root user
    groupadd --gid $USER_GID $USERNAME; \
    useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME; \
    adduser $USERNAME audio; \
    adduser $USERNAME video; \
    adduser $USERNAME pulse; \
    #
    # setup pulseaudio
    mkdir -p /home/$USERNAME/.config/pulse/; \
    echo "default-server=unix:/tmp/pulseaudio.socket" > /home/$USERNAME/.config/pulse/client.conf; \
    #
    # workaround for an X11 problem: http://blog.tigerteufel.de/?p=476
    mkdir /tmp/.X11-unix; \
    chmod 1777 /tmp/.X11-unix; \
    chown $USERNAME /tmp/.X11-unix/; \
    #
    # make directories for neko
    mkdir -p /etc/neko /var/www /var/log/neko; \
    chmod 1777 /var/log/neko; \
    chown $USERNAME /var/log/neko/; \
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
        fonts-indic; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy runtime configs
COPY --chown=neko:neko runtime/.Xresources /home/$USERNAME/.Xresources
COPY runtime/dbus /usr/bin/dbus
COPY runtime/default.pa /etc/pulse/default.pa
COPY runtime/supervisord.conf /etc/neko/supervisord.conf
COPY runtime/xorg.conf /etc/neko/xorg.conf

#
# copy runtime folders
COPY --chown=neko:neko runtime/icon-theme /home/$USERNAME/.icons/default
COPY runtime/fontconfig/* /etc/fonts/conf.d/
COPY runtime/fonts /usr/local/share/fonts

#
# set default envs
ENV USER=$USERNAME
ENV DISPLAY=:99.0
ENV NEKO_SERVER_BIND=:8080
ENV NEKO_PLUGINS_ENABLED=true
ENV NEKO_PLUGINS_DIR=/etc/neko/plugins/

#
# copy plugins from previous stage
COPY --from=build /src/bin/plugins/ $NEKO_PLUGINS_DIR

#
# copy executable from previous stage
COPY --from=build /src/bin/neko /usr/bin/neko

#
# add healthcheck
HEALTHCHECK --interval=10s --timeout=5s --retries=8 \
    CMD wget -O - http://localhost:${NEKO_SERVER_BIND#*:}/health || exit 1

#
# run neko
CMD ["/usr/bin/supervisord", "-s", "-c", "/etc/neko/supervisord.conf"]
