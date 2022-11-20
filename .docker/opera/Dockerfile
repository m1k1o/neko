ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG API_URL="https://download5.operacdn.com/pub/opera/desktop/"
ARG LIBFFMPEG_API_URL="https://api.github.com/repos/nwjs-ffmpeg-prebuilt/nwjs-ffmpeg-prebuilt/releases/latest"

#
# install opera
RUN set -eux; apt-get update; \
    #
    # fetch latest release
    VERSION="$(wget -O - "${API_URL}" 2>/dev/null | sed -n 's/.*href="\([^"/]*\).*/\1/p' | tail -1)"; \
    wget -O /tmp/opera.deb "${API_URL}${VERSION}/linux/opera-stable_${VERSION}_amd64.deb"; \
    apt-get install -y --no-install-recommends openbox jq unzip /tmp/opera.deb; \
    #
    # install libffmpeg
    LIBFFMPEG_URL="$(wget -O - "${LIBFFMPEG_API_URL}" 2>/dev/null | jq -r "[.assets[] | select(.browser_download_url | contains(\"linux-x64.zip\"))][-1] | .browser_download_url")"; \
    wget -O /tmp/libffmpeg.zip $LIBFFMPEG_URL; \
    unzip -o /tmp/libffmpeg.zip libffmpeg.so -d /usr/lib/x86_64-linux-gnu/opera/lib_extra; \
    echo '[]' > /usr/lib/x86_64-linux-gnu/opera/resources/ffmpeg_preload_config.json; \
    #
    # clean up
    apt-get --purge autoremove -y unzip jq; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/opera.conf
COPY openbox.xml /etc/neko/openbox.xml

