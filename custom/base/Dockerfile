#
# Custom base image for n.eko
#
# This Dockerfile inherits from the official n.eko base image.
# You can add your own custom commands here to install additional software
# or make other modifications to the base image.
#

ARG BASE_IMAGE=ghcr.io/m1k1o/neko/base:latest
FROM $BASE_IMAGE

#
# Add your custom commands here.
#
# For example, to install a package, you can use:
# RUN set -eux; apt-get update; \
#     apt-get install -y --no-install-recommends <package-name>; \
#     #
#     # clean up
#     apt-get clean -y; \
#     rm -rf /var/lib/apt/lists/* /var/cache/apt/*
#
