FROM nurdism/neko:deps

#
# avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# install neko dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends wget ca-certificates; \
    #
    # create a non-root user
    groupadd --gid $USER_GID $USERNAME; \
    useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME; \
    adduser $USERNAME audio; \
    adduser $USERNAME video; \
    adduser $USERNAME pulse; \
    #
    # setup pulseaudio
    mkdir -p /home/$USERNAME/.config/pulse/; \
    echo "default-server=unix:/tmp/pulseaudio.socket" > /home/$USERNAME/.config/pulse/client.conf; \
    #
    # workaround for an X11 problem: http://blog.tigerteufel.de/?p=476
    mkdir /tmp/.X11-unix; chmod 1777 /tmp/.X11-unix; chown $USERNAME /tmp/.X11-unix/; \
    #
    # make directories for neko
    mkdir -p /etc/neko /var/www /var/log/neko; chmod 1777 /var/log/neko; chown $USERNAME /var/log/neko/;  \
    chown -R $USERNAME:$USERNAME /home/$USERNAME; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# env
ENV USER=$USERNAME
ENV DISPLAY=:99.0
ENV NEKO_PASSWORD=neko
ENV NEKO_PASSWORD_ADMIN=admin
ENV NEKO_BIND=:8080

#
# neko config
COPY .docker/files/dbus /usr/bin/dbus
COPY .docker/files/openbox.xml /etc/neko/openbox.xml
COPY .docker/files/supervisord.conf /etc/neko/supervisord.conf
COPY .docker/files/xorg.conf /etc/neko/xorg.conf
COPY .docker/files/default.pa /etc/pulse/default.pa
COPY .docker/files/neko/supervisord.conf /etc/neko/supervisord/neko.conf

#
# neko dist
COPY client/dist/ /var/www
COPY server/bin/neko /usr/bin/neko

#
# run neko
CMD ["/usr/bin/supervisord", "-c", "/etc/neko/supervisord.conf"]