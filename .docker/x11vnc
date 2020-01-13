#!/bin/bash

if [ ! -f "${HOME}/.vnc/passwd" ]; then
  x11vnc -storepasswd
fi

/usr/bin/x11vnc -display :0 -6 -xkb -rfbport 5901 -rfbauth $HOME/.vnc/passwd -wait 20 -nap -noxrecord -nopw -noxfixes -noxdamage -repeat
