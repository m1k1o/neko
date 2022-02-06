ARG BASE_IMAGE=m1k1o/neko:arm-base
FROM $BASE_IMAGE

#
# install firefox-esr
RUN set -eux; apt-get update; \
    # TODO: Bring back DRM support with arm32v7/debian:buster-slim image.
    apt-get install -y --no-install-recommends openbox firefox-esr; \
    #
    # install extensions
    mkdir -p /usr/lib/firefox-esr/distribution/extensions; \
    wget -O '/usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi' https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi; \
    wget -O '/usr/lib/firefox-esr/distribution/extensions/sponsorBlocker@ajay.app.xpi' https://addons.mozilla.org/firefox/downloads/latest/sponsorblock/latest.xpi; \
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
COPY neko.js /usr/lib/firefox-esr/mozilla.cfg
COPY autoconfig.js /usr/lib/firefox-esr/defaults/pref/autoconfig.js
COPY policies.json /usr/lib/firefox-esr/distribution/policies.json
COPY --chown=neko profiles.ini /home/neko/.mozilla/firefox/profiles.ini
COPY openbox.xml /etc/neko/openbox.xml
