#!/bin/bash
cd "$(dirname "$0")/.."

HELP_FILE="$(realpath -m docs/configuration/help.txt)"

pushd ../server
go run cmd/neko/main.go serve --help > $HELP_FILE
popd

# remove all lines with " V2: "
sed -i '/ V2: /d' $HELP_FILE
# remove all lines with " V2 DEPRECATED: "
sed -i '/ V2 DEPRECATED: /d' $HELP_FILE
# remove --legacy
sed -i '/--legacy/d' $HELP_FILE

# remove evething until first "Flags:"
sed -i '1,/Flags:/d' $HELP_FILE
# remove --help
sed -i '/--help/d' $HELP_FILE

npm run gen-config
