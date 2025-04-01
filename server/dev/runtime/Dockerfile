ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

ARG SRC_URL="https://download.mozilla.org/?product=firefox-latest&os=linux64&lang=en-US"

#
# install xfce and firefox
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends \
        dbus-x11 xfce4 xfce4-terminal sudo \
        xz-utils libgtk-3-0 libdbus-glib-1-2; \
    #
    # fetch latest firefox release
    wget -O /tmp/firefox-setup.tar.xz "${SRC_URL}"; \
    mkdir /usr/lib/firefox; \
    tar -xvf /tmp/firefox-setup.tar.xz -C /usr/lib; \
    rm -f /tmp/firefox-setup.tar.xz; \
    ln -s /usr/lib/firefox/firefox /usr/bin/firefox; \
    #
    # add user to sudoers
    usermod -aG sudo neko; \
    echo "neko:neko" | chpasswd; \
    echo "%sudo ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers; \
    # clean up
    apt-get --purge autoremove -y xz-utils; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/xfce.conf
