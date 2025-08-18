#
# Custom Firefox image for n.eko
#
# This Dockerfile inherits from the official n.eko firefox image.
# You can add your own custom commands here to install additional software
# or make other modifications to the image.
#

ARG BASE_IMAGE=ghcr.io/m1k1o/neko/firefox:latest
FROM $BASE_IMAGE

#
# Add your custom commands here.
#
# For example, to install a browser extension, you can use:
#
# ARG ADDON_URL="<URL to the .xpi file>"
# RUN wget -O /tmp/addon.xpi "${ADDON_URL}" && \
#     firefox -install-global-extension /tmp/addon.xpi && \
#     rm /tmp/addon.xpi
#
