ARG BASE_IMAGE=m1k1o/neko:base
FROM $BASE_IMAGE

ARG SRC_URL="https://download.opera.com/download/get/?id=58545&location=415&nothanks=yes&sub=marine&utm_tryagain=yes"

ARG LIBFFMPEG_URL="https://github.com/nwjs-ffmpeg-prebuilt/nwjs-ffmpeg-prebuilt/releases/download/0.67.1/0.67.1-linux-x64.zip"

#
# install opera
RUN apt-get update
RUN wget -O /tmp/opera.deb $SRC_URL
RUN apt-get install -y --no-install-recommends openbox unzip /tmp/opera.deb

## install libffmpeg
RUN wget -O /tmp/libffmpeg.zip $LIBFFMPEG_URL
RUN unzip -o /tmp/libffmpeg.zip libffmpeg.so -d /usr/lib/x86_64-linux-gnu/opera/lib_extra
RUN echo '[]' > /usr/lib/x86_64-linux-gnu/opera/resources/ffmpeg_preload_config.json
#
# clean up
RUN apt-get clean -y
RUN rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# copy configuation files
COPY supervisord.conf /etc/neko/supervisord/opera.conf
COPY openbox.xml /etc/neko/openbox.xml

