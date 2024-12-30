#!/bin/bash
: 'Copyright (C) 2017, Martin Kepplinger <martink@posteo.de>

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software
is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.'

# Script to build release-archives with. This requires a checkout from git.
# Before running it, change the version in configure.ac. Only do this. Do
# *not* commit this change. This script will do it.

# WARNING: This script is very dangerous! It may delete any untracked files.

have_version=0

usage()
{
        echo "Usage: $0 -v version"
}

args=$(getopt -o v:s -- "$@")
if [ $? -ne 0 ] ; then
        usage
        exit 1
fi
eval set -- "$args"
while [ $# -gt 0 ]
do
        case "$1" in
        -v)
                version=$2
                have_version=1
                shift
                ;;
        --)
                shift
                break
                ;;
        *)
                echo "Invalid option: $1"
                usage
                exit 1
                ;;
        esac
        shift
done

# Do we have a desired version number?
if [ "$have_version" -gt 0 ] ; then
       echo "trying to build version $version"
else
       echo "please specify a version"
       usage
       exit 1
fi

# Version number sanity check
if grep ${version} configure.ac
then
       echo "configurations seems ok"
else
       echo "please check your configure.ac"
       exit 1
fi

# Check that we are on master
branch=$(git rev-parse --abbrev-ref HEAD)
echo "we are on branch $branch"

if [ ! "${branch}" = "master" ] ; then
	echo "you don't seem to be on the master branch"
	exit 1
fi

if git diff-index --quiet HEAD --; then
	# no changes
	echo "there are no uncommitted changes (version bump)"
	exit 1
fi
echo "======================================================"
echo "    are you fine with the following version bump?"
echo "======================================================"
git diff
echo "======================================================"
read -p "           Press enter to continue"
echo "======================================================"

./autogen.sh && ./configure && make distcheck
./autogen-clean.sh
git clean -d -f

git commit -a -m "xf86-input-neko ${version}"
git tag -s ${version} -m "xf86-input-neko ${version}"

./autogen.sh && ./configure && make distcheck
sha256sum xf86-input-neko-${version}.tar.xz > xf86-input-neko-${version}.tar.xz.sha256
sha256sum xf86-input-neko-${version}.tar.gz > xf86-input-neko-${version}.tar.gz.sha256
sha256sum xf86-input-neko-${version}.tar.bz2 > xf86-input-neko-${version}.tar.bz2.sha256

sha512sum xf86-input-neko-${version}.tar.xz > xf86-input-neko-${version}.tar.xz.sha512
sha512sum xf86-input-neko-${version}.tar.gz > xf86-input-neko-${version}.tar.gz.sha512
sha512sum xf86-input-neko-${version}.tar.bz2 > xf86-input-neko-${version}.tar.bz2.sha512

gpg -b -a xf86-input-neko-${version}.tar.xz
gpg -b -a xf86-input-neko-${version}.tar.gz
gpg -b -a xf86-input-neko-${version}.tar.bz2

