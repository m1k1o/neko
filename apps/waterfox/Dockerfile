ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

ARG SRC_URL="https://cdn1.waterfox.net/waterfox/releases/latest/linux"

#
# install waterfox
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox \
    xz-utils bzip2 libgtk-3-0 libdbus-glib-1-2; \
    #
    # fetch latest release
    wget --user-agent="Mozilla/5.0" -O /tmp/waterfox-setup.tar.bz2 "${SRC_URL}"; \
    mkdir /usr/lib/waterfox; \
    tar -xjf /tmp/waterfox-setup.tar.bz2 -C /usr/lib; \
    rm -f /tmp/waterfox-setup.tar.bz2; \
    ln -s /usr/lib/waterfox/waterfox /usr/bin/waterfox; \
    #
    # create a profile directory
    mkdir -p /home/neko/.waterfox/profile.default/extensions; \
    chown -R neko:neko /home/neko/.waterfox/profile.default; \
    #
    # clean up
    apt-get --purge autoremove -y xz-utils bzip2; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/waterfox.conf
COPY neko.js /usr/lib/waterfox/mozilla.cfg
COPY autoconfig.js /usr/lib/waterfox/defaults/pref/autoconfig.js
COPY policies.json /usr/lib/waterfox/distribution/policies.json
COPY --chown=neko profiles.ini /home/neko/.waterfox/profiles.ini
COPY openbox.xml /etc/neko/openbox.xml
