FROM nurdism/neko:base

#
# install openbox
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox;

#
# copy openbox conf and supervisord conf
COPY .docker/files/openbox/supervisord.conf /etc/neko/supervisord/openbox.conf
COPY .docker/files/openbox/conf.xml /etc/neko/openbox.xml
