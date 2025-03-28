ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

#
# install kde
RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends kde-full kwin-x11 sudo; \
    #
    # add user to sudoers
    usermod -aG sudo neko; \
    echo "neko:neko" | chpasswd; \
    echo "%sudo ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers; \
    # clean up
    apt remove xserver-xorg-legacy -y; \
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# disable autolock
RUN kwriteconfig5 --file /home/neko/.config/kscreenlockerrc --group Daemon --key Autolock false; \
    chown neko:neko /home/neko/.config/kscreenlockerrc

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/kde.conf

