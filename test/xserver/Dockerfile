FROM debian:bullseye-slim

#
# install dependencies
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends x11-xserver-utils xserver-xorg-video-dummy; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY xorg.conf /etc/neko/xorg.conf

ENTRYPOINT [ "/usr/bin/X", "-config", "/etc/neko/xorg.conf", "-nolisten", "local", "-logfile", "/dev/stderr" ]
