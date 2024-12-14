#!/usr/bin/sh

VERSION=202403162131

docker build -t neko-cave/base-nvidia:$VERSION -f ./.docker/base/Dockerfile.nvidia .

docker build --build-arg="BASE_IMAGE=neko-cave/base-nvidia:$VERSION" -t cave-xfce:$VERSION -f ./.docker/xfce/Dockerfile ./.docker/xfce/