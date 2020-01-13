FROM buildpack-deps:stretch

ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update \
    && apt-get -y install curl supervisor openbox dbus-x11 ttf-freefont xvfb pulseaudio consolekit firefox-esr \
    && apt-get -y install gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-pulseaudio libxcb-xkb-dev libxkbcommon-x11-dev \
    #
    # Create a non-root user
    && groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME \
    && adduser $USERNAME audio \
    && adduser $USERNAME video \
    && adduser $USERNAME pulse \
    #
    # Install uBlock
    && mkdir -p /usr/lib/firefox-esr/distribution/extensions \
    && curl -o /usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/addon-607454-latest.xpi \
    #
    # Make directories for neko
    && mkdir -p /etc/neko /var/www \
    #
    # Setup Pulse Audio
    mkdir -p /home/$USERNAME/.config/pulse/ \
    && echo "default-server=unix:/tmp/pulseaudio.socket" > /home/$USERNAME/.config/pulse/client.conf \
    && chown -R $USERNAME:$USERNAME /home/$USERNAME \
    #
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# Copy configuation files
COPY .docker/pulseaudio.pa /etc/neko/pulseaudio.pa
COPY .docker/openbox.xml /etc/neko/openbox.xml
COPY .docker/supervisord.conf /etc/neko/supervisord.conf
COPY .docker/policies.json /usr/lib/firefox-esr/distribution/policies.json

#
# Neko files
COPY client/dist/ /var/www
COPY server/bin/neko /usr/bin/neko

#
# Neko Env
ENV NEKO_USER=$USERNAME
ENV NEKO_DISPLAY=0
ENV NEKO_WIDTH=1280
ENV NEKO_HEIGHT=720
ENV NEKO_URL=https://www.youtube.com/embed/QH2-TGUlwu4
ENV NEKO_PASSWORD=neko
ENV NEKO_BIND=0.0.0.0:80
ENV NEKO_KEY=
ENV NEKO_CERT=

#
# Copy entrypoint
COPY .docker/entrypoint.sh /entrypoint.sh

#
# Run entrypoint
CMD ["/bin/bash", "/entrypoint.sh"]