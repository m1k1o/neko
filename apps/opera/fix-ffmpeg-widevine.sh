#!/bin/bash
# From https://github.com/Ld-Hagen/fix-opera-linux-ffmpeg-widevine.

# Config section
readonly FIX_FFMPEG=true
readonly FIX_WIDEVINE=true
readonly FIX_DIR='/tmp/opera-fix'
# readonly FFMPEG_SRC_MAIN='https://api.github.com/repos/nwjs-ffmpeg-prebuilt/nwjs-ffmpeg-prebuilt/releases'
readonly FFMPEG_SRC_MAIN='https://api.github.com/repos/Ld-Hagen/nwjs-ffmpeg-prebuilt/releases'
readonly FFMPEG_SRC_ALT='https://api.github.com/repos/Ld-Hagen/fix-opera-linux-ffmpeg-widevine/releases'
readonly WIDEVINE_SRC='https://raw.githubusercontent.com/mozilla-firefox/firefox/refs/heads/main/toolkit/content/gmp-sources/widevinecdm.json'
readonly FFMPEG_SO_NAME='libffmpeg.so'
readonly WIDEVINE_SO_NAME='libwidevinecdm.so'
readonly WIDEVINE_MANIFEST_NAME='manifest.json'

OPERA_VERSIONS=()
if [ -x "$(command -v opera)" ]; then
  OPERA_VERSIONS+=("opera")
fi
if [ -x "$(command -v opera-beta)" ]; then
  OPERA_VERSIONS+=("opera-beta")
fi

# Getting download links
printf 'Getting download links...\n'
## ffmpeg
if $FIX_FFMPEG; then
  readonly FFMPEG_URL_MAIN=$(curl -sL4 $FFMPEG_SRC_MAIN | jq -rS 'sort_by(.published_at) | .[-1].assets[0].browser_download_url')
  readonly FFMPEG_URL_ALT=$(curl -sL4 $FFMPEG_SRC_ALT | jq -rS 'sort_by(.published_at) | .[-1].assets[0].browser_download_url')
  [[ $(basename $FFMPEG_URL_ALT) < $(basename $FFMPEG_URL_MAIN) ]] && readonly FFMPEG_URL=$FFMPEG_URL_MAIN || readonly FFMPEG_URL=$FFMPEG_URL_ALT
  if [[ -z $FFMPEG_URL ]]; then
    printf 'Failed to get ffmpeg download URL. Exiting...\n'
    exit 1
  fi
fi
## Widevine
if $FIX_WIDEVINE; then
  readonly WIDEVINE_URL=$(curl -sL4 $WIDEVINE_SRC | jq -r '.vendors."gmp-widevinecdm".platforms."Linux_x86_64-gcc3".mirrorUrls[0]')
fi

# Downloading files
printf 'Downloading files...\n'
mkdir -p "$FIX_DIR"
## ffmpeg
if $FIX_FFMPEG; then
  curl -L4 --progress-bar $FFMPEG_URL -o "$FIX_DIR/ffmpeg.zip"
  if [ $? -ne 0 ]; then
    printf 'Failed to download ffmpeg. Check your internet connection or try later\n'
    exit 1
  fi
fi
## Widevine
if $FIX_WIDEVINE;  then
  curl -L4 --progress-bar "$WIDEVINE_URL" -o "$FIX_DIR/widevine.zip"
  if [ $? -ne 0 ]; then
    printf 'Failed to download Widevine CDM. Check your internet connection or try later\n'
    exit 1
  fi
fi

# Extracting files
## ffmpeg
if $FIX_FFMPEG; then
  echo "Extracting ffmpeg..."
  unzip -o "$FIX_DIR/ffmpeg.zip" -d $FIX_DIR > /dev/null
fi
## Widevine
if $FIX_WIDEVINE; then
  echo "Extracting WidevineCDM..."
  unzip -oj "$FIX_DIR/widevine.zip" -d $FIX_DIR > /dev/null 2>/dev/null
fi

for opera in ${OPERA_VERSIONS[@]}; do
  echo "Doing $opera"
  EXECUTABLE=$(command -v "$opera")
  OPERA_DIR=$(dirname $(readlink -f $EXECUTABLE))
  OPERA_LIB_DIR="$OPERA_DIR/lib_extra"
  OPERA_WIDEVINE_DIR="$OPERA_LIB_DIR/WidevineCdm"
  OPERA_WIDEVINE_SO_DIR="$OPERA_WIDEVINE_DIR/_platform_specific/linux_x64"
  OPERA_WIDEVINE_CONFIG="$OPERA_DIR/resources/widevine_config.json"

  # Removing old libraries and preparing directories
  printf 'Removing old libraries & making directories...\n'
  ## ffmpeg
  if $FIX_FFMPEG; then
    rm -f "$OPERA_LIB_DIR/$FFMPEG_SO_NAME"
    mkdir -p "$OPERA_LIB_DIR"
  fi
  ## Widevine
  if $FIX_WIDEVINE; then
    rm -rf "$OPERA_WIDEVINE_DIR"
    mkdir -p "$OPERA_WIDEVINE_SO_DIR"
  fi

  # Moving libraries to their place
  printf 'Moving libraries to their places...\n'
  ## ffmpeg
  if $FIX_FFMPEG; then
    cp -f "$FIX_DIR/$FFMPEG_SO_NAME" "$OPERA_LIB_DIR"
    chmod 0644 "$OPERA_LIB_DIR/$FFMPEG_SO_NAME"
  fi
  ## Widevine
  if $FIX_WIDEVINE; then
    cp -f "$FIX_DIR/$WIDEVINE_SO_NAME" "$OPERA_WIDEVINE_SO_DIR"
    chmod 0644 "$OPERA_WIDEVINE_SO_DIR/$WIDEVINE_SO_NAME"
    cp -f "$FIX_DIR/$WIDEVINE_MANIFEST_NAME" "$OPERA_WIDEVINE_DIR"
    chmod 0644 "$OPERA_WIDEVINE_DIR/$WIDEVINE_MANIFEST_NAME"
    printf "[\n      {\n         \"preload\": \"$OPERA_WIDEVINE_DIR\"\n      }\n]\n" > "$OPERA_WIDEVINE_CONFIG"
  fi
done

# Removing temporary files
printf 'Removing temporary files...\n'
rm -rf "$FIX_DIR"
