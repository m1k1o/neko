ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG SRC_URL="https://download.opera.com/download/get/?id=58545&location=415&nothanks=yes&sub=marine&utm_tryagain=yes"

ARG WV_VER_URL="https://dl.google.com/widevine-cdm/versions.txt"
ARG WV_SRC_URL="https://dl.google.com/widevine-cdm/VERSION-linux-x64.zip"

ARG LIBFFMPEG_URL="https://github.com/nwjs-ffmpeg-prebuilt/nwjs-ffmpeg-prebuilt/releases/download/0.67.1/0.67.1-linux-x64.zip"

#
# install opera
RUN apt-get update
RUN wget -O /tmp/opera.deb $SRC_URL
RUN apt-get install -y --no-install-recommends openbox unzip /tmp/opera.deb
# install widevine
#RUN echo $WV_SRC_URL | sed -e "s/VERSION/$(wget -q -O - $WV_VER_URL | tail -n1)/" | xargs wget -O /tmp/wv.zip
#RUN mkdir -p /usr/lib/x86_64-linux-gnu/opera/lib_extra/WidevineCdm/_platform_specific/linux_x64
#RUN unzip /tmp/wv.zip LICENSE.txt -d /usr/lib/x86_64-linux-gnu/opera/lib_extra/WidevineCdm/LICENSE
#RUN unzip /tmp/wv.zip manifest.json -d /usr/lib/x86_64-linux-gnu/opera/lib_extra/WidevineCdm
#RUN unzip /tmp/wv.zip libwidevinecdm.so -d /usr/lib/x86_64-linux-gnu/opera/lib_extra/WidevineCdm/_platform_specific/linux_x64
#RUN echo '[{"preload":"/usr/lib/x86_64-linux-gnu/opera/lib_extra/WidevineCdm"}]' > /usr/lib/x86_64-linux-gnu/opera/resources/widevine_config.json
#RUN find /usr/lib/x86_64-linux-gnu/opera/lib_extra -type d -exec chmod 0755 '{}' \;
#RUN find /usr/lib/x86_64-linux-gnu/opera/lib_extra -type f -exec chmod 0644 '{}' \;
## install libffmpeg
RUN wget -O /tmp/libffmpeg.zip $LIBFFMPEG_URL
RUN unzip -o /tmp/libffmpeg.zip libffmpeg.so -d /usr/lib/x86_64-linux-gnu/opera/lib_extra
RUN echo '[]' > /usr/lib/x86_64-linux-gnu/opera/resources/ffmpeg_preload_config.json

RUN ls -R /usr/lib/x86_64-linux-gnu/opera
#
# clean up
RUN apt-get clean -y
RUN rm -rf /var/lib/apt/lists/* /var/cache/apt/*

ENV VLC_MEDIA="/media"

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/vlc.conf
COPY --chown=neko preferences.json /home/neko/.config/opera/Default/Preferences
COPY policies.json /etc/opt/opera/policies/managed/policies.json
COPY openbox.xml /etc/neko/openbox.xml
