# neko fork for onkernel

This is a public fork of the [m1k1o/neko](https://github.com/m1k1o/neko) repository.

## Overview

We maintain this fork to provide a customized `base` image that is used in our browser images at [onkernel/kernel-images](https://github.com/onkernel/kernel-images/tree/main/images).

## Building

To build images from this fork, use the build script from the repository root:

```bash
# Build the base image
./build base

# Build with custom repository and tag
./build base --repository your-repo/neko --tag custom-tag
```

The `--repository` and `--tag` options allow you to specify exactly which image you're building, making it easy to reference back to specific builds in `kernel-images`.

## Keeping in sync with upstream

To merge the latest changes from the upstream neko repository:

```bash
# Run the sync script to create a new branch and merge upstream changes
./scripts/sync-upstream.sh

# Or merge directly into your current branch
./scripts/sync-upstream.sh --no-new-branch

# To merge a specific upstream branch
./scripts/sync-upstream.sh --upstream-branch $branch
```

After running the sync script:

1. Resolve any merge conflicts
2. Test the build to ensure compatibility
3. Push the changes and create a PR for review

## Releasing New Images

We have github actions that will build and publish images to [ghcr](./.github/workflows/ghcr.yml)

### Picking a tag

Tags are structured as `v$UPSTREAM-v$INTERNAL` (e.g. `v3.0.6-v1.0.1`)

- If you've sync'ed the upstream use the [latest upstream tag](https://github.com/m1k1o/neko/tags) and update the `upstream` tag. (using our example, `v3.0.7-v1.0.1`)
- Otherwise use the latest [internal tag](https://github.com/onkernel/neko/tags) and bump the `internal` tag following semantic versioning (e.g. `v3.0.6-v1.2.0`)

### Tag and push

```bash
git tag $TAG
git push origin $TAG
```
