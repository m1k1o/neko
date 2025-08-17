---
id: build-script
title: Build Script
---

The `build` script in the root of the repository is a shell script used to build the Docker images for n.eko. This document explains how to use the script and its various options.

## Usage

The basic usage of the script is:

```bash
./build [options] [image]
```

### Options

The script accepts several options to customize the build process:

-   `-p, --platform`: The platform to build the image for (e.g., `linux/amd64`, `linux/arm64`). Defaults to the system's architecture.
-   `-r, --repository`: The Docker repository prefix. Defaults to `ghcr.io/m1k1o/neko`.
-   `-t, --tag`: The image tag. This can be specified multiple times. If not specified, it defaults to `latest` and the current git semantic version tag (if available).
-   `-f, --flavor`: The image flavor (e.g., `nvidia`). If not specified, it builds without a flavor.
-   `-b, --base_image`: The name of the base image. Defaults to `<repository>/[<flavor>-]base:<tag>`.
-   `-a, --application`: The application to build. If not specified, it builds the base image.
-   `-y, --yes`: Skips confirmation prompts.
-   `--no-cache`: Builds the Docker image without using the cache.
-   `--push`: Pushes the image to the registry after a successful build.
-   `-h, --help`: Shows the help message.

### Positional Arguments

-   `<image>`: The full image name can be provided as a positional argument. The script will extract the repository, flavor, application, and tag from the image name. For example, `ghcr.io/m1k1o/neko/nvidia-firefox:latest`.

## Examples

### Build the base image

To build the base image for the local architecture with the tag `latest`:

```bash
./build
```

### Build a specific application

To build the Firefox application image:

```bash
./build --application firefox
```

### Build an application with a flavor

To build the Firefox application with the `nvidia` flavor:

```bash
./build --application firefox --flavor nvidia
```

### Build and push an image with a specific tag

To build the Google Chrome image with the tag `v1.0.0` and push it to the repository:

```bash
./build --application google-chrome --tag v1.0.0 --push
```
