#!/usr/bin/sh

VERSION=202401261754

docker build -t neko-cave/base-nvidia:$VERSION -f ./.docker/base/Dockerfile.nvidia .

docker build --build-arg="BASE_IMAGE=neko-cave/base-nvidia:$VERSION" -t cave-firefox:$VERSION -f ./.docker/firefox/Dockerfile.nvidia ./.docker/firefox/