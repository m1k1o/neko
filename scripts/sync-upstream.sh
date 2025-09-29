#!/usr/bin/env bash
set -euo pipefail

# Usage/help
usage() {
    cat <<'EOF'
Usage: scripts/sync-upstream-manual.sh [options]

Options:
  -u, --upstream-branch BRANCH   Upstream branch to merge (default: master)
      --no-new-branch            Merge into current branch instead of creating a new one
  -h, --help                     Show this help and exit
EOF
}

# Defaults
UPSTREAM_BRANCH="master"
CREATE_NEW_BRANCH="true"

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        -u|--upstream-branch)
            if [[ $# -lt 2 ]]; then
                echo "Missing value for $1" >&2
                usage
                exit 1
            fi
            UPSTREAM_BRANCH="$2"
            shift 2
            ;;
        --no-new-branch)
            CREATE_NEW_BRANCH="false"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage
            exit 1
            ;;
    esac
done

# Check if upstream remote exists, add if not
if ! git remote get-url upstream >/dev/null 2>&1; then
    echo "Adding upstream remote..."
    git remote add upstream https://github.com/m1k1o/neko.git
else
    echo "Upstream remote already exists"
fi

if [[ "${CREATE_NEW_BRANCH}" == "true" ]]; then
    branch="sync/upstream-$(date +%Y%m%d)"
    git checkout -b "$branch"  # work on a throw-away branch
else
    echo "Merging into existing branch: $(git rev-parse --abbrev-ref HEAD)"
fi

# merge selected upstream branch into this branch
git fetch upstream
git merge --no-edit "upstream/${UPSTREAM_BRANCH}" || true
echo "there are now likely conflicts to resolve. Go forth and resolve them!"

# For each stop: 
#  - run git status to see the conflicted files
#  - open each file, remove the conflict markers (<<<<<<<, =======, >>>>>>>) and keep the version you want,
# OR, if you simply want the upstream copy: git checkout --theirs path/to/file.
# If you want to keep your fork's version: git checkout --ours path/to/file.
#  - git add <file> (or git rm if you really mean to delete it).
#  - git merge --continue.


# When git merge is completed, push the branch back to GitHub:
#    git push --force-with-lease -u origin HEAD
