ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

#
# install firefox
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox; \
    ARCH=$(dpkg --print-architecture); \
    if [ "${ARCH}" = "armhf" ]; then \
        #
        # install firefox-esr for armhf
        apt-get install -y --no-install-recommends firefox-esr; \
        ln -s /usr/lib/firefox-esr /usr/lib/firefox; \
        #
        # install extensions
        mkdir -p /usr/lib/firefox-esr/distribution/extensions; \
        wget -O '/usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi' https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi; \
        wget -O '/usr/lib/firefox-esr/distribution/extensions/sponsorBlocker@ajay.app.xpi' https://addons.mozilla.org/firefox/downloads/latest/sponsorblock/latest.xpi; \
    else \
        #
        # fetch latest release (for amd64 and arm64)
        if [ "${ARCH}" = "amd64" ]; then \
            SRC_URL="https://download.mozilla.org/?product=firefox-latest&os=linux64&lang=en-US"; \
        elif [ "${ARCH}" = "arm64" ]; then \
            SRC_URL="https://download.mozilla.org/?product=firefox-latest&os=linux64-aarch64&lang=en-US"; \
        fi; \
        if [ ! -z "${SRC_URL}" ]; then \
            apt-get install -y --no-install-recommends xz-utils libgtk-3-0 libdbus-glib-1-2; \
            wget -O /tmp/firefox-setup.tar.xz "${SRC_URL}"; \
            mkdir /usr/lib/firefox; \
            tar -xvf /tmp/firefox-setup.tar.xz -C /usr/lib; \
            rm -f /tmp/firefox-setup.tar.xz; \
            ln -s /usr/lib/firefox/firefox /usr/bin/firefox; \
            apt-get --purge autoremove -y xz-utils; \
        else \
            echo "Unsupported architecture: ${ARCH}"; \
            exit 1; \
        fi; \
    fi; \
    #
    # create a profile directory
    mkdir -p /home/neko/.mozilla/firefox/profile.default/extensions; \
    chown -R neko:neko /home/neko/.mozilla/firefox/profile.default; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/firefox.conf
COPY neko.js /usr/lib/firefox/mozilla.cfg
COPY autoconfig.js /usr/lib/firefox/defaults/pref/autoconfig.js
COPY policies.json /usr/lib/firefox/distribution/policies.json
COPY --chown=neko profiles.ini /home/neko/.mozilla/firefox/profiles.ini
COPY openbox.xml /etc/neko/openbox.xml
