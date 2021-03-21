FROM m1k1o/neko:base

#
# install firefox-esr
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox; \
    apt-get install -y wesnoth; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/wesnoth.conf
COPY openbox.xml /etc/neko/openbox.xml
