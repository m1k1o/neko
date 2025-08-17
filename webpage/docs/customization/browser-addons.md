---
id: browser-addons
title: Installing Browser Add-ons
---

You can customize the n.eko browser images by installing your favorite browser add-ons (extensions). This can be done by modifying the `Dockerfile` for the respective browser.

## Chromium-based Browsers (Google Chrome, Chromium, Brave, etc.)

For Chromium-based browsers, you can install extensions by using a policy file. This method tells the browser to install the extension from the Chrome Web Store on its first run.

1.  **Find the Extension ID:**
    Go to the Chrome Web Store and find the extension you want to install. The URL will look something like this: `https://chrome.google.com/webstore/detail/ublock-origin/cjpalhdlnbpafiamejdnhcphjbkeiagm`. The last part of the URL (`cjpalhdlnbpafiamejdnhcphjbkeiagm`) is the extension ID.

2.  **Modify the Dockerfile:**
    In the `Dockerfile` for your chosen browser (e.g., `apps/google-chrome/Dockerfile`), add the following lines:

    ```dockerfile
    # Create the directory for the policy file
    RUN mkdir -p /etc/opt/chrome/policies/managed/

    # Create the policy file to install the uBlock Origin extension
    RUN echo '{ "ExtensionInstallForcelist": [ "cjpalhdlnbpafiamejdnhcphjbkeiagm;https://clients2.google.com/service/update2/crx" ] }' > /etc/opt/chrome/policies/managed/policies.json
    ```

    You can add multiple extensions to the `ExtensionInstallForcelist` array.

    **Note:** The path `/etc/opt/chrome/policies/managed/` is for Google Chrome. For other Chromium-based browsers, the path might be different. For example, for Brave it is `/etc/brave/policies/managed/`.

## Firefox

For Firefox, you can install extensions by downloading the `.xpi` file and using a command to install it globally.

1.  **Find the Add-on URL:**
    Go to the Firefox Add-ons website and find the add-on you want to install. Right-click the "Add to Firefox" button and copy the link address. The URL will end with `.xpi`.

2.  **Modify the Dockerfile:**
    In the `apps/firefox/Dockerfile`, add the following lines. This example installs the uBlock Origin add-on.

    ```dockerfile
    # URL of the uBlock Origin .xpi file
    ARG UBLOCK_URL="https://addons.mozilla.org/firefox/downloads/file/4232420/ublock_origin-1.55.0-an-fx.xpi"

    # Download and install the extension
    RUN wget -O /tmp/ublock.xpi "${UBLOCK_URL}" && \
        firefox -install-global-extension /tmp/ublock.xpi && \
        rm /tmp/ublock.xpi
    ```

    **Note:** The URL for the `.xpi` file might change with new versions of the add-on. You should always check for the latest URL.

After modifying the `Dockerfile`, you can build the image using the `build` script. The new image will have the specified add-ons installed.
