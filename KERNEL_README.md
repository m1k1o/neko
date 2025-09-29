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
