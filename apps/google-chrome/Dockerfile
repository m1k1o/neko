ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

# Set this to a complete .deb URL to pin a specific Chrome version. By default,
# download the package matching the image architecture.
ARG SRC_URL=""

#
# install google chrome
RUN set -eux; \
    ARCH="$(dpkg --print-architecture)"; \
    case "${ARCH}" in \
        amd64|arm64) ;; \
        *) echo "Google Chrome is not available for ${ARCH}" >&2; exit 1 ;; \
    esac; \
    SRC_URL="${SRC_URL:-https://dl.google.com/linux/direct/google-chrome-stable_current_${ARCH}.deb}"; \
    apt-get update; \
    wget -O /tmp/google-chrome.deb "${SRC_URL}"; \
    apt-get install -y --no-install-recommends openbox /tmp/google-chrome.deb; \
    # Install the ARM64 CDM prepared by the runtime where Chrome discovers it.
    if [ "${ARCH}" = "arm64" ]; then \
        mkdir -p /opt/google/chrome/WidevineCdm; \
        find /var/lib/widevine/WidevineCdm/ ! -name "WidevineCdm" -exec cp -RL {} /opt/google/chrome/WidevineCdm/ \; ; \
    fi; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/google-chrome.conf
COPY --chown=neko preferences.json /home/neko/.config/google-chrome/Default/Preferences
COPY policies.json /etc/opt/chrome/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
