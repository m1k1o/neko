ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

#
# install neko chromium
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends unzip chromium chromium-sandbox openbox; \
    #
    # install widevine module
    WIDEVINE_VERSION=$(wget --quiet -O - https://dl.google.com/widevine-cdm/versions.txt | tail -n 1); \
    wget -O /tmp/widevine.zip "https://dl.google.com/widevine-cdm/$WIDEVINE_VERSION-linux-x64.zip"; \
    unzip -p /tmp/widevine.zip libwidevinecdm.so > /usr/lib/chromium/libwidevinecdm.so; \
    chmod 644 /usr/lib/chromium/libwidevinecdm.so; \
    rm /tmp/widevine.zip; \
    #
    # clean up
    apt-get --purge autoremove -y unzip; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/chromium.conf
COPY --chown=neko preferences.json /home/neko/.config/chromium/Default/Preferences
COPY policies.json /etc/chromium/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
