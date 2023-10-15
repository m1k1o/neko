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

#
# build server
COPY server/ .
RUN ./build

#
# STAGE 2: CLIENT
#
FROM node:18-bullseye-slim as client
WORKDIR /src

#
# install dependencies
COPY client/package*.json ./
RUN npm install

#
# build client
COPY client/ .
RUN npm run build

#
# STAGE 3: RUNTIME
#
FROM debian:bullseye-slim

#
# set custom user
ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN set -eux; \
    apt-get update; \
    #
    # install dependencies
    apt-get install -y --no-install-recommends pulseaudio dbus-x11 xserver-xorg-video-dummy; \
    apt-get install -y --no-install-recommends libcairo2 libxcb1 libxrandr2 libxv1 libopus0 libvpx6; \
    #
    # gst
    apt-get install -y --no-install-recommends \
        libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
        gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
        gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \ 
        gstreamer1.0-pulseaudio; \
    #
    # create a non-root user
    groupadd --gid $USER_GID $USERNAME; \
    useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME; \
    #
    # make directories for neko
    mkdir -p /etc/neko /var/www /var/log/neko; \
    chmod 1777 /var/log/neko; \
    chown -R $USERNAME:$USERNAME /home/$USERNAME; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# set default envs
ENV USER=$USERNAME
ENV NEKO_PASSWORD=neko
ENV NEKO_PASSWORD_ADMIN=admin
ENV NEKO_BIND=:8080

#
# copy static files from previous stages
COPY --from=server /src/bin/neko /usr/bin/neko
COPY --from=client /src/dist/ /var/www

USER $USERNAME

ENTRYPOINT [ "/usr/bin/neko" ]

CMD [ "serve", "--static", "/var/www" ]
