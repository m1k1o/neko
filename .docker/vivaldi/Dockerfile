ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG VIVALDI_VERSION="5.3.2679.34-1"
# TODO: Get chromium version from vivaldi
ARG CHROMIUM_VERSION="102.0.5005.72"

#
# install vivaldi
SHELL ["/bin/bash", "-c"]
RUN set -eux; apt-get update; \
    wget -O /tmp/vivaldi.deb "https://downloads.vivaldi.com/stable/vivaldi-stable_${VIVALDI_VERSION}_amd64.deb"; \
    apt-get install -y --no-install-recommends wget unzip xz-utils jq openbox /tmp/vivaldi.deb; \
    /opt/vivaldi/update-ffmpeg; \
    #
    # install latest version of uBlock Origin and SponsorBlock for YouTube
    EXTENSIONS_DIR="/usr/share/chromium/extensions"; \
    EXTENSIONS=( \
      cjpalhdlnbpafiamejdnhcphjbkeiagm \
      mnjggcdmjocbbbhaepdhchncahnbgone \
    ); \
    mkdir -p "${EXTENSIONS_DIR}"; \
    for EXT_ID in "${EXTENSIONS[@]}"; \
    do \
      EXT_URL="https://clients2.google.com/service/update2/crx?response=redirect&nacl_arch=x86-64&prodversion=${CHROMIUM_VERSION}&acceptformat=crx2,crx3&x=id%3D${EXT_ID}%26installsource%3Dondemand%26uc"; \
      EXT_PATH="${EXTENSIONS_DIR}/${EXT_ID}.crx"; \
      wget -O "${EXT_PATH}" "${EXT_URL}"; \
      EXT_VERSION="$(unzip -p "${EXT_PATH}" manifest.json 2>/dev/null | jq -r ".version")"; \
      echo -e "{\n  \"external_crx\": \"${EXT_PATH}\",\n  \"external_version\": \"${EXT_VERSION}\"\n}" > "${EXTENSIONS_DIR}"/"${EXT_ID}".json; \
    done; \
    #
    # clean up
    apt-get --purge autoremove -y xz-utils jq; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/vivaldi-browser.conf
COPY --chown=neko preferences.json /home/neko/.config/vivaldi/Default/Preferences
COPY policies.json /etc/opt/vivaldi/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
