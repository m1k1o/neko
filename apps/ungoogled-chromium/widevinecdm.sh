#!/usr/bin/env bash
set -e

TARGET_DIR="$(realpath "$1")"
if [ -z "$TARGET_DIR" ]; then
    echo "Usage: $0 /path/to/install/WidevineCdm"
    exit 1
fi

TMPDIR=$(mktemp -d)
cd "$TMPDIR"

function cleanup {
    rm -rf "$TMPDIR"
}
trap cleanup EXIT

# Fetch manifest and extract URL
URL=$(python3 -c "
import json, urllib.request
data = json.load(urllib.request.urlopen('https://raw.githubusercontent.com/mozilla/gecko-dev/master/toolkit/content/gmp-sources/widevinecdm.json'))
for v in data['vendors'].values():
    for k, p in v['platforms'].items():
        if 'Linux_x86_64-gcc3' in k:
            print(p['fileUrl'])
            break
")

# Download CRX
curl -L -o widevinecdm.crx "$URL"

# Install go-crx3
echo "Fetching latest go-crx3 version..."
VERSION=$(curl -s https://api.github.com/repos/m1k1o/go-crx3/releases/latest | grep 'tag_name' | cut -d '"' -f4)
ARTIFACT="go-crx3_${VERSION#v}_linux_amd64.tar.gz"
URL="https://github.com/m1k1o/go-crx3/releases/download/${VERSION}/${ARTIFACT}"
echo "Downloading $URL"
curl -L -o "$ARTIFACT" "$URL"
tar -xzf "$ARTIFACT"

# Unpack with go-crx3
./go-crx3 unpack widevinecdm.crx
mkdir -p "$TARGET_DIR"
cp -ar widevinecdm/* "$TARGET_DIR"
