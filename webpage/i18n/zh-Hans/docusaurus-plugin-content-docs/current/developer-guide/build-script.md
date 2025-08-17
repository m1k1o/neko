---
id: build-script
title: 构建脚本
---

仓库根目录中的 `build` 脚本是一个 shell 脚本，用于构建 n.eko 的 Docker 镜像。本文档介绍了如何使用该脚本及其各种选项。

## 用法

该脚本的基本用法是：

```bash
./build [选项] [镜像]
```

### 选项

该脚本接受几个选项来自定义构建过程：

-   `-p, --platform`: 构建镜像的目标平台（例如，`linux/amd64`，`linux/arm64`）。默认为系统的体系结构。
-   `-r, --repository`: Docker 仓库前缀。默认为 `ghcr.io/m1k1o/neko`。
-   `-t, --tag`: 镜像标签。可以多次指定。如果未指定，则默认为 `latest` 和当前的 git 语义版本标签（如果可用）。
-   `-f, --flavor`: 镜像风格（例如，`nvidia`）。如果未指定，则构建时没有风格。
-   `-b, --base_image`: 基础镜像的名称。默认为 `<repository>/[<flavor>-]base:<tag>`。
-   `-a, --application`: 要构建的应用程序。如果未指定，则构建基础镜像。
-   `-y, --yes`: 跳过确认提示。
-   `--no-cache`: 构建 Docker 镜像时不使用缓存。
-   `--push`: 成功构建后将镜像推送到仓库。
-   `-h, --help`: 显示帮助信息。

### 位置参数

-   `<image>`: 完整的镜像名称可以作为位置参数提供。脚本将从镜像名称中提取仓库、风格、应用程序和标签。例如，`ghcr.io/m1k1o/neko/nvidia-firefox:latest`。

## 示例

### 构建基础镜像

为本地体系结构构建标签为 `latest` 的基础镜像：

```bash
./build
```

### 构建特定的应用程序

构建 Firefox 应用程序镜像：

```bash
./build --application firefox
```

### 构建带有风格的应用程序

构建带有 `nvidia` 风格的 Firefox 应用程序：

```bash
./build --application firefox --flavor nvidia
```

### 构建并推送带有特定标签的镜像

构建标签为 `v1.0.0` 的 Google Chrome 镜像并将其推送到仓库：

```bash
./build --application google-chrome --tag v1.0.0 --push
```
