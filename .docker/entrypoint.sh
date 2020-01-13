#!/bin/bash

sudo /etc/init.d/dbus start

echo "Starting supervisord"
/usr/bin/supervisord -c /etc/neko/supervisord.conf