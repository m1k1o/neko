ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG SRC_URL="https://download.mozilla.org/?product=firefox-latest&os=linux64&lang=en-US"

#
# install firefox
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox \
    xz-utils bzip2 libgtk-3-0 libdbus-glib-1-2; \
    #
    # fetch latest release
    wget -O /tmp/firefox-setup.tar.bz2 "${SRC_URL}"; \
    mkdir /usr/lib/firefox; \
    tar -xjf /tmp/firefox-setup.tar.bz2 -C /usr/lib; \
    rm -f /tmp/firefox-setup.tar.bz2; \
    ln -s /usr/lib/firefox/firefox /usr/bin/firefox; \
    #
    # clean up
    apt-get --purge autoremove -y xz-utils bzip2; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/firefox.conf
COPY neko.js /usr/lib/firefox/mozilla.cfg
COPY autoconfig.js /usr/lib/firefox/defaults/pref/autoconfig.js
COPY policies.json /usr/lib/firefox/distribution/policies.json
COPY openbox.xml /etc/neko/openbox.xml
