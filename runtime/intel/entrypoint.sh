#!/bin/bash

# supervisord reads /etc/group to determine the supplementary groups of every
# program it spawns, so the render group has to be assigned before supervisord
# starts. Doing it from a program spawned by supervisord (as the previous
# 'rendergroup-init' did) races with the neko program: neko is usually spawned
# before 'usermod' has finished and is left without access to /dev/dri/render*,
# which in turn makes the VA-API encoders fail to register.
#
# Assign the render group to the neko user first, then hand over to supervisord.
if ! /usr/bin/add-render-group.sh; then
  echo "entrypoint: could not assign the render group, VA-API may not work (is /dev/dri passed to the container?)" >&2
fi

exec "$@"
