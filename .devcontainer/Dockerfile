FROM golang:stretch

# This Dockerfile adds a non-root user with sudo access. Use the "remoteUser"
# property in devcontainer.json to use it. On Linux, the container user's GID/UIDs
# will be updated to match your local UID/GID (when using the dockerFile property).
# See https://aka.ms/vscode-remote/containers/non-root-user for details.
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# Runtime for testing and compileing
RUN apt-get update \
  && apt-get -y install supervisor openbox dbus-x11 ttf-freefont xvfb pulseaudio consolekit firefox-esr x11vnc \
  && apt-get -y install libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev gstreamer1.0-plugins-good gstreamer1.0-pulseaudio \
  && apt-get -y install gcc libc6-dev \
  && apt-get -y install libx11-dev xorg-dev libxtst-dev libpng++-dev \
  && apt-get -y install xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev \
  && apt-get -y install libxkbcommon-dev \
  && apt-get -y install xsel xclip 

# Configure apt, install packages and tools
RUN apt-get update \
  && apt-get -y install --no-install-recommends apt-utils apt-transport-https dialog 2>&1 \
  #
  # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
  && apt-get -y install git iproute2 procps lsb-release xz-utils \
  #
  # Install Docker CE CLI
  && apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common lsb-release \
  && curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | (OUT=$(apt-key add - 2>&1) || echo $OUT) \
  && add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" \
  && apt-get update \
  && apt-get install -y docker-ce-cli \
  #
  # Install Docker Compose
  && curl -sSL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
  && chmod +x /usr/local/bin/docker-compose \
  #
  # Create a non-root user to use if preferred - see https://aka.ms/vscode-remote/containers/non-root-user.
  && groupadd --gid $USER_GID $USERNAME \
  && useradd --uid $USER_UID --gid $USERNAME --shell /bin/bash --create-home $USERNAME \
  && adduser $USERNAME audio \
  && adduser $USERNAME video \
  && adduser $USERNAME pulse \
  # Add sudo support for the non-root user
  && apt-get install -y sudo \
  && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
  && chmod 0440 /etc/sudoers.d/$USERNAME

ENV NODE_VERSION 12.14.1

RUN ARCH= && dpkgArch="$(dpkg --print-architecture)" \
  && case "${dpkgArch##*-}" in \
    amd64) ARCH='x64';; \
    ppc64el) ARCH='ppc64le';; \
    s390x) ARCH='s390x';; \
    arm64) ARCH='arm64';; \
    armhf) ARCH='armv7l';; \
    i386) ARCH='x86';; \
    *) echo "unsupported architecture"; exit 1 ;; \
  esac \
  # gpg keys listed at https://github.com/nodejs/node#release-keys
  && set -ex \
  && for key in \
    94AE36675C464D64BAFA68DD7434390BDBE9B9C5 \
    FD3A5288F042B6850C66B31F09FE44734EB7990E \
    71DCFD284A79C3B38668286BC97EC7A07EDE3FC1 \
    DD8F2338BAE7501E3DD5AC78C273792F7D83545D \
    C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8 \
    B9AE9905FFD7803F25714661B63B535A4C206CA9 \
    77984A986EBC2AA786BC0F66B01FBB92821C587A \
    8FCCA13FEF1D0C2E91008E09770F7A9A5AE15600 \
    4ED778F539E3634C779C87C6D7062848A1AB005C \
    A48C2BEE680E841632CD4E44F07496B3EB3C1762 \
    B9E2F5981AA6E0CD28160D9FF13993A75599653C \
  ; do \
    gpg --batch --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys "$key" || \
    gpg --batch --keyserver hkp://ipv4.pool.sks-keyservers.net --recv-keys "$key" || \
    gpg --batch --keyserver hkp://pgp.mit.edu:80 --recv-keys "$key" ; \
  done \
  && curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-$ARCH.tar.xz" \
  && curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc" \
  && gpg --batch --decrypt --output SHASUMS256.txt SHASUMS256.txt.asc \
  && grep " node-v$NODE_VERSION-linux-$ARCH.tar.xz\$" SHASUMS256.txt | sha256sum -c - \
  && tar -xJf "node-v$NODE_VERSION-linux-$ARCH.tar.xz" -C /usr/local --strip-components=1 --no-same-owner \
  && rm "node-v$NODE_VERSION-linux-$ARCH.tar.xz" SHASUMS256.txt.asc SHASUMS256.txt \
  && ln -s /usr/local/bin/node /usr/local/bin/nodejs

USER $USERNAME

# Install packages and tools
RUN go get -x -d github.com/stamblerre/gocode 2>&1 \
    && go build -o gocode-gomod github.com/stamblerre/gocode \
    && mv gocode-gomod $GOPATH/bin/ \
    #
    # Install Go tools
    && go get -u -v \
      github.com/mdempsky/gocode \
      github.com/uudashr/gopkgs/cmd/gopkgs \
      github.com/ramya-rao-a/go-outline \
      github.com/acroca/go-symbols \
      github.com/godoctor/godoctor \
      golang.org/x/tools/cmd/guru \
      golang.org/x/tools/cmd/gorename \
      github.com/rogpeppe/godef \
      github.com/zmb3/gogetdoc \
      github.com/haya14busa/goplay/cmd/goplay \
      github.com/sqs/goreturns \
      github.com/josharian/impl \
      github.com/davidrjenni/reftools/cmd/fillstruct \
      github.com/fatih/gomodifytags \
      github.com/cweill/gotests/... \
      golang.org/x/tools/cmd/goimports \
      golang.org/x/lint/golint \
      golang.org/x/tools/gopls \
      github.com/alecthomas/gometalinter \
      honnef.co/go/tools/... \
      github.com/golangci/golangci-lint/cmd/golangci-lint \
      github.com/mgechev/revive \
      github.com/derekparker/delve/cmd/dlv 2>&1 

ENV GO111MODULE=on

# Switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog

