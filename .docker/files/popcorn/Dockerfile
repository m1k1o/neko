FROM nurdism/neko:openbox

#
# install popcorn time
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends ; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY .docker/files/popcorn/supervisord.conf /etc/neko/supervisord/popcorn.conf
COPY .docker/files/popcorn/openbox.xml /etc/neko/openbox.xml