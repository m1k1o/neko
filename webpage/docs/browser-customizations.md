---
sidebar_position: 4
---

# Browser Customizations

## 

- In order to **install own add-ons, set custom bookmarks** etc. you would need to modify the existing policy file and mount it to your container.
- For Firefox, copy [this](https://github.com/m1k1o/neko/blob/master/.docker/firefox/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/usr/lib/firefox/distribution/policies.json'`
- For Chromium, copy [this](https://github.com/m1k1o/neko/blob/master/.docker/chromium/policies.json) file, modify and mount it as: ` -v '${PWD}/policies.json:/etc/chromium/policies/managed/policies.json'`
- For others, see where existing `policies.json` is placed in their `Dockerfile`.

## Allow file uploading & downloading

- From security perspective, browser is not enabled to access local file data.
- If you want to enable this, you need to modify following policies:

```json title="policies.json"
{
  // ...
  "DownloadRestrictions": 0,
  "AllowFileSelectionDialogs": true,
  "URLAllowlist": [
      "file:///home/neko/Downloads"
  ],
  // ...
}
```

## Preserve browser data between restarts

- You need to mount browser profile as volume.
- For Firefox, that is this `/home/neko/.mozilla/firefox/profile.default` folder, mount it as: ` -v '${PWD}/data:/home/neko/.mozilla/firefox/profile.default'`
- For Chromium, that is this `/home/neko/.config/chromium` folder, mount it as: ` -v '${PWD}/data:/home/neko/.config/chromium'`
- For other chromium based browsers, see in `supervisord.conf` folder that is specified in `--user-data-dir`.

### Allow persistent data in policies

- From security perspective, browser is set up to forget all cookies and browsing history when its closed.
- If you want to enable this, you need to modify following policies:

```json title="policies.json"
{
  // ...
  "DefaultCookiesSetting": 1,
  "RestoreOnStartup": 1,
  // ...
}
```
