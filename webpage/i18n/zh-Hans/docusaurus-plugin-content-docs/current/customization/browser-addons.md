---
id: browser-addons
title: 安装浏览器插件
---

您可以通过安装您喜欢的浏览器插件（扩展）来自定义 n.eko 浏览器镜像。这可以通过修改相应浏览器的 `Dockerfile` 来完成。

## 基于 Chromium 的浏览器（Google Chrome、Chromium、Brave 等）

对于基于 Chromium 的浏览器，您可以使用策略文件来安装扩展。此方法会告知浏览器在首次运行时从 Chrome 网上应用店安装扩展。

1.  **查找扩展 ID：**
    转到 Chrome 网上应用店，找到您要安装的扩展。URL 类似于：`https://chrome.google.com/webstore/detail/ublock-origin/cjpalhdlnbpafiamejdnhcphjbkeiagm`。URL 的最后一部分（`cjpalhdlnbpafiamejdnhcphjbkeiagm`）就是扩展 ID。

2.  **修改 Dockerfile：**
    在您选择的浏览器的 `Dockerfile` 中（例如 `apps/google-chrome/Dockerfile`），添加以下行：

    ```dockerfile
    # 创建策略文件目录
    RUN mkdir -p /etc/opt/chrome/policies/managed/

    # 创建策略文件以安装 uBlock Origin 扩展
    RUN echo '{ "ExtensionInstallForcelist": [ "cjpalhdlnbpafiamejdnhcphjbkeiagm;https://clients2.google.com/service/update2/crx" ] }' > /etc/opt/chrome/policies/managed/policies.json
    ```

    您可以将多个扩展添加到 `ExtensionInstallForcelist` 数组中。

    **注意：** 路径 `/etc/opt/chrome/policies/managed/` 适用于 Google Chrome。对于其他基于 Chromium 的浏览器，路径可能不同。例如，对于 Brave，路径是 `/etc/brave/policies/managed/`。

## Firefox

对于 Firefox，您可以通过下载 `.xpi` 文件并使用命令进行全局安装来安装扩展。

1.  **查找插件 URL：**
    转到 Firefox 附加组件网站，找到您要安装的插件。右键单击“添加到 Firefox”按钮并复制链接地址。URL 将以 `.xpi` 结尾。

2.  **修改 Dockerfile：**
    在 `apps/firefox/Dockerfile` 中，添加以下行。此示例安装 uBlock Origin 插件。

    ```dockerfile
    # uBlock Origin .xpi 文件的 URL
    ARG UBLOCK_URL="https://addons.mozilla.org/firefox/downloads/file/4232420/ublock_origin-1.55.0-an-fx.xpi"

    # 下载并安装扩展
    RUN wget -O /tmp/ublock.xpi "${UBLOCK_URL}" && \
        firefox -install-global-extension /tmp/ublock.xpi && \
        rm /tmp/ublock.xpi
    ```

    **注意：** `.xpi` 文件的 URL 可能会随着插件的新版本而更改。您应该始终检查最新的 URL。

修改 `Dockerfile` 后，您可以使用 `build` 脚本来构建镜像。新镜像将安装指定的插件。
