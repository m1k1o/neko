FROM nurdism/neko:dev

# Use the "remoteUser" property in devcontainer.json to use it. On Linux, the container 
# user's GID/UIDs will be updated to match your local UID/GID (when using the dockerFile property).
# See https://aka.ms/vscode-remote/containers/non-root-user for details.
ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# Set to false to skip installing zsh and Oh My ZSH!
ARG INSTALL_ZSH="true"

#
# Location and expected SHA for common setup script - SHA generated on release
ARG COMMON_SCRIPT_SOURCE="https://raw.githubusercontent.com/microsoft/vscode-dev-containers/master/script-library/common-debian.sh"
ARG COMMON_SCRIPT_SHA="dev-mode"

#
# Docker Compose version
ARG COMPOSE_VERSION=1.24.0

#
# Verify git, common tools / libs installed, add/modify non-root user, optionally install zsh
RUN set -eux; \
    wget -q -O /tmp/common-setup.sh $COMMON_SCRIPT_SOURCE; \
    if [ "$COMMON_SCRIPT_SHA" != "dev-mode" ]; then echo "$COMMON_SCRIPT_SHA /tmp/common-setup.sh" | sha256sum -c - ; fi; \
    /bin/bash /tmp/common-setup.sh "$INSTALL_ZSH" "$USERNAME" "$USER_UID" "$USER_GID"; \
    rm /tmp/common-setup.sh; \
    #
    # Install docker
    apt-get install -y apt-transport-https gnupg-agent software-properties-common lsb-release; \
    curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | (OUT=$(apt-key add - 2>&1) || echo $OUT); \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable"; \
    apt-get update; apt-get install -y docker-ce-cli; \
    #
    # Install docker compose
    curl -sSL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose; \
    chmod +x /usr/local/bin/docker-compose; \
    #
    # Set alternate global install location that both users have rights to access
    mkdir -p /usr/local/share/npm-global; \
    chown ${USERNAME}:root /usr/local/share/npm-global; \
    npm config -g set prefix /usr/local/share/npm-global; \
    sudo -u ${USERNAME} npm config -g set prefix /usr/local/share/npm-global

ENV PATH=/usr/local/share/npm-global/bin:$PATH

#
# switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog
