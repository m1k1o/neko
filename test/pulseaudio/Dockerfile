FROM debian:bullseye-slim

#
# set custom user
ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# install dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends pulseaudio; \
    #
    # create a non-root user
    groupadd --gid $USER_GID $USERNAME; \
    useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME; \
    #
    # make directories
    mkdir -p /home/$USERNAME/.config/pulse; \
    chown -R $USERNAME:$USERNAME /home/$USERNAME; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# set default envs
ENV USER=$USERNAME

#
# copy configuation files
COPY default.pa /etc/pulse/default.pa

USER $USERNAME

ENTRYPOINT [ "/usr/bin/pulseaudio" ]

CMD [ "--log-level=info", "--disallow-module-loading", "--disallow-exit", "--exit-idle-time=-1" ]
