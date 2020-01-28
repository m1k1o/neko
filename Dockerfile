FROM debian:stretch-slim

# avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

ENV GSTPATH /gst
ENV GST_VERSION 1.16

# build gstreamer
RUN set -eux && \
    apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates build-essential perl python pkg-config autoconf automake autopoint libtool bison flex \
    gettext nasm openssl libglib2.0-dev libopus-dev libvpx-dev libpulse-dev libx11-dev libxv-dev libxt-dev \
    libxtst-dev libxfixes-dev libssl-dev \
    ## set up dir
    && mkdir $GSTPATH \
    #
    # build openh264
    && cd $GSTPATH \
    && git clone https://github.com/cisco/openh264.git \
    && cd openh264 \
    && make && make install \
    && cd $GSTPATH && rm -rf openh264 \
    #
    # build gstreamer
    && for MODULE in \
      gstreamer \
      gst-plugins-base \
      gst-plugins-good \
      gst-plugins-bad \
    ; do \
        git clone git://anongit.freedesktop.org/gstreamer/$MODULE; \
        cd $MODULE; \
        git checkout $GST_VERSION; \
        PATH=$GSTPATH/local/bin:$PATH PKG_CONFIG_PATH=$GSTPATH/local/lib/pkgconfig ./autogen.sh --prefix $GSTPATH/local --disable-gtk-doc; \
        make && make install; \
        cd $GSTPATH && rm -rf $MODULE; \
    done \
    #
    # remove build deps
    && apt-get --purge autoremove -y build-essential perl python pkg-config autoconf automake autopoint libtool bison flex \
    gettext nasm openssl libglib2.0-dev libopus-dev libvpx-dev libpulse-dev libx11-dev libxv-dev libxt-dev \
    libxtst-dev libxfixes-dev libssl-dev

ENV PATH=$GSTPATH/local/bin:$PATH
ENV LD_LIBRARY_PATH=$GSTPATH/local/lib:$LD_LIBRARY_PATH
ENV PKG_CONFIG_PATH=$GSTPATH/local/lib/pkgconfig:$PKG_CONFIG_PATH

ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# install neko dependencies
RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends wget ca-certificates pulseaudio openbox dbus-x11 xvfb libxv1 xclip firefox-esr supervisor \
    #
    # create a non-root user
    && groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME \
    && adduser $USERNAME audio \
    && adduser $USERNAME video \
    && adduser $USERNAME pulse \
    #
    # install extensions
    && mkdir -p /usr/lib/firefox-esr/distribution/extensions \
    && wget -O /usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi \
    && wget -O /usr/lib/firefox-esr/distribution/extensions/nordvpnproxy@nordvpn.com.xpi https://addons.mozilla.org/firefox/downloads/latest/nordvpn-proxy-extension/latest.xpi \
    #
    # setup pulseaudio
    && mkdir -p /home/$USERNAME/.config/pulse/ \
    && echo "default-server=unix:/tmp/pulseaudio.socket" > /home/$USERNAME/.config/pulse/client.conf \
    #
    # workaround for an X11 problem: http://blog.tigerteufel.de/?p=476
    && mkdir /tmp/.X11-unix && chmod 1777 /tmp/.X11-unix && chown $USERNAME /tmp/.X11-unix/ \
    #
    # make directories for neko
    && mkdir -p /etc/neko /var/www /home/$USERNAME/.neko/logs \
    && chown -R $USERNAME:$USERNAME /home/$USERNAME \
    #
    # clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/*

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
COPY .docker/files/firefox/neko.js /usr/lib/firefox-esr/mozilla.cfg
COPY .docker/files/firefox/autoconfig.js /usr/lib/firefox-esr/defaults/pref/autoconfig.js
COPY .docker/files/firefox/policies.json /usr/lib/firefox-esr/distribution/policies.json

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