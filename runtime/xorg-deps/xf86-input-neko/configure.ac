#  Copyright 2005 Adam Jackson.
#  Copyright 2017 Martin Kepplinger.
#
#  Permission is hereby granted, free of charge, to any person obtaining a
#  copy of this software and associated documentation files (the "Software"),
#  to deal in the Software without restriction, including without limitation
#  on the rights to use, copy, modify, merge, publish, distribute, sub
#  license, and/or sell copies of the Software, and to permit persons to whom
#  the Software is furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice (including the next
#  paragraph) shall be included in all copies or substantial portions of the
#  Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NON-INFRINGEMENT.  IN NO EVENT SHALL
#  ADAM JACKSON BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
#  IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
#  CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
#
# Process this file with autoconf to produce a configure script

AC_PREREQ(2.59)
AC_INIT([xf86-input-neko],[0.0.1],[https://github.com/m1k1o/neko/issues],[xf86-input-neko],[https://github.com/m1k1o/neko])

AC_CONFIG_SRCDIR([Makefile.am])
AC_CONFIG_AUX_DIR(.)
AC_CONFIG_MACRO_DIR([m4])
AM_INIT_AUTOMAKE([dist-bzip2 dist-xz])

DRIVER_NAME=neko
AC_SUBST([DRIVER_NAME])

AC_CONFIG_HEADERS(config.h)

# Checks for programs.
AC_DISABLE_STATIC
AC_PROG_LIBTOOL
AC_PROG_CC

# Initialize libtool
LT_INIT

m4_ifndef([XORG_MACROS_VERSION],
          [m4_fatal([you must install xorg-macros before running autoconf/autogen.
  Hint: either install from source, git://anongit.freedesktop.org/xorg/util/macros or,
  depending on you distribution, try package 'xutils-dev' or 'xorg-x11-util-macros'])])
XORG_DEFAULT_OPTIONS

AH_TOP([#include "xorg-server.h"])

#AC_DEFINE(XFree86LOADER,1,[Stub define for loadable drivers])
#
#AC_ARG_ENABLE(XINPUT, AS_HELP_STRING([--enable-xinput],
#              [Build XInput support (default: yes)]),
#              [XINPUT=$enableval],[XINPUT=yes])
#AM_CONDITIONAL(XINPUT, test "x$XINPUT" = "xyes")
#if test "x$XINPUT" = "xyes" ; then
#    AC_DEFINE(XINPUT,1,[Enable XInput support])
#fi
#
#AC_ARG_ENABLE(XKB, AS_HELP_STRING([--enable-xkb],
#              [Build XKB support (default: yes)]),
#              [XKB=$enableval],[XKB=yes])
#AM_CONDITIONAL(XKB, test "x$XKB" = "xyes")
#if test "x$XKB" = "xyes" ; then
#    AC_DEFINE(XKB,1,[Enable XKB support])
#fi

AC_ARG_WITH(xorg-module-dir,
            AS_HELP_STRING(--with-xorg-module-dir=DIR,Default xorg module directory [[default=$libdir/xorg/modules]]),
            [moduledir="$withval"],
            [moduledir="$libdir/xorg/modules"])
inputdir=${moduledir}/input
AC_SUBST(inputdir)

AC_ARG_WITH(xorg-conf-dir,
            AC_HELP_STRING([--with-xorg-conf-dir=DIR],
                           [Default xorg.conf.d directory [[default=${prefix}/share/X11/xorg.conf.d/]]]),
            [xorgconfdir="$withval"],
            [xorgconfdir='${prefix}/share/X11/xorg.conf.d'])
AC_SUBST(xorgconfdir)

# Checks for extensions
XORG_DRIVER_CHECK_EXT(RANDR, randrproto)
XORG_DRIVER_CHECK_EXT(XINPUT, inputproto)

# Checks for pkg-config packages
PKG_CHECK_MODULES(XORG, xorg-server xproto $REQUIRED_MODULES)
sdkdir=$(pkg-config --variable=sdkdir xorg-server)

CFLAGS="$CFLAGS $XORG_CFLAGS "' -I$(top_srcdir)/src'
CFLAGS="-O2 -Wall -W -fPIC $CFLAGS"
AC_SUBST([CFLAGS])

# Checks for libraries.
#AC_CHECK_LIB(ts, ts_open, [], [
#        echo "Error! You need to have libts."
#        exit -1
#        ])

# Checks for header files.
AC_HEADER_STDC

AC_CONFIG_FILES([Makefile
		 src/Makefile
		 xorg-neko.pc])
AC_OUTPUT
