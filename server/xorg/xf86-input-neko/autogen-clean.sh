#!/bin/sh
if [ -f Makefile ]; then
	echo "Making make distclean..."
	make distclean
fi
echo "Removing autogenned files..."
rm -f config.guess config.sub configure install-sh missing mkinstalldirs Makefile.in ltmain.sh stamp-h.in */Makefile.in ltconfig stamp-h config.h.in* aclocal.m4 compile depcomp
echo "Done."
