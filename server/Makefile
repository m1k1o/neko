BUILD_TIME=`date -u +'%Y-%m-%dT%H:%M:%SZ'`
GIT_COMMIT=`git rev-parse --short HEAD`
GIT_BRANCH=`git rev-parse --symbolic-full-name --abbrev-ref HEAD`
GIT_DIRTY=`git diff-index --quiet HEAD -- || echo "✗-"`

LDFLAGS=-ldflags "-s -X version.buildTime=${BUILD_TIME} -X version.gitRevision=${GIT_DIRTY}${GIT_REVISION} -X version.gitBranch=${GIT_BRANCH}"

build:
  go build -o bin/neko ${LDFLAGS} -i cmd/neko/main.go
