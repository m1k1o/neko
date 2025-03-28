ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

# install remmina
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends \
      remmina-plugin-rdp remmina-plugin-vnc \
      # remmina-plugin-x2go # not in bullseye
      remmina-plugin-spice remmina-plugin-nx; \
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/remmina.conf
COPY --chown=neko remmina.pref /home/neko/.config/remmina/remmina.pref
COPY --chown=neko rdp.remmina spice.remmina vnc.remmina /home/neko/.local/share/remmina/
COPY run-remmina.sh /usr/bin/run-remmina.sh
ENV REMMINA_URL=
ENV REMMINA_PROFILE=
