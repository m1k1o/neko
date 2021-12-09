ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG API_URL="https://packages.microsoft.com/repos/edge/pool/main/m/microsoft-edge-stable/"

#
# install microsoft edge
RUN set -eux; apt-get update; \
    #
    # fetch latest release
    SRC_URL="${API_URL}$(wget -O - "${API_URL}" 2>/dev/null | sed -n 's/.*href="\([^"]*\).*/\1/p' | tail -1)"; \
    wget -O /tmp/microsoft-edge.deb "${SRC_URL}"; \
    apt-get install -y --no-install-recommends openbox /tmp/microsoft-edge.deb; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/microsoft-edge.conf
COPY --chown=neko preferences.json /home/neko/.config/microsoft-edge/Default/Preferences
COPY policies.json /etc/opt/edge/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
