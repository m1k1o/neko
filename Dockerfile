FROM frolvlad/alpine-glibc

ENV GST_VERSION=1.16.2

# build gstreamer
RUN apk add --no-cache --virtual .gst-build-deps \
    build-base bison flex perl python glib-dev zlib-dev \
    opus-dev \
    pulseaudio-dev libx11-dev libxv-dev libxt-dev libxfixes-dev libvpx-dev \
    git nasm openssl-dev \
    #
    # build gstreamer
    && cd /tmp \
    && wget "https://gstreamer.freedesktop.org/src/gstreamer/gstreamer-$GST_VERSION.tar.xz" \
    && tar xvfJ "gstreamer-$GST_VERSION.tar.xz" > /dev/null \
    && cd "gstreamer-$GST_VERSION" \
    && ./configure --prefix=/usr \
    && make && make install \
    && cd /tmp && rm -rf "gstreamer-$GST_VERSION" \
    #
    # build gst-plugins-base
    && wget "https://gstreamer.freedesktop.org/src/gst-plugins-base/gst-plugins-base-$GST_VERSION.tar.xz" \
    && tar xvfJ "gst-plugins-base-$GST_VERSION.tar.xz" > /dev/null \
    && cd "gst-plugins-base-$GST_VERSION" \
    && ./configure --prefix=/usr \
    && make && make install \
    && cd /tmp && rm -rf "gst-plugins-base-$GST_VERSION" \
    #
    # build gst-plugins-good
    && wget "https://gstreamer.freedesktop.org/src/gst-plugins-good/gst-plugins-good-$GST_VERSION.tar.xz" \
    && tar xvfJ "gst-plugins-good-$GST_VERSION.tar.xz" > /dev/null \
    && cd "gst-plugins-good-$GST_VERSION" \
    && ./configure --prefix=/usr \
    && make && make install \
    && cd /tmp && rm -rf "gst-plugins-good-$GST_VERSION" \
    #
    # build openh264
    && git clone https://github.com/cisco/openh264.git \
    && cd openh264 \
    && make && make install \
    && cd /tmp && rm -rf openh264 \
    #
    # build gst-plugins-bad 
    && wget "https://gstreamer.freedesktop.org/src/gst-plugins-bad/gst-plugins-bad-$GST_VERSION.tar.xz" \
    && tar xvfJ "gst-plugins-bad-$GST_VERSION.tar.xz" > /dev/null \
    && cd "gst-plugins-bad-$GST_VERSION" \
    && ./configure --prefix=/usr \
    && make && make install \
    && cd /tmp && rm -rf "gst-plugins-bad-$GST_VERSION" \
    #
    # remove build deps
    && apk del .gst-build-deps

ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
# RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
# RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

# install neko dependencies
RUN apk add --no-cache supervisor openbox dbus-x11 xvfb pulseaudio alsa-plugins-pulse opus libvpx libxv libxtst libxfixes xclip ttf-freefont \
    && apk add --no-cache libevent --repository "http://dl-cdn.alpinelinux.org/alpine/edge/main" \
    && apk add --no-cache firefox-esr --repository "http://dl-cdn.alpinelinux.org/alpine/edge/community" \
    #
    # create a non-root user
    && addgroup -g $USER_GID $USERNAME \
    && adduser -D -u $USER_UID -G $USERNAME -s /bin/ash -h /home/$USERNAME $USERNAME \
    && adduser $USERNAME audio \
    && adduser $USERNAME video \
    && adduser $USERNAME pulse \
    #
    # install uBlock
    && mkdir -p /usr/lib/firefox/distribution/extensions \
    && wget -O /usr/lib/firefox/distribution/extensions/uBlock0@raymondhill.net.xpi https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/addon-607454-latest.xpi \
    #
    # setup pulseaudio
    && mkdir -p /home/$USERNAME/.config/pulse/ \
    && echo "default-server=unix:/tmp/pulseaudio.socket" > /home/$USERNAME/.config/pulse/client.conf \
    && chown -R $USERNAME:$USERNAME /home/$USERNAME \
    #
    # workaround for an X11 problem: http://blog.tigerteufel.de/?p=476
    && mkdir /tmp/.X11-unix && chmod 1777 /tmp/.X11-unix && chown $USERNAME /tmp/.X11-unix/ \
    #
    # make directories for neko
    && mkdir -p /etc/neko /var/www

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
COPY .docker/files/supervisord.conf /etc/neko/supervisord.conf
COPY .docker/files/default.pa /etc/pulse/default.pa
COPY .docker/files/firefox/neko.js /usr/lib/firefox/mozilla.cfg
COPY .docker/files/firefox/autoconfig.js /usr/lib/firefox/defaults/pref/autoconfig.js
COPY .docker/files/firefox/policies.json /usr/lib/firefox/distribution/policies.json

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