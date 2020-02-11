FROM debian:stretch-slim

#
# cluster fuck of packages for neko, node, go and gstreamer
RUN set -eux; apt-get update; apt-get install -y --no-install-recommends \
  autoconf ca-certificates curl netbase wget \
  bzr git mercurial openssh-client subversion procps cmake automake bzip2 dpkg-dev file g++ gcc \
  libbz2-dev libc6-dev libcurl4-openssl-dev libdb-dev libevent-dev libffi-dev libgdbm-dev libglib2.0-dev libgmp-dev \
  libjpeg-dev libkrb5-dev liblzma-dev libmagickcore-dev libmagickwand-dev libmaxminddb-dev libncurses5-dev libncursesw5-dev \
  libpng-dev libpq-dev libreadline-dev libsqlite3-dev libssl-dev libtool libwebp-dev libxml2-dev libxslt-dev libyaml-dev  \
  make patch unzip xz-utils zlib1g-dev pkg-config \
  build-essential perl python autopoint bison flex \
  gettext openssl libopus-dev libvpx-dev libpulse-dev libx11-dev libxv-dev libxt-dev libxrandr-dev \
  libxfixes-dev apt-utils x11vnc libxtst-dev dialog \
  pulseaudio openbox chromium firefox-esr dbus-x11 xserver-xorg-video-dummy supervisor; \
  if ! command -v gpg > /dev/null; then \
    apt-get install -y --no-install-recommends gnupg dirmngr; \
  fi

#
# set up env for gst
ENV PATH=/gst/local/bin:$PATH
ENV LD_LIBRARY_PATH=/gst/local/lib:$LD_LIBRARY_PATH
ENV PKG_CONFIG_PATH=/gst/local/lib/pkgconfig:$PKG_CONFIG_PATH

#
# add node
ENV NODE_VERSION 12.14.1
RUN set -eux; \
  ARCH= ; dpkgArch="$(dpkg --print-architecture)" \
  ;case "${dpkgArch##*-}" in \
    amd64) ARCH='x64';; \
    ppc64el) ARCH='ppc64le';; \
    s390x) ARCH='s390x';; \
    arm64) ARCH='arm64';; \
    armhf) ARCH='armv7l';; \
    i386) ARCH='x86';; \
    *) echo "unsupported architecture"; exit 1 ;; \
  esac;\
  # gpg keys listed at https://github.com/nodejs/node#release-keys
  for key in \
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
    gpg --batch --keyserver hkp://pgp.mit.edu:80 --recv-keys "$key"; \
  done; \
  curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-$ARCH.tar.xz"; \
  curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc"; \
  gpg --batch --decrypt --output SHASUMS256.txt SHASUMS256.txt.asc; \
  grep " node-v$NODE_VERSION-linux-$ARCH.tar.xz\$" SHASUMS256.txt | sha256sum -c - ; \
  tar -xJf "node-v$NODE_VERSION-linux-$ARCH.tar.xz" -C /usr/local --strip-components=1 --no-same-owner; \
  rm "node-v$NODE_VERSION-linux-$ARCH.tar.xz" SHASUMS256.txt.asc SHASUMS256.txt; \
  ln -s /usr/local/bin/node /usr/local/bin/nodejs

#
# add go
ENV GOLANG_VERSION 1.13.6
RUN set -eux; \
  dpkgArch="$(dpkg --print-architecture)"; \
  case "${dpkgArch##*-}" in \
    amd64) goRelArch='linux-amd64'; goRelSha256='a1bc06deb070155c4f67c579f896a45eeda5a8fa54f35ba233304074c4abbbbd' ;; \
    armhf) goRelArch='linux-armv6l'; goRelSha256='37a1a83e363dcf146a67fa839d170fd1afb13009585fdd493d0a3370fbe6f785' ;; \
    arm64) goRelArch='linux-arm64'; goRelSha256='0a18125c4ed80f9c3045cf92384670907c4796b43ed63c4307210fe93e5bbca5' ;; \
    i386) goRelArch='linux-386'; goRelSha256='27feb013106da784f09e560720aa41ab395c67f7eed4c4a0fce04bc6e3d01c7d' ;; \
    ppc64el) goRelArch='linux-ppc64le'; goRelSha256='26a977a8af5dc50a562f0a57b58dded5fa3bacfe77722cf8a84ea54ca54728dd' ;; \
    s390x) goRelArch='linux-s390x'; goRelSha256='5cd9900a1fa0f0cac657930b648381cad9b8c5e2bbc77caf86a6fb5cedad0017' ;; \
    *) goRelArch='src'; goRelSha256='aae5be954bdc40bcf8006eb77e8d8a5dde412722bc8effcdaf9772620d06420c'; \
      echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
  esac; \
  url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
  wget -O go.tgz "$url"; \
  echo "${goRelSha256} *go.tgz" | sha256sum -c - ; \
  tar -C /usr/local -xzf go.tgz; \
  rm go.tgz; \
  if [ "$goRelArch" = 'src' ]; then \
    echo >&2; \
    echo >&2 'error: UNIMPLEMENTED'; \
    echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
    echo >&2; \
    exit 1; \
  fi; \
  export PATH="/usr/local/go/bin:$PATH"; \
  go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

#
# install libclipboard
RUN set -eux; \
    cd /tmp ; \
    git clone https://github.com/jtanx/libclipboard ; \
    cd libclipboard ; \
    cmake . ; \
    make -j4; \
    make install; \ 
    rm -rf /tmp/libclipboard

#
# install Go tools w/module support
RUN set -eux; \
    mkdir -p /tmp/gotools; \
    cd /tmp/gotools; \
    GO111MODULE=on go get -v golang.org/x/tools/gopls@latest 2>&1; \
    GO111MODULE=on go get -v \
        honnef.co/go/tools/...@latest \
        golang.org/x/tools/cmd/gorename@latest \
        golang.org/x/tools/cmd/goimports@latest \
        golang.org/x/tools/cmd/guru@latest \
        golang.org/x/lint/golint@latest \
        github.com/mdempsky/gocode@latest \
        github.com/cweill/gotests/...@latest \
        github.com/haya14busa/goplay/cmd/goplay@latest \
        github.com/sqs/goreturns@latest \
        github.com/josharian/impl@latest \
        github.com/davidrjenni/reftools/cmd/fillstruct@latest \
        github.com/uudashr/gopkgs/cmd/gopkgs@latest  \
        github.com/ramya-rao-a/go-outline@latest  \
        github.com/acroca/go-symbols@latest  \
        github.com/godoctor/godoctor@latest  \
        github.com/rogpeppe/godef@latest  \
        github.com/zmb3/gogetdoc@latest \
        github.com/fatih/gomodifytags@latest  \
        github.com/mgechev/revive@latest  \
        github.com/go-delve/delve/cmd/dlv@latest 2>&1; \
    #
    # install Go tools w/o module support
    go get -v github.com/alecthomas/gometalinter 2>&1; \
    #
    # install gocode-gomod
    go get -x -d github.com/stamblerre/gocode 2>&1; \
    go build -o gocode-gomod github.com/stamblerre/gocode; \
    mv gocode-gomod $GOPATH/bin/; \
    #
    # install golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin 2>&1; \
    mkdir -p "$GOPATH/src" "$GOPATH/bin" "$GOPATH/pkg/mod"; chmod -R 777 "$GOPATH"
#
# turn on go modules
ENV GO111MODULE=on

# the node image includes a non-root user with sudo access. Use the "remoteUser"
# property in devcontainer.json to use it. On Linux, the container user's GID/UIDs
# will be updated to match your local UID/GID (when using the dockerFile property).
# See https://aka.ms/vscode-remote/containers/non-root-user for details.
ARG USERNAME=neko
ARG USER_UID=1000
ARG USER_GID=$USER_UID

#
# set to false to skip installing zsh and Oh My ZSH!
ARG INSTALL_ZSH="true"

#
# location and expected SHA for common setup script - SHA generated on release
ARG COMMON_SCRIPT_SOURCE="https://raw.githubusercontent.com/microsoft/vscode-dev-containers/master/script-library/common-debian.sh"
ARG COMMON_SCRIPT_SHA="dev-mode"

#
# Docker Compose version
ARG COMPOSE_VERSION=1.24.0

#
# verify git, common tools / libs installed, add/modify non-root user, optionally install zsh
RUN set -eux; \
    wget -q -O /tmp/common-setup.sh $COMMON_SCRIPT_SOURCE; \
    if [ "$COMMON_SCRIPT_SHA" != "dev-mode" ]; then echo "$COMMON_SCRIPT_SHA /tmp/common-setup.sh" | sha256sum -c - ; fi; \
    /bin/bash /tmp/common-setup.sh "$INSTALL_ZSH" "$USERNAME" "$USER_UID" "$USER_GID"; \
    rm /tmp/common-setup.sh; \
    #
    # install docker
    apt-get install -y apt-transport-https gnupg-agent software-properties-common lsb-release; \
    curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | (OUT=$(apt-key add - 2>&1) || echo $OUT); \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable"; \
    apt-get update; apt-get install -y docker-ce-cli; \
    #
    # install docker compose
    curl -sSL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose; \
    chmod +x /usr/local/bin/docker-compose; \
    #
    # set alternate global install location that both users have rights to access
    mkdir -p /usr/local/share/npm-global; \
    chown ${USERNAME}:root /usr/local/share/npm-global; \
    npm config -g set prefix /usr/local/share/npm-global; \
    sudo -u ${USERNAME} npm config -g set prefix /usr/local/share/npm-global

ENV PATH=/usr/local/share/npm-global/bin:$PATH

#
# switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog
