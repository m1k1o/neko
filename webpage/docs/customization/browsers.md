---
sidebar_label: "Browsers"
description: "Customize your browser settings and configurations in Neko."
---

import { AppIcon } from '@site/src/components/AppIcon';
import { ProfileDirectoryPaths, PolicyFilePaths } from './browsers'

# Browsers Customization

Browsers use policies to manage settings and configurations programmatically. This is useful for Neko containers that can be set up with a specific configuration every time a fresh container is created. It also prevents users from changing certain settings, which is useful for security and privacy. For example, not allowing users to install extensions can prevent them from installing malicious extensions that could compromise their privacy and security.

However, as a user of Neko, you may want to customize your browser settings (install your own extensions) or wish to have your bookmarks and settings persist across sessions.

## Persistent Browser Profile {#persistent-profile}

When you run a browser in a container, the browser runs in a fresh environment every time you create (or upgrade) the container. This means that any changes you make to the browser settings or your browsing history will be lost when you stop the container. This is because the container is ephemeral and does not persist data across sessions.

If you want to persist your browser settings, bookmarks, extensions, and browsing history across sessions, you need to mount a volume to the browser's profile directory. This allows you to store your browser data outside of the container so it can be accessed even after the container is stopped or removed.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    # highlight-start
    volumes:
      - "./profile:/home/neko/.mozilla/firefox/profile.default"
    # highlight-end
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```

Replace `./profile` with the path to the directory where you want to store your browser data. This directory will be created if it does not exist.

Make sure to set the correct permissions for the directory so that the container can access it. The Neko user inside the container has a UID of `1000`, so you need to set the ownership of the directory to `1000:1000`. You can do this by running the following command:

```bash
sudo chown -R 1000:1000 ./profile
```

The path inside the container will be `/home/neko/.mozilla/firefox/profile.default`, which is the default profile directory for Firefox. You can find the profile directory for other browsers in the table below:

<ProfileDirectoryPaths />

## Browser Policy Files {#policy-files}

Browser policy files are JSON files that contain settings and configurations for the browser. These files are used to manage the browser settings programmatically and can be used to enforce certain policies, such as disabling extensions, setting the homepage, and more.

:::note
In the example below, we are using the Firefox browser. Make sure you replace `/usr/lib/firefox/distribution/policies.json` with the correct path for the browser you are using according to the tables below.
:::

If you want to customize the policy file, you can mount your own policy file to the container. This allows you to customize the browser settings and configurations to your liking.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    # highlight-start
    volumes:
      - "./policy.json:/usr/lib/firefox/distribution/policies.json"
    # highlight-end
    environment:
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```

If you just want to modify the default policy file, you can copy the current policy file from the container to your local machine:

```bash
# Create a container without starting it
docker create --name neko ghcr.io/m1k1o/neko/firefox:latest
# Copy the policy file from the container to your local machine
docker cp neko:/usr/lib/firefox/distribution/policies.json ./policy.json
# Remove the container
docker rm neko
```

Or you can download the default policy file from the repository directly:

```bash
# Replace firefox with the browser you are using
curl -o ./policy.json https://raw.githubusercontent.com/m1k1o/neko/refs/heads/main/apps/firefox/policies.json
```

If you wish to disable the policies altogether, you can just create an empty JSON file. This will disable all policies and allow you to customize the browser settings as you wish.

```json title="policy.json"
{}
```

### Firefox-based Browsers {#firefox-based}

The full configuration options for the Firefox-based policy JSON file can be found in the [Mozilla Policy Templates](https://mozilla.github.io/policy-templates/) documentation.

The policy files are located in the following paths:

<PolicyFilePaths flavors={['firefox-based']} />

**Allow persistent data in policies**

By default, the browsers in Neko are set up to forget all cookies and browsing history when they are closed. If you want to allow persistent data, you can set the following policies in the JSON file:

```json title="policy.json"
{
  "policies": {
    "SanitizeOnShutdown": false,
    "Homepage": {
      "StartPage": "previous-session"
    }
  }
}
```

**Manage extensions**

By default, the browsers in Neko do not allow installing extensions except for the ones that are pre-installed.

- `installation_mode`: The installation mode for the extension. It can be one of the following:
  - `allowed`: The extension can be installed by the user.
  - `blocked`: The extension cannot be installed by the user.
  - `force_installed`: The extension is installed automatically and cannot be removed by the user.

```json title="policy.json"
{
  "policies": {
    "ExtensionSettings": {
      "*": {
        "installation_mode": "blocked"
      },
      "sponsorBlocker@ajay.app": {
        "install_url": "https://addons.mozilla.org/firefox/downloads/latest/sponsorblock/latest.xpi",
        "installation_mode": "force_installed"
      },
      "uBlock0@raymondhill.net": {
        "install_url": "https://addons.mozilla.org/firefox/downloads/latest/ublock-origin/latest.xpi",
        "installation_mode": "force_installed"
      }
    }
  }
}
```

<details>
  <summary>How to find the extension ID?</summary>

  Extension IDs for Firefox are not available in the URL like in Chrome. You can find the extension ID by navigating to the `about:debugging#/runtime/this-firefox` page and clicking on the extension you want to install. The extension ID will be displayed in the URL.

  Another way is to find the extension on the [Official Add-ons Webpage](https://addons.mozilla.org/en-US/firefox/), then open DevTools (<code>F12</code>) and go to the `Console` tab. Enter the following command:

  ```javascript
  Object.keys(JSON.parse(document.getElementById('redux-store-state').textContent).addons.byGUID)[0]
  ```

  This will return the ID of the first extension on the page.
</details>

### Chromium-based Browsers {#chromium-based}

The full configuration options for the Chromium-based policy JSON file can be found in the [Chrome Enterprise](https://chromeenterprise.google/policies) documentation.

The policy files are located in the following paths:

<PolicyFilePaths flavors={['chromium-based']} />

**Allow file uploading & downloading**

By default, the browsers in Neko do not allow local file access. If you want to allow file uploading and downloading, you can set the following policies in the JSON file:

```json title="policy.json"
{
  "DownloadRestrictions": 0,
  "AllowFileSelectionDialogs": true,
  "URLAllowlist": [
    "file:///home/neko/Downloads"
  ]
}
```

**Allow persistent data in policies**

By default, the browsers in Neko are set up to forget all cookies and browsing history when they are closed. If you want to allow persistent data, you can set the following policies in the JSON file:

```json title="policy.json"
{
  "DefaultCookiesSetting": 1,
  "RestoreOnStartup": 1
}
```

**Manage extensions**

By default, the browsers in Neko do not allow installing extensions except for the ones that are pre-installed.

- `ExtensionInstallForcelist`: These extensions will be installed automatically and cannot be removed by the user.
- `ExtensionInstallAllowlist`: These extensions can be installed by the user when needed; they are not pre-installed.
- `ExtensionInstallBlocklist`: These extensions cannot be installed by the user, which is `*` (all extensions) by default.

The ID of the extension can be found in the URL of the extension in the Chrome Web Store. For example, the ID of the [uBlock Origin](https://chromewebstore.google.com/detail/ublock-origin/cjpalhdlnbpafiamejdnhcphjbkeiagm) extension is `cjpalhdlnbpafiamejdnhcphjbkeiagm`.

```json title="policy.json"
{
  "ExtensionInstallForcelist": [
    "cjpalhdlnbpafiamejdnhcphjbkeiagm;https://clients2.google.com/service/update2/crx",
    "mnjggcdmjocbbbhaepdhchncahnbgone;https://clients2.google.com/service/update2/crx"
  ],
  "ExtensionInstallAllowlist": [
    "cjpalhdlnbpafiamejdnhcphjbkeiagm",
    "mnjggcdmjocbbbhaepdhchncahnbgone"
  ],
  "ExtensionInstallBlocklist": [
    "*"
  ]
}
```
