FROM nurdism/neko:openbox

#
# install dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends curl xz-utils file libgtk-3-0 libdbus-glib-1-2; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

WORKDIR /home/neko
USER neko

#
# download TOR browser
RUN DOWNLOAD_URI="$(curl -s -N https://www.torproject.org/download/ | grep -Po -m 1 '(?=(dist/torbrowser)).*(?<=.tar.xz)')"; \
	echo "Downloading $DOWNLOAD_URI"; \
	curl -sSL -o tor.tar.xz "https://www.torproject.org/$DOWNLOAD_URI"; \
    tar -xvJf tor.tar.xz; \
    rm -f tor.tar.xz*;

USER root

#
# copy configuation file
COPY .docker/files/tor-browser/supervisord.conf /etc/neko/supervisord/tor-browser.conf
COPY .docker/files/tor-browser/openbox.xml /etc/neko/openbox.xml
