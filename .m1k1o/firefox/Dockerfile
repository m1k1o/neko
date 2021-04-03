ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

#
# install firefox-esr
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox firefox-esr; \
    #
    # install extensions
    mkdir -p /usr/lib/firefox-esr/distribution/extensions; \
    wget -O '/usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi' https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi; \
    wget -O /usr/lib/firefox-esr/distribution/extensions/nordvpnproxy@nordvpn.com.xpi https://addons.mozilla.org/firefox/downloads/latest/nordvpn-proxy-extension/latest.xpi; \
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
COPY openbox.xml /etc/neko/openbox.xml
