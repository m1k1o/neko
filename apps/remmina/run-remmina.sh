#!/bin/bash

err() {
  echo "ERROR: $*" >&2
  exit 1
}

profile_dir="/home/neko/.local/share/remmina"
if [[ -n "$REMMINA_PROFILE" ]]; then
  profile=${REMMINA_PROFILE%.remmina}.remmina
  file=${profile##/*/}
  [[ "$file" = "$profile" ]] && profile="$profile_dir"/"$file"
  [[ -f "$profile" ]] || err "Connection profile $profile not found"
  echo "Running remmina with connection profile $profile"
  exec remmina -c "$profile"
fi

if [[ ! -z "$REMMINA_URL" ]]; then
  readarray -t arr < <( echo -n "$REMMINA_URL" | perl -pe 's|^(\w+)\:\/\/(?:([^:]+)(?::([^@]+))?@)?(.*)$|\1\n\2\n\3\n\4|' )
  proto="${arr[0]}"
  user="${arr[1]}"
  pw="${arr[2]}"
  host="${arr[3]}"
  echo "Parsed url in 'REMMINA_URL': proto:$proto username:$user host:$host"

  [[ "$proto" != "vnc" && "$proto" != "rdp" && "$proto" != "spice" ]] && err "Unsupported protocol $proto in connection url 'REMMINA_URL'"

  profile="$profile_dir"/"$proto".remmina
  remmina --set-option username="$user" --update-profile "$profile"
  remmina --set-option password="$pw" --update-profile "$profile"
  remmina --set-option server="$host" --update-profile "$profile"

  # remmina --set-option window_maximize=1 --update-profile "$profile"
  # remmina --set-option scale=1 --update-profile "$profile"

  echo "Running remmina with URL $REMMINA_URL"
  exec remmina -c "$profile"
fi


echo "Running remmina without connection profile"
exec remmina
