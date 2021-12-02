ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

#
# install vncviewer
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox xtightvncviewer; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

ENV NEKO_VNC_URL=""

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/vncviewer.conf
COPY openbox.xml /etc/neko/openbox.xml
