FROM nurdism/neko:deps

#
# neko files
COPY .docker/files/neko/supervisord.conf /etc/neko/supervisord/neko.conf
COPY client/dist/ /var/www
COPY server/bin/neko /usr/bin/neko

#
# neko env
ENV NEKO_PASSWORD=neko
ENV NEKO_PASSWORD_ADMIN=admin
ENV NEKO_BIND=:8080