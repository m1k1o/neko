#!/bin/bash

# if no var is passed, noop
[[ -z "$RENDER_GID" ]] && exit 0

cnt_gid=$(getent group render | cut -d: -f3)
[[ -z "$cnt_gid" ]] && exit 1

# note that this could conceivably be a security risk...
cnt_group=$(getent group "$RENDER_GID" | cut -d: -f1)
if [[ -z "$cnt_group" ]]; then
  groupadd -g "$RENDER_GID" nekorender
  cnt_group=nekorender
fi
usermod -a -G "$cnt_group" "$USER"
