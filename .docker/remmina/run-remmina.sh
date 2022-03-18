#!/bin/bash
set -u

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

[[ -z "$REMMINA_URL" ]] && err "Neither `REMMINA_PROFILE` nor `REMMINA_URL` found in env vars"

readarray -t arr < <( echo -n "$REMMINA_URL" | perl -pe 's|^(\w+\:\/\/)?(\w*:)?(.+@)?([^:]+)(:\d+)?$|\1\n\2\n\3\n\4\n\5|' )
proto=$(echo "${arr[0]}" | cut -d: -f1)
user=$(echo "${arr[1]}" | cut -d: -f1)
pw=$(echo "${arr[2]}" | cut -d@ -f1)
host="${arr[3]}"
# port=$(echo "${arr[4]}" | cut -d: -f2)
port="${arr[4]}" #keep the :
echo "Parsed url in `REMMINA_URL`: proto:$proto username:$user host:$host port:$port"

[[ "$proto" != "vnc" && "$proto" != "rdp" && "$proto" != "spice" ]] && err "Unsupported protocol $proto in connection url `REMMINA_URL`"

profile="$profile_dir"/"$proto".remmina
if [[ -n "$pw" ]]; then
  encpw=$(echo "$pw" | remmina --encrypt-password | grep Encrypted | sed 's/Encrypted password: //')
  remmina --set-option password="$encpw" --update-profile "$profile"
else
  remmina --set-option password= --update-profile "$profile"
fi
remmina --set-option username="$user" --update-profile "$profile"
remmina --set-option server="$host$port" --update-profile "$profile"

# remmina --set-option window_maximize=1 --update-profile "$profile"
# remmina --set-option scale=1 --update-profile "$profile"

exec remmina -c "$profile"
