ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

ARG API_URL="https://download5.operacdn.com/pub/opera/desktop/"

COPY fix-ffmpeg-widevine.sh /tmp/fix-ffmpeg-widevine.sh

#
# install opera
RUN set -eux; apt-get update; \
    #
    # fetch latest available release with linux package
    VERSIONS="$(wget -O - "${API_URL}" 2>/dev/null | sed -n 's/.*href="\([^"/]*\).*/\1/p' | sort --version-sort -r | head -n 10)"; \
    VERSION=""; \
    for v in $VERSIONS; do \
        if wget --spider "${API_URL}${v}/linux/opera-stable_${v}_amd64.deb" 2>/dev/null; then \
            VERSION="$v"; \
            break; \
        fi; \
    done; \
    if [ -z "$VERSION" ]; then \
        echo "No Opera version with Linux package found"; \
        exit 1; \
    fi; \
    echo "Using Opera version: $VERSION"; \
    wget -O /tmp/opera.deb "${API_URL}${VERSION}/linux/opera-stable_${VERSION}_amd64.deb"; \
    apt-get install -y --no-install-recommends openbox curl jq unzip /tmp/opera.deb; \
    /tmp/fix-ffmpeg-widevine.sh; \
    rm -f /tmp/opera.deb /tmp/fix-ffmpeg-widevine.sh; \
    #
    # clean up
    apt-get --purge autoremove -y curl jq unzip; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/opera.conf
COPY openbox.xml /etc/neko/openbox.xml

