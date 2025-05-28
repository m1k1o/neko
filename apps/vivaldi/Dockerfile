ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

#
# install vivaldi
SHELL ["/bin/bash", "-c"]
RUN set -eux; apt-get update; \
    ARCH=$(dpkg --print-architecture); \
    wget -O /tmp/vivaldi.deb "https://downloads.vivaldi.com/stable/vivaldi-stable_${ARCH}.deb"; \
    apt-get install -y --no-install-recommends wget unzip xz-utils jq openbox; \
    apt install -y --no-install-recommends /tmp/vivaldi.deb; \
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
      EXT_URL="https://clients2.google.com/service/update2/crx?response=redirect&prodversion=100&acceptformat=crx2,crx3&x=id%3D${EXT_ID}%26installsource%3Dondemand%26uc"; \
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
