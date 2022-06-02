ARG BASE_IMAGE=xeddius/neko:base
FROM $BASE_IMAGE

ARG SRC_URL="https://downloads.vivaldi.com/stable/vivaldi-stable_5.3.2679.34-1_amd64.deb"

#
# install vivaldi
RUN set -eux; apt-get update; \
    wget -O /tmp/vivaldi.deb "${SRC_URL}"; \
    apt-get install -y --no-install-recommends xz-utils openbox /tmp/vivaldi.deb; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/vivaldi-browser.conf
COPY --chown=neko preferences.json /home/neko/.config/vivaldi/Default/Preferences
COPY policies.json /etc/opt/vivaldi/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
COPY Default.tar.gz /tmp/
RUN tar -xzf /tmp/Default.tar.gz --directory /home/neko/.config/vivaldi/ && rm /tmp/Default.tar.gz
RUN /opt/vivaldi/update-ffmpeg
