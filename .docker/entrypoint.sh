#!/bin/bash

echo "Starting dbus"
/etc/init.d/dbus start

echo "Starting supervisord"
su -p -l $NEKO_USER -c '/usr/bin/supervisord -c /etc/neko/supervisord.conf' -s /bin/bash