FROM debian:stretch-slim

ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies ttf-freefont
RUN apt-get update \
    && apt-get -y install curl apt-utils supervisor openbox dbus-x11 xvfb pulseaudio firefox-esr \
    && apt-get -y install gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-pulseaudio \
    && apt-get -y install libxcb-xkb-dev libxkbcommon-x11-dev xclip \
    #
    # Create a non-root user
    && groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME \
    && adduser $USERNAME audio \
    && adduser $USERNAME video \
    && adduser $USERNAME pulse \
    #
    # Add sudo support for the non-root user
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    #
    # Install uBlock
    && mkdir -p /usr/lib/firefox-esr/distribution/extensions \
    && curl -o /usr/lib/firefox-esr/distribution/extensions/uBlock0@raymondhill.net.xpi https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/addon-607454-latest.xpi \
    #
    # Setup Pulse Audio
    && mkdir -p /home/$USERNAME/.config/pulse/ \
    && echo "default-server=unix:/tmp/pulseaudio.socket" > /home/$USERNAME/.config/pulse/client.conf \
    && chown -R $USERNAME:$USERNAME /home/$USERNAME \
    #
    # Workaround for an X11 problem. See: http://blog.tigerteufel.de/?p=476
    && mkdir /tmp/.X11-unix && chmod 1777 /tmp/.X11-unix && chown $USERNAME /tmp/.X11-unix/ \
    #
    # Make directories for neko
    && mkdir -p /etc/neko /var/www \
    #
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# Copy configuation files
COPY .docker/openbox.xml /etc/neko/openbox.xml
COPY .docker/supervisord.conf /etc/neko/supervisord.conf
COPY .docker/default.pa /etc/pulse/default.pa
COPY .docker/firefox/neko.js /usr/lib/firefox-esr/mozilla.cfg
COPY .docker/firefox/autoconfig.js /usr/lib/firefox-esr/defaults/pref/autoconfig.js
COPY .docker/firefox/policies.json /usr/lib/firefox-esr/distribution/policies.json

#
# Neko files
COPY client/dist/ /var/www
COPY server/bin/neko /usr/bin/neko

USER $USERNAME

#
# Neko Env
ENV NEKO_USER=$USERNAME
ENV NEKO_DISPLAY=0
ENV NEKO_WIDTH=1280
ENV NEKO_HEIGHT=720
ENV NEKO_PASSWORD=neko
ENV NEKO_ADMIN=admin
ENV NEKO_BIND=:8080

# Neko Nyan (=^-ω-^=)
ENV NEKO_URL=https://www.youtube.com/embed/QH2-TGUlwu4 

#
# Copy entrypoint
COPY .docker/entrypoint.sh /entrypoint.sh

#
# Run Service
CMD ["/bin/bash", "/entrypoint.sh"]