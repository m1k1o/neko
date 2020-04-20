FROM debian:stretch-slim

#
# cluster fuck of packages for neko, node, go and gstreamer
RUN set -eux; apt-get update; apt-get install -y --no-install-recommends \
  apt-transport-https gnupg-agent software-properties-common lsb-release \
  autoconf ca-certificates curl netbase wget gnupg dirmngr libatomic1 \
  libghc-zlib-dev libexpat1-dev \
  bzr mercurial openssh-client subversion procps cmake automake bzip2 dpkg-dev file g++ gcc \
  libbz2-dev libc6-dev libcurl4-openssl-dev libdb-dev libevent-dev libffi-dev libgdbm-dev libglib2.0-dev libgmp-dev \
  libjpeg-dev libkrb5-dev liblzma-dev libmagickcore-dev libmagickwand-dev libmaxminddb-dev libncurses5-dev libncursesw5-dev \
  libpng-dev libpq-dev libreadline-dev libsqlite3-dev libssl-dev libtool libwebp-dev libxml2-dev libxslt-dev libyaml-dev  \
  make patch unzip xz-utils zlib1g-dev pkg-config \
  build-essential perl python autopoint bison flex \
  gettext openssl libopus-dev libvpx-dev libpulse-dev libx11-dev libxv-dev libxt-dev libxrandr-dev \
  libxfixes-dev apt-utils x11vnc libxtst-dev dialog \
  pulseaudio openbox chromium firefox-esr dbus-x11 xserver-xorg-video-dummy supervisor;

#
# add git
ENV GIT_VERSION 2.21.0

RUN set -eux; \
  cd /tmp/; \
  wget https://github.com/git/git/archive/v$GIT_VERSION.zip -O latestgit.zip; \
  unzip latestgit.zip; \
  cd git-$GIT_VERSION; \
  make prefix=/usr/local all; \
  make prefix=/usr/local install; \
  rm -rf /tmp/latestgit.zip /tmp/git-$GIT_VERSION;

#
# install libclipboard
RUN set -eux; \
    cd /tmp; \
    git clone https://github.com/jtanx/libclipboard; \
    cd libclipboard; \
    cmake . ; \
    make -j4; \
    make install; \
    rm -rf /tmp/libclipboard;

#
# set up env for gst
ENV PATH=/gst/local/bin:$PATH
ENV LD_LIBRARY_PATH=/gst/local/lib:$LD_LIBRARY_PATH
ENV PKG_CONFIG_PATH=/gst/local/lib/pkgconfig:$PKG_CONFIG_PATH

#
# copy gst
COPY .build/gst/local /gst/local/

#
# add node
ENV NODE_VERSION 12.16.2
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
    && apt-mark auto '.*' > /dev/null \
    && find /usr/local -type f -executable -exec ldd '{}' ';' \
      | awk '/=>/ { print $(NF-1) }' \
      | sort -u \
      | xargs -r dpkg-query --search \
      | cut -d: -f1 \
      | sort -u \
      | xargs -r apt-mark manual \
    && ln -s /usr/local/bin/node /usr/local/bin/nodejs \
    # smoke tests
    && node --version \
    && npm --version

#
# add golang
ENV GOLANG_VERSION 1.14.2
RUN set -eux; \
	dpkgArch="$(dpkg --print-architecture)"; \
	case "${dpkgArch##*-}" in \
		amd64) goRelArch='linux-amd64'; goRelSha256='6272d6e940ecb71ea5636ddb5fab3933e087c1356173c61f4a803895e947ebb3' ;; \
		armhf) goRelArch='linux-armv6l'; goRelSha256='eb4550ba741506c2a4057ea4d3a5ad7ed5a887de67c7232f1e4795464361c83c' ;; \
		arm64) goRelArch='linux-arm64'; goRelSha256='bb6d22fe5806352c3d0826676654e09b6e41eb1af52e8d506d3fa85adf7f8d88' ;; \
		i386) goRelArch='linux-386'; goRelSha256='cab5f51e6ffb616c6ee963c3d0650ca4e3c4108307c44f2baf233fcb8ff098f6' ;; \
		ppc64el) goRelArch='linux-ppc64le'; goRelSha256='48c22268c81ced9084a43bbe2c1596d3e636b5560b30a32434a7f15e561de160' ;; \
		s390x) goRelArch='linux-s390x'; goRelSha256='501cc919648c9d85b901963303c5061ea6814c80f0d35fda9e62980d3ff58cf4' ;; \
		*) goRelArch='src'; goRelSha256='98de84e69726a66da7b4e58eac41b99cbe274d7e8906eeb8a5b7eb0aadee7f7c'; \
			echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
	esac; \
	url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
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
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

#
# install Go tools w/module support
RUN mkdir -p /tmp/gotools \
    && cd /tmp/gotools \
    && GOPATH=/tmp/gotools GO111MODULE=on go get -v golang.org/x/tools/gopls@latest 2>&1 \
    && GOPATH=/tmp/gotools GO111MODULE=on go get -v \
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
        github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest  \
        github.com/ramya-rao-a/go-outline@latest  \
        github.com/acroca/go-symbols@latest  \
        github.com/godoctor/godoctor@latest  \
        github.com/rogpeppe/godef@latest  \
        github.com/zmb3/gogetdoc@latest \
        github.com/fatih/gomodifytags@latest  \
        github.com/mgechev/revive@latest  \
        github.com/go-delve/delve/cmd/dlv@latest 2>&1 \
    #
    # build Go tools w/o module support
    && GOPATH=/tmp/gotools go get -v github.com/alecthomas/gometalinter 2>&1 \
    #
    # build gocode-gomod
    && GOPATH=/tmp/gotools go get -x -d github.com/stamblerre/gocode 2>&1 \
    && GOPATH=/tmp/gotools go build -o gocode-gomod github.com/stamblerre/gocode \
    #
    # install Go tools
    && mv /tmp/gotools/bin/* /usr/local/bin/ \
    && mv gocode-gomod /usr/local/bin/ \
    #
    # install golangci-lint
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin 2>&1 \
    #
    # setup paths
    && mkdir -p "$GOPATH/src" "$GOPATH/bin" "$GOPATH/pkg/mod" \
    && chmod -R 777 "$GOPATH" 

#
# turn on go modules
ENV GO111MODULE=on

#
# install docker
RUN set -eux; \
    curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | (OUT=$(apt-key add - 2>&1) || echo $OUT); \
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable"; \
    apt-get update; apt-get install -y docker-ce-cli;