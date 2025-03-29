ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

#
# install dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox curl xz-utils file libgtk-3-0 libdbus-glib-1-2; \
    #
    # download TOR browser
    DOWNLOAD_URI="$(curl -s -N https://www.torproject.org/download/ | grep -Po -m 1 '(?=(dist/torbrowser)).*(?<=.tar.xz)')"; \
	echo "Downloading $DOWNLOAD_URI"; \
	curl -sSL -o /tmp/tor.tar.xz "https://www.torproject.org/$DOWNLOAD_URI"; \
    tar -xvJf /tmp/tor.tar.xz -C /opt; \
    mv /opt/tor-browser* /opt/tor-browser_en-US; \
    chown -R neko:neko /opt/tor-browser_en-US/; \
    rm -f /tmp/tor.tar.xz; \
    #
    # clean up
    apt-get --purge autoremove -y curl xz-utils; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*;

#
# copy configuation file
COPY supervisord.conf /etc/neko/supervisord/tor-browser.conf
COPY openbox.xml /etc/neko/openbox.xml
