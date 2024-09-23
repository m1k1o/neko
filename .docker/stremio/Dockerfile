ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG STREMIO_VERSION=4.4.168
ARG NODE_VERSION=18

# 
# install packages from bookworm
RUN set -eux; echo "deb http://deb.debian.org/debian bookworm main contrib non-free" > /etc/apt/sources.list.d/bookworm.list; \
    apt-get update; apt-get install -y --no-install-recommends libfdk-aac2 xdg-utils; \
    #
    # clean up
    rm /etc/apt/sources.list.d/bookworm.list; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*;
# 
# install stremio
RUN set -eux; apt-get update; apt-get install -y --no-install-recommends ca-certificates curl openbox wget libmpv1 \
    qml-module-qt-labs-platform qml-module-qtquick-controls qml-module-qtquick-dialogs qml-module-qtwebchannel \
    qml-module-qtwebengine qml-module-qt-labs-folderlistmodel qml-module-qt-labs-settings; \
    curl -sL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash -; \
    apt-get install -y --no-install-recommends nodejs; \
    mkdir /usr/share/desktop-directories/; mkdir -p /usr/share/icons/hicolor/ \
    #
    # fix xdg-icon-resource
    echo '#!/bin/sh\nexit 0' > /usr/bin/xdg-icon-resource; chmod +x /usr/bin/xdg-icon-resource; \
    #
    # download and install stremio
    wget -qO /tmp/stremio.deb https://dl.strem.io/shell-linux/v${STREMIO_VERSION}/stremio_${STREMIO_VERSION}-1_amd64.deb && dpkg -i /tmp/stremio.deb; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*;

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/stremio.conf
COPY openbox.xml /etc/neko/openbox.xml
