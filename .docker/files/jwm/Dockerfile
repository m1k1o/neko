FROM nurdism/neko:base

#
# install jwm
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends jwm;

#
# copy jwm conf and supervisord conf
COPY .docker/files/jwm/supervisord.conf /etc/neko/supervisord/jwm.conf
COPY .docker/files/jwm/conf.xml /etc/neko/jwm.xml
