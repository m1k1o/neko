# Custom n.eko Images

This directory contains examples of how to build your own custom n.eko images with your own modifications and browser extensions.

## Directory Structure

-   `base/`: Contains a `Dockerfile` for a custom base image. You can add your own software here that will be available in all browser images.
-   `firefox/`: Contains a `Dockerfile` for a custom Firefox image.
-   `google-chrome/`: Contains a `Dockerfile` for a custom Google Chrome image.

## How to Build

You can use the `build` script in the root of the repository to build your custom images.

### 1. Build the Custom Base Image

First, you need to build the custom base image. This image will be used as the base for the browser images.

```bash
./build --application base --repository "custom" --tag "latest" -f ""
```

**Note:** The `--repository "custom"` argument tells the build script to look for the `base` application in the `custom` directory. We are not using a flavor, so we pass `-f ""` to override any default flavor.

### 2. Build the Custom Browser Images

After building the base image, you can build the browser images.

#### Firefox

```bash
./build --application firefox --repository "custom" --base_image "custom/base:latest" --tag "latest" -f ""
```

#### Google Chrome

```bash
./build --application google-chrome --repository "custom" --base_image "custom/base:latest" --tag "latest" -f ""
```

**Note:** The `--base_image "custom/base:latest"` argument tells the build script to use your custom base image.

## How to Add Browser Extensions

To add your own browser extensions, you need to modify the `Dockerfile` for the respective browser.

### Firefox

1.  Open `custom/firefox/Dockerfile`.
2.  Uncomment the lines for installing an extension.
3.  Replace `<URL to the .xpi file>` with the URL of the extension's `.xpi` file.

### Google Chrome

1.  Open `custom/google-chrome/Dockerfile`.
2.  Uncomment the lines for installing an extension.
3.  Replace `<extension_id>` with the ID of the extension from the Chrome Web Store.

After modifying the `Dockerfile`, rebuild the browser image using the commands from the previous section.
