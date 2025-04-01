ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

#
# install neko chromium
RUN set -eux; \
    echo "deb http://ftp.de.debian.org/debian bookworm main" >> /etc/apt/sources.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends chromium chromium-common chromium-sandbox openbox; \
    #
    # install widevine module (only for amd64)
    CHROMIUM_DIR="/usr/lib/chromium"; \
    ARCH=$(dpkg --print-architecture); \
    if [ "${ARCH}" = "amd64" ]; then \
        # https://commondatastorage.googleapis.com/chromeos-localmirror/distfiles/chromeos-lacros-arm64-squash-zstd-120.0.6098.0
        apt-get install -y --no-install-recommends unzip; \
        WIDEVINE_ARCH="x64"; \
        WIDEVINE_VERSION=$(wget --quiet -O - https://dl.google.com/widevine-cdm/versions.txt | sort --version-sort | tail -n 1); \
        wget -O /tmp/widevine.zip "https://dl.google.com/widevine-cdm/${WIDEVINE_VERSION}-linux-${WIDEVINE_ARCH}.zip"; \
        mkdir -p "${CHROMIUM_DIR}/WidevineCdm/_platform_specific/linux_${WIDEVINE_ARCH}"; \
        unzip -p /tmp/widevine.zip LICENSE.txt > "${CHROMIUM_DIR}/WidevineCdm/LICENSE"; \
        unzip -p /tmp/widevine.zip manifest.json > "${CHROMIUM_DIR}/WidevineCdm/manifest.json"; \
        unzip -p /tmp/widevine.zip libwidevinecdm.so > "${CHROMIUM_DIR}/WidevineCdm/_platform_specific/linux_${WIDEVINE_ARCH}/libwidevinecdm.so"; \
        find "${CHROMIUM_DIR}/WidevineCdm" -type d -exec chmod 0755 '{}' \;; \
        find "${CHROMIUM_DIR}/WidevineCdm" -type f -exec chmod 0644 '{}' \;; \
        rm /tmp/widevine.zip; \
        apt-get --purge autoremove -y unzip; \
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
