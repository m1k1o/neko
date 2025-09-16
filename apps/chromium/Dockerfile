ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

COPY ./widevinecdm.sh /widevine.sh

#
# install neko chromium
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends chromium chromium-common chromium-sandbox openbox; \
    #
    # install widevine module (only for amd64)
    CHROMIUM_DIR="/usr/lib/chromium"; \
    ARCH=$(dpkg --print-architecture); \
    if [ "${ARCH}" = "amd64" ]; then \
        apt-get install -y --no-install-recommends xz-utils; \
        ./widevine.sh "${CHROMIUM_DIR}/WidevineCdm"; \
        rm -f /widevine.sh; \
        apt-get --purge autoremove -y xz-utils; \
    else \
        echo "Widevine is not supported on ${ARCH}"; \
    fi; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/chromium.conf
COPY --chown=neko preferences.json /home/neko/.config/chromium/Default/Preferences
COPY policies.json /etc/chromium/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
