ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

RUN set -eux; apt-get update; \
    apt-get install -y --no-install-recommends apt-transport-https curl openbox; \
    #
    # install brave browser
    curl -fsSLo /usr/share/keyrings/brave-browser-archive-keyring.gpg https://brave-browser-apt-release.s3.brave.com/brave-browser-archive-keyring.gpg; \
    echo "deb [signed-by=/usr/share/keyrings/brave-browser-archive-keyring.gpg arch=amd64] https://brave-browser-apt-release.s3.brave.com/ stable main" \
        | tee /etc/apt/sources.list.d/brave-browser-release.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends brave-browser; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/brave.conf
COPY --chown=neko preferences.json /home/neko/.config/brave/Default/Preferences
COPY policies.json /etc/brave/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
