FROM nurdism/neko:base

#
# install xfce4
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends xfce4;

#
# copy xfce4 conf and supervisord conf
COPY .docker/files/xfce4/supervisord.conf /etc/neko/openbox.xml
# COPY .docker/files/xfce4/xfconf /etc/neko/xfconf
