FROM m1k1o/neko:base

#
# install firefox
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends firefox-esr; \
    #
    # create a non-root user
    #groupadd --gid 1000 neko; \
    #useradd --uid 1000 --gid neko --shell /bin/bash --create-home neko; \
    #
    # install fonts
    apt-get install -y --no-install-recommends \
        # Google emojis
        fonts-noto-color-emoji \
        # Japanese fonts
        fonts-takao-mincho \
        # Chinese fonts
        fonts-wqy-zenhei xfonts-intl-chinese xfonts-wqy \
        # Korean fonts
        fonts-wqy-microhei; \
    #
    # clean up
    apt-get --purge autoremove -y xz-utils bzip2; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

USER neko

ENTRYPOINT [ "/usr/bin/firefox" ]

CMD [ "--display", $DISPLAY, "-setDefaultBrowser", "-width", "1280", "-height", "720" ]
