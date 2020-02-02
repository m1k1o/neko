FROM debian:stretch-slim

#
# avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# install libclipboard
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates git cmake pkg-config build-essential libx11-dev ; \
    cd /tmp ; \
    git clone https://github.com/jtanx/libclipboard ; \
    cd libclipboard ; \
    cmake -DBUILD_SHARED_LIBS=ON -DLIBCLIPBOARD_FORCE_X11=on -DLIBCLIPBOARD_ADD_SOVERSION=ON --prefix=/usr/local . ; \
    make -j4; \
    make install; \ 
    rm -rf /tmp/libclipboard ; \
    #
    # clean up
    apt-get autoremove -y git cmake pkg-config build-essential libx11-dev; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# install neko dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends wget ca-certificates pulseaudio openbox dbus-x11 xvfb supervisor; \
    apt-get install -y --no-install-recommends libxv1 libopus0 libvpx4; \
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
    mkdir /tmp/.X11-unix; chmod 1777 /tmp/.X11-unix; chown $USERNAME /tmp/.X11-unix/; \
    #
    # make directories for neko
    mkdir -p /etc/neko /var/www /var/log/neko; chmod 1777 /var/log/neko; chown $USERNAME /var/log/neko/;  \
    chown -R $USERNAME:$USERNAME /home/$USERNAME; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# add gst to env
ENV PATH=/gst/local/bin:$PATH
ENV LD_LIBRARY_PATH=/gst/local/lib:$LD_LIBRARY_PATH
ENV PKG_CONFIG_PATH=/gst/local/lib/pkgconfig:$PKG_CONFIG_PATH

#
# copy gst
COPY .build/gst/local /gst/local/

#
# env
ENV USER=$USERNAME
ENV SCREEN_WIDTH=1280
ENV SCREEN_HEIGHT=720
ENV SCREEN_DEPTH=24
ENV DISPLAY=:99.0

#
# copy configuation files
COPY .docker/files/dbus /usr/bin/dbus
COPY .docker/files/openbox.xml /etc/neko/openbox.xml
COPY .docker/files/neko/supervisord.conf /etc/neko/supervisord/neko.conf
COPY .docker/files/supervisord.conf /etc/neko/supervisord.conf
COPY .docker/files/default.pa /etc/pulse/default.pa

#
# neko files
COPY client/dist/ /var/www
COPY server/bin/neko /usr/bin/neko

#
# neko env
ENV NEKO_PASSWORD=neko
ENV NEKO_ADMIN=admin
ENV NEKO_BIND=:8080

#
# run neko
CMD ["/usr/bin/supervisord", "-c", "/etc/neko/supervisord.conf"]