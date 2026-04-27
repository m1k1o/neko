ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

# install openbox + Ghostty (mkasberg/ghostty-ubuntu prebuilt deb for trixie/arm64)
ARG GHOSTTY_DEB_URL=https://github.com/mkasberg/ghostty-ubuntu/releases/download/1.3.1-0-ppa2/ghostty_1.3.1-0.ppa2_arm64_trixie.deb
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        openbox \
        # Ghostty runtime deps (GTK4 + libadwaita per https://ghostty.org/docs/linux)
        libgtk-4-1 libadwaita-1-0 \
        # font basics
        fonts-jetbrains-mono fonts-firacode \
        # sudo so the neko user can install more tools at runtime
        sudo \
        # baseline CLI toolkit so the terminal is useful out of the box
        tmux vim git htop less tree jq man-db \
        ca-certificates curl; \
    curl -fsSL "$GHOSTTY_DEB_URL" -o /tmp/ghostty.deb; \
    apt-get install -y --no-install-recommends /tmp/ghostty.deb; \
    rm -f /tmp/ghostty.deb; \
    # passwordless sudo for the neko user so `sudo apt install foo` just works
    echo 'neko ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/neko; \
    chmod 0440 /etc/sudoers.d/neko; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

COPY supervisord.conf /etc/neko/supervisord/ghostty.conf
COPY openbox.xml /etc/neko/openbox.xml
COPY --chown=neko:neko ghostty-config /home/neko/.config/ghostty/config
