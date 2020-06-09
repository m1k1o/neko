FROM nurdism/neko:openbox

#
# install firefox-esr
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends firefox-esr; \
    #
    # install extensions
    mkdir -p /usr/lib/firefox-esr/distribution/extensions; \
    wget -O /usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi; \
    wget -O /usr/lib/firefox-esr/distribution/extensions/nordvpnproxy@nordvpn.com.xpi https://addons.mozilla.org/firefox/downloads/latest/nordvpn-proxy-extension/latest.xpi; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY .docker/files/firefox/supervisord.conf /etc/neko/supervisord/firefox.conf
COPY .docker/files/firefox/neko.js /usr/lib/firefox-esr/mozilla.cfg
COPY .docker/files/firefox/autoconfig.js /usr/lib/firefox-esr/defaults/pref/autoconfig.js
COPY .docker/files/firefox/policies.json /usr/lib/firefox-esr/distribution/policies.json
COPY .docker/files/firefox/openbox.xml /etc/neko/openbox.xml
