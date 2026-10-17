#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: utils/release.sh <major|minor|patch> [--push]

Bumps the client and core package versions together, commits the release, and
creates an annotated vX.Y.Z tag. The tag is shared by the backend and client
release workflows. By default, prints the command to push the branch and tag.
Pass --push to push both atomically to origin.
EOF
}

if [[ ${1:-} == --help || ${1:-} == -h ]]; then
  usage
  exit 0
fi

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage >&2
  exit 2
fi

bump=$1
case "$bump" in
  major|minor|patch) ;;
  *)
    usage >&2
    exit 2
    ;;
esac

push=0
if [[ $# -eq 2 ]]; then
  if [[ $2 != --push ]]; then
    usage >&2
    exit 2
  fi
  push=1
fi

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
root=$(git -C "$script_dir" rev-parse --show-toplevel 2>/dev/null) || {
  printf 'Error: run this script from inside a git repository.\n' >&2
  exit 1
}
cd "$root"

if [[ -n $(git status --porcelain) ]]; then
  printf 'Error: working tree must be clean before releasing.\n' >&2
  exit 1
fi

branch=$(git branch --show-current)
if [[ -z $branch ]]; then
  printf 'Error: release from a branch, not a detached HEAD.\n' >&2
  exit 1
fi

if [[ $push -eq 1 ]] && ! git remote get-url origin >/dev/null 2>&1; then
  printf 'Error: remote "origin" is not configured.\n' >&2
  exit 1
fi

client_version=$(node -p "require('./client/package.json').version")
core_version=$(node -p "require('./client/core/package.json').version")
if [[ $client_version != "$core_version" ]]; then
  printf 'Error: client version (%s) and core version (%s) differ.\n' "$client_version" "$core_version" >&2
  exit 1
fi

next_version=$(node - "$client_version" "$bump" <<'NODE'
const [, , version, bump] = process.argv
const match = /^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/.exec(version)
if (!match) throw new Error(`Expected a stable semver version, got ${version}`)

let [major, minor, patch] = match.slice(1).map(BigInt)
if (bump === 'major') [major, minor, patch] = [major + 1n, 0n, 0n]
if (bump === 'minor') [minor, patch] = [minor + 1n, 0n]
if (bump === 'patch') patch += 1n
console.log(`${major}.${minor}.${patch}`)
NODE
)
tag="v$next_version"

if git rev-parse --verify --quiet "refs/tags/$tag" >/dev/null; then
  printf 'Error: tag %s already exists locally.\n' "$tag" >&2
  exit 1
fi

(cd "$root/client" && npm version "$next_version" --no-git-tag-version)
(cd "$root/client/core" && npm version "$next_version" --no-git-tag-version)

git add -- \
  client/package.json \
  client/package-lock.json \
  client/core/package.json \
  client/core/package-lock.json
git commit -m "release: $tag"
git tag --annotate "$tag" --message "Release $tag"

push_command=(git push --atomic origin "HEAD:refs/heads/$branch" "refs/tags/$tag")
if [[ $push -eq 1 ]]; then
  "${push_command[@]}"
else
  printf '\nRelease %s created locally. To push the release commit and tag, run:\n  ' "$tag"
  printf '%q ' "${push_command[@]}"
  printf '\n'
fi
