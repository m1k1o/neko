FROM debian:bullseye-slim

#
# set custom user
ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# install dependencies
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends openbox; \
    #
    # create a non-root user
    groupadd --gid $USER_GID $USERNAME; \
    useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# set default envs
ENV USER=$USERNAME

#
# copy configuation files
COPY openbox.xml /etc/neko/openbox.xml

USER $USERNAME

ENTRYPOINT [ "/usr/bin/openbox" ]

CMD [ "--config-file", "/etc/neko/openbox.xml" ]
