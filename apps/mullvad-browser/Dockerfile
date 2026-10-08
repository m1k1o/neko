ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

#
# install dependencies
RUN set -eux; apt-get update; \
    #
    # download mullvad browser
    curl -fsSLo /usr/share/keyrings/mullvad-keyring.asc https://repository.mullvad.net/deb/mullvad-keyring.asc; \
    echo "deb [signed-by=/usr/share/keyrings/mullvad-keyring.asc arch=$( dpkg --print-architecture )] https://repository.mullvad.net/deb/stable stable main" | tee /etc/apt/sources.list.d/mullvad.list; \
    apt update; \
    apt install -y --no-install-recommends openbox mullvad-browser; \
    #
    # clean up
    apt-get clean -y; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*; \
    #
    # make default profile
    mkdir -p /home/neko/.mullvad-browser; \
    chown neko:neko /home/neko/.mullvad-browser; \
    runuser -u neko -- env HOME=/home/neko /usr/bin/mullvad-browser --verbose --headless --no-remote -CreateProfile "default /home/neko/.mullvad-browser/.mullvad/mullvadbrowser/profile.default";

#
# copy configuation file
COPY supervisord.conf /etc/neko/supervisord/mullvad-browser.conf
COPY neko.js /usr/lib/mullvad-browser/mozilla.cfg
COPY autoconfig.js /usr/lib/mullvad-browser/defaults/pref/autoconfig.js
COPY policies.json /usr/lib/mullvad-browser/distribution/policies.json
COPY openbox.xml /etc/neko/openbox.xml
