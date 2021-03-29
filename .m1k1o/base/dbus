#!/bin/sh

if [ ! -d /var/run/dbus ]; then
  mkdir -p /var/run/dbus
fi

if [ -f /var/run/dbus/pid ]; then
  rm -f /var/run/dbus/pid
fi

/usr/bin/dbus-daemon --nofork --print-pid --config-file=/usr/share/dbus-1/system.conf
