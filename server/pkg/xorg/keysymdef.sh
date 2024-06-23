#!/bin/sh

wget https://cgit.freedesktop.org/xorg/proto/x11proto/plain/keysymdef.h
sed -i -E 's/\#define (XK_[a-zA-Z_0-9]+\s+)(0x[0-9a-f]+)/const \1 = \2/g' keysymdef.h
sed -i -E 's/^\#/\/\//g' keysymdef.h
echo "package xorg" | cat - keysymdef.h > keysymdef.go && rm keysymdef.h
