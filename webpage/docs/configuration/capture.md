---
description: Configuration related to Gstreamer capture in Neko.
---

import { Def, Opt } from '@site/src/components/Anchor';
import { ConfigurationTab } from '@site/src/components/Configuration';
import configOptions from './help.json';

# Audio & Video Capture

This guide will show you how to configure the audio and video capture settings in neko.

Neko uses [Gstreamer](https://gstreamer.freedesktop.org/) to capture and encode audio and video in the following scenarios:

- WebRTC clients use the [Video](#video) and [Audio](#audio) pipelines to receive the audio and video streams from the server.
- The [Broadcast](#broadcast) feature allows you to broadcast the audio and video to a third-party service using RTMP.
- The WebRTC Fallback mechanism allows you to capture the display in the form of JPEG images and serve them over HTTP using [Screencast](#screencast).
- Clients can share their [Webcam](#webcam) and [Microphone](#microphone) with the server using WebRTC.

## WebRTC Video {#video}

Neko allows you to capture the display and encode it in real-time using Gstreamer. The encoded video is then sent to the client using WebRTC. This allows you to share the display with the client in real-time.

There can exist multiple video pipelines in neko that are referenced by their unique pipeline id. Each video pipeline can have its own configuration settings and clients can either choose which pipeline they want to use or let neko choose the best pipeline for them.

:::info Limitation
All video pipelines must use the same video codec (defined in the <Opt id="video.codec" /> setting).
:::

The Gstreamer pipeline is started when the first client requests the video stream and is stopped after the last client disconnects.

<ConfigurationTab options={configOptions} filter={[
  "capture.video.display",
  "capture.video.codec",
  "capture.video.ids",
  "capture.video.pipeline",
  "capture.video.pipelines",
]} comments={false} />

- <Def id="video.display" /> is the name of the [X display](https://www.x.org/wiki/) that you want to capture. If not specified, the environment variable `DISPLAY` will be used.
- <Def id="video.codec" /> available codecs are `vp8`, `vp9`, `av1`, `h264`. [Supported video codecs](https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/WebRTC_codecs#supported_video_codecs) are dependent on the WebRTC implementation used by the client, `vp8` and `h264` are supported by all WebRTC implementations.
- <Def id="video.ids" /> is a list of pipeline ids that are defined in the <Opt id="video.pipelines" /> section. The first pipeline in the list will be the default pipeline.
- <Def id="video.pipeline" /> is a shorthand for defining [Gstreamer pipeline description](#video.gst_pipeline) for a single pipeline. This is option is ignored if <Opt id="video.pipelines" /> is defined.
- <Def id="video.pipelines" /> is a dictionary of pipeline configurations. Each pipeline configuration is defined by a unique pipeline id. They can be defined in two ways: either by building the pipeline dynamically using [Expression-Driven Configuration](#video.expression) or by defining the pipeline using a [Gstreamer Pipeline Description](#video.gst_pipeline).

### Expression-Driven Configuration {#video.expression}

Expression allows you to build the pipeline dynamically based on the current resolution and framerate of the display. Expressions are evaluated using the [gval](https://github.com/PaesslerAG/gval) library. Available variables are <Opt id="video.pipelines.width" />, <Opt id="video.pipelines.height" />, and <Opt id="video.pipelines.fps" /> of the display at the time of capture.

```yaml title="config.yaml"
capture:
  video:
    ...
    pipelines:
      <pipeline_id>:
        width: "<expression>"
        height: "<expression>"
        fps: "<expression>"
        gst_prefix: "<gst_pipeline>"
        gst_encoder: "<gst_encoder_name>"
        gst_params:
          <param_name>: "<expression>"
        gst_suffix: "<gst_pipeline>"
        show_pointer: true
```

- <Def id="video.pipelines.width" />, <Def id="video.pipelines.height" />, and <Def id="video.pipelines.fps" /> are the expressions that are evaluated to get the stream resolution and framerate. They can be different from the display resolution and framerate if downscaling or upscaling is desired.
- <Def id="video.pipelines.gst_prefix" /> and <Def id="video.pipelines.gst_suffix" /> allow you to add custom Gstreamer elements before and after the encoder. Both parameters need to start with `!` and then be followed by the Gstreamer elements.
- <Def id="video.pipelines.gst_encoder" /> is the name of the Gstreamer encoder element, such as `vp8enc` or `x264enc`.
- <Def id="video.pipelines.gst_params" /> are the parameters that are passed to the encoder element specified in <Opt id="video.pipelines.gst_encoder" />.
- <Def id="video.pipelines.show_pointer" /> is a boolean value that determines whether the mouse pointer should be captured or not.

<details>
  <summary>Example pipeline configuration</summary>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="vp8" label="VP8 configuration">

    ```yaml title="config.yaml"
    capture:
      video:
        codec: vp8
        # HQ is the default pipeline
        ids: [ hq, lq ]
        pipelines:
          hq:
            fps: 25
            gst_encoder: vp8enc
            gst_params:
              target-bitrate: round(3072 * 650)
              cpu-used: 4
              end-usage: cbr
              threads: 4
              deadline: 1
              undershoot: 95
              buffer-size: (3072 * 4)
              buffer-initial-size: (3072 * 2)
              buffer-optimal-size: (3072 * 3)
              keyframe-max-dist: 25
              min-quantizer: 4
              max-quantizer: 20
          lq:
            fps: 25
            gst_encoder: vp8enc
            gst_params:
              target-bitrate: round(1024 * 650)
              cpu-used: 4
              end-usage: cbr
              threads: 4
              deadline: 1
              undershoot: 95
              buffer-size: (1024 * 4)
              buffer-initial-size: (1024 * 2)
              buffer-optimal-size: (1024 * 3)
              keyframe-max-dist: 25
              min-quantizer: 4
              max-quantizer: 20
    ```

  </TabItem>
  <TabItem value="h264" label="H264 configuration">

    ```yaml title="config.yaml"
    capture:
      video:
        codec: h264
        ids: [ main ]
        pipelines:
          main:
            width: (width / 3) * 2
            height: (height / 3) * 2
            fps: 20
            gst_prefix: "! video/x-raw,format=I420"
            gst_encoder: "x264enc"
            gst_params:
              threads: 4
              bitrate: 4096
              key-int-max: 15
              byte-stream: true
              tune: zerolatency
              speed-preset: veryfast 
            gst_suffix: "! video/x-h264,stream-format=byte-stream"
    ```

  </TabItem>
</Tabs>

</details>

### Gstreamer Pipeline Description {#video.gst_pipeline}

If you want to define the pipeline using a [Gstreamer pipeline description](https://gstreamer.freedesktop.org/documentation/tools/gst-launch.html?gi-language=c#pipeline-description), you can do so by setting the <Def id="video.pipelines.gst_pipeline" /> parameter.

```yaml title="config.yaml"
capture:
  video:
    ...
    pipelines:
      <pipeline_id>:
        gst_pipeline: "<gstreamer_pipeline>"
```

Since now you have to define the whole pipeline, you need to specify the src element to get the video frames and the sink element to send the encoded video frames to neko. In your pipeline, you can use `{display}` as a placeholder for the display name that will be replaced by the actual display name at runtime. You need to set the `name` property of the sink element to `appsink` so that neko can capture the video frames.

Your typical pipeline string would look like this:

```
ximagesrc display-name={display} show-pointer=true use-damage=false ! <your_elements> ! appsink name=appsink
```

See documentation for [ximagesrc](https://gstreamer.freedesktop.org/documentation/ximagesrc/index.html) and [appsink](https://gstreamer.freedesktop.org/documentation/app/appsink.html) for more information.

<details>
  <summary>Example pipeline configuration</summary>

<Tabs>
  <TabItem value="vp8" label="VP8 configuration">

    ```yaml title="config.yaml"
    capture:
      video:
        codec: vp8
        ids: [ hq, lq ]
        pipelines:
          hq:
            gst_pipeline: |
              ximagesrc display-name={display} show-pointer=true use-damage=false
              ! videoconvert ! queue
              ! vp8enc
                name=encoder
                target-bitrate=3072000
                cpu-used=4
                end-usage=cbr
                threads=4
                deadline=1
                undershoot=95
                buffer-size=12288
                buffer-initial-size=6144
                buffer-optimal-size=9216
                keyframe-max-dist=25
                min-quantizer=4
                max-quantizer=20
              ! appsink name=appsink
          lq:
            gst_pipeline: |
              ximagesrc display-name={display} show-pointer=true use-damage=false
              ! videoconvert ! queue
              ! vp8enc
                name=encoder
                target-bitrate=1024000
                cpu-used=4
                end-usage=cbr
                threads=4
                deadline=1
                undershoot=95
                buffer-size=4096
                buffer-initial-size=2048
                buffer-optimal-size=3072
                keyframe-max-dist=25
                min-quantizer=4
                max-quantizer=20
              ! appsink name=appsink
    ```

  </TabItem>
  <TabItem value="h264" label="H264 configuration">

    ```yaml title="config.yaml"
    capture:
      video:
        codec: h264
        ids: [ main ]
        pipelines:
          main:
            gst_pipeline: |
              ximagesrc display-name={display} show-pointer=true use-damage=false
              ! videoconvert ! queue
              ! x264enc
                name=encoder
                threads=4
                bitrate=4096
                key-int-max=15
                byte-stream=true
                tune=zerolatency
                speed-preset=veryfast
              ! video/x-h264,stream-format=byte-stream
              ! appsink name=appsink
    ```
  </TabItem>
  <TabItem value="nvh264enc" label="NVENC H264 configuration">

    ```yaml title="config.yaml"
    capture:
      video:
        codec: h264
        ids: [ main ]
        pipelines:
          main:
            gst_pipeline: |
              ximagesrc display-name={display} show-pointer=true use-damage=false
              ! videoconvert ! queue
              ! video/x-raw,format=NV12
              ! nvh264enc
                name=encoder
                preset=2
                gop-size=25
                spatial-aq=true
                temporal-aq=true
                bitrate=4096
                vbv-buffer-size=4096
                rc-mode=6
              ! h264parse config-interval=-1
              ! video/x-h264,stream-format=byte-stream
              ! appsink name=appsink
    ```

    This configuration requires [Nvidia GPU](https://developer.nvidia.com/cuda-gpus) with [NVENC](https://developer.nvidia.com/nvidia-video-codec-sdk) support.

    ```yaml title="config.yaml"
    capture:
      video:
        codec: h264
        ids: [ main ]
        pipelines:
          main:
            gst_pipeline: |
              ximagesrc display-name={display} show-pointer=true use-damage=false
              ! cudaupload ! cudaconvert ! queue
              ! video/x-raw(memory:CUDAMemory),format=NV12
              ! nvh264enc
                name=encoder
                preset=2
                gop-size=25
                spatial-aq=true
                temporal-aq=true
                bitrate=4096
                vbv-buffer-size=4096
                rc-mode=6
              ! h264parse config-interval=-1
              ! video/x-h264,stream-format=byte-stream
              ! appsink name=appsink
    ```

    This configuration requires [Nvidia GPU](https://developer.nvidia.com/cuda-gpus) with [NVENC](https://developer.nvidia.com/nvidia-video-codec-sdk) support and [Cuda](https://developer.nvidia.com/cuda-zone) support.

  </TabItem>
</Tabs>

</details>

Overview of available encoders for each codec is shown in the table below. The encoder name is used in the <Opt id="video.pipelines.gst_encoder" /> parameter. The parameters for each encoder are different and you can find the documentation for each encoder in the links below.

| codec | encoder | vaapi encoder | nvenc encoder |
| ----- | ------- | ------------- | ------------- |
| VP8   | [vp8enc](https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html?gi-language=c) | [vaapivp8enc](https://github.com/GStreamer/gstreamer-vaapi/blob/master/gst/vaapi/gstvaapiencode_vp8.c) | ? |
| VP9   | [vp9enc](https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html?gi-language=c) | [vaapivp9enc](https://github.com/GStreamer/gstreamer-vaapi/blob/master/gst/vaapi/gstvaapiencode_vp9.c) | ? |
| AV1   | [av1enc](https://gstreamer.freedesktop.org/documentation/aom/av1enc.html?gi-language=c) | ? | [nvav1enc](https://gstreamer.freedesktop.org/documentation/nvcodec/nvav1enc.html?gi-language=c) |
| H264  | [x264enc](https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c) | [vaapih264enc](https://gstreamer.freedesktop.org/documentation/vaapi/vaapih264enc.html?gi-language=c) | [nvh264enc](https://gstreamer.freedesktop.org/documentation/nvcodec/nvh264enc.html?gi-language=c) |
| H265  | [x265enc](https://gstreamer.freedesktop.org/documentation/x265/index.html?gi-language=c) | [vaapih265enc](https://gstreamer.freedesktop.org/documentation/vaapi/vaapih265enc.html?gi-language=c) | [nvh265enc](https://gstreamer.freedesktop.org/documentation/nvcodec/nvh265enc.html?gi-language=c) |


## WebRTC Audio {#audio}

Only one audio pipeline can be defined in neko. The audio pipeline is used to capture and encode audio, similar to the video pipeline. The encoded audio is then sent to the client using WebRTC.

The Gstreamer pipeline is started when the first client requests the video stream and is stopped after the last client disconnects.

<ConfigurationTab options={configOptions} filter={[
  "capture.audio.device",
  "capture.audio.codec",
  "capture.audio.pipeline",
]} comments={false} />

- <Def id="audio.device" /> is the name of the [pulseaudio device](https://wiki.archlinux.org/title/PulseAudio/Examples) that you want to capture. If not specified, the default audio device will be used.
- <Def id="audio.codec" /> available codecs are `opus`, `g722`, `pcmu`, `pcma`. [Supported audio codecs](https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/WebRTC_codecs#supported_audio_codecs) are dependent on the WebRTC implementation used by the client, `opus` is supported by all WebRTC implementations.
- <Def id="audio.pipeline" /> is the Gstreamer pipeline description that is used to capture and encode audio. You can use `{device}` as a placeholder for the audio device name that will be replaced by the actual device name at runtime.

<details>
  <summary>Example pipeline configuration</summary>

```yaml title="config.yaml"
capture:
  audio:
    codec: opus
    pipeline: |
      pulsesrc device={device}
      ! audioconvert
      ! opusenc
        bitrate=320000
      ! appsink name=appsink
```

</details>

## Broadcast {#broadcast}

Neko allows you to broadcast out-of-the-box the display and audio capture to a third-party service. This can be used to broadcast the display and audio to a streaming service like [Twitch](https://www.twitch.tv/) or [YouTube](https://www.youtube.com/), or to a custom RTMP server like [OBS](https://obsproject.com/), [Nginx RTMP module](https://github.com/arut/nginx-rtmp-module), or [MediaMTX](https://github.com/bluenviron/mediamtx).

The Gstreamer pipeline is started when the broadcast is started and is stopped when the broadcast is stopped regardless of the clients connected.

<ConfigurationTab options={configOptions} filter={[
  "capture.broadcast.audio_bitrate",
  "capture.broadcast.video_bitrate",
  "capture.broadcast.preset",
  "capture.broadcast.pipeline",
  "capture.broadcast.url",
  "capture.broadcast.autostart",
]} comments={false} />

The default encoder uses `h264` for video and `aac` for audio, muxed in the `flv` container and sent over the `rtmp` protocol. You can change the encoder settings by setting a custom Gstreamer pipeline description in the <Opt id="broadcast.pipeline" /> parameter.

- <Def id="broadcast.audio_bitrate" /> and <Def id="broadcast.video_bitrate" /> are the bitrate settings for the default audio and video encoders expressed in kilobits per second.
- <Def id="broadcast.preset" /> is the encoding speed preset for the default video encoder. See available presets [here](https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c#GstX264EncPreset).
- <Def id="broadcast.pipeline" /> when set, encoder settings above are ignored and the custom Gstreamer pipeline description is used. In the pipeline, you can use `{hostname}`, `{display}`, `{device}` and `{url}` as placeholders for the X display name, pulseaudio audio device name, and broadcast URL respectively.
- <Def id="broadcast.url" /> is the URL of the RTMP server where the broadcast will be sent e.g. `rtmp://<server>/<application>/<stream_key>`. This can be set later using the API if the URL is not known at the time of configuration or is expected to change.
- <Def id="broadcast.autostart" /> is a boolean value that determines whether the broadcast should start automatically when neko starts, works only if the URL is set.

<details>
  <summary>Example pipeline configuration</summary>

<Tabs>
  <TabItem value="x264" label="X264 configuration">

    ```yaml title="config.yaml"
    capture:
      broadcast:
        pipeline: |
          flvmux name=mux
            ! rtmpsink location={url}
          pulsesrc device={device}
            ! audio/x-raw,channels=2
            ! audioconvert
            ! voaacenc
            ! mux.
          ximagesrc display-name={display} show-pointer=false use-damage=false
            ! video/x-raw,framerate=28/1
            ! videoconvert
            ! queue
            ! x264enc bframes=0 key-int-max=0 byte-stream=true tune=zerolatency speed-preset=veryfast
            ! mux.
    ```

  </TabItem>
  <TabItem value="nvh264enc" label="NVENC H264 configuration">

    ```yaml title="config.yaml"
    capture:
      broadcast:
        pipeline: |
          flvmux name=mux
            ! rtmpsink location={url}
          pulsesrc device={device}
            ! audio/x-raw,channels=2
            ! audioconvert
            ! voaacenc
            ! mux.
          ximagesrc display-name={display} show-pointer=false use-damage=false
            ! video/x-raw,framerate=30/1
            ! videoconvert
            ! queue
            ! video/x-raw,format=NV12
            ! nvh264enc name=encoder preset=low-latency-hq gop-size=25 spatial-aq=true temporal-aq=true bitrate=2800 vbv-buffer-size=2800 rc-mode=6
            ! h264parse config-interval=-1
            ! video/x-h264,stream-format=byte-stream,profile=high
            ! h264parse
            ! mux.
    ```

    This configuration requires [Nvidia GPU](https://developer.nvidia.com/cuda-gpus) with [NVENC](https://developer.nvidia.com/nvidia-video-codec-sdk) support and [Nvidia docker image](/docs/v3/installation/docker-images#nvidia) of neko.

  </TabItem>
</Tabs>

</details>

## Screencast {#screencast}

As a fallback mechanism, neko can capture the display in the form of JPEG images and the client can request these images over HTTP. This is useful when the client does not support WebRTC or when the client is not able to establish a WebRTC connection, or there is a temporary issue with the WebRTC connection and the client should not miss the content being shared.

:::note
This is a fallback mechanism and should not be used as a primary video stream because of the high latency, low quality, and high bandwidth requirements.
:::

The Gstreamer pipeline is started in the background when the first client requests the screencast and is stopped after a period of inactivity.

<ConfigurationTab options={configOptions} filter={[
  "capture.screencast.enabled",
  "capture.screencast.rate",
  "capture.screencast.quality",
  "capture.screencast.pipeline",
]} comments={false} />

- <Def id="screencast.enabled" /> is a boolean value that determines whether the screencast is enabled or not.
- <Def id="screencast.rate" /> is the framerate of the screencast. It is expressed as a fraction of frames per second, for example, `10/1` means 10 frames per second.
- <Def id="screencast.quality" /> is the quality of the JPEG images. It is expressed as a percentage, for example, `60` means 60% quality.
- <Def id="screencast.pipeline" /> when set, the default pipeline settings above are ignored and the custom Gstreamer pipeline description is used. In the pipeline, you can use `{display}` as a placeholder for the X display name.

<details>
  <summary>Example pipeline configuration</summary>

```yaml title="config.yaml"
capture:
  screencast:
    enabled: true
    pipeline: |
      ximagesrc display-name={display} show-pointer=true use-damage=false
        ! video/x-raw,framerate=10/1
        ! videoconvert
        ! queue
        ! jpegenc quality=60
        ! appsink name=appsink
```

</details>

## Webcam {#webcam}

:::danger
This feature is experimental and may not work on all platforms.
:::

Neko allows you to capture the webcam on the client machine and send it to the server using WebRTC. This can be used to share the webcam feed with the server.

The Gstreamer pipeline is started when the client shares their webcam and is stopped when the client stops sharing the webcam. Maximum one webcam pipeline can be active at a time.

<ConfigurationTab options={configOptions} filter={[
  "capture.webcam.enabled",
  "capture.webcam.device",
  "capture.webcam.width",
  "capture.webcam.height",
]} comments={false} />

- <Def id="webcam.enabled" /> is a boolean value that determines whether the webcam capture is enabled or not.
- <Def id="webcam.device" /> is the name of the [video4linux device](https://www.kernel.org/doc/html/v4.12/media/v4l-drivers/index.html) that will be used as a virtual webcam.
- <Def id="webcam.width" /> and <Def id="webcam.height" /> are the resolution of the virtual webcam feed.

In order to use the webcam feature, the server must have the [v4l2loopback](https://github.com/v4l2loopback/v4l2loopback) kernel module installed and loaded. The module can be loaded using the following command:

```bash
# Install the required packages (Debian/Ubuntu)
sudo apt install v4l2loopback-dkms v4l2loopback-utils linux-headers-`uname -r` linux-modules-extra-`uname -r`
# Load the module with exclusive_caps=1 to allow multiple applications to access the virtual webcam
sudo modprobe v4l2loopback exclusive_caps=1
```

This is needed even if neko is running inside a Docker container. In that case, the `v4l2loopback` module must be loaded on the host machine and the device must be mounted inside the container.

```yaml title="docker-compose.yaml"
services:
  neko:
    ...
    # highlight-start
    devices:
      - /dev/video0:/dev/video0
    # highlight-end
    ...
```

## Microphone {#microphone}

Neko allows you to capture the microphone on the client machine and send it to the server using WebRTC. This can be used to share the microphone feed with the server.

The Gstreamer pipeline is started when the client shares their microphone and is stopped when the client stops sharing the microphone. Maximum one microphone pipeline can be active at a time.

<ConfigurationTab options={configOptions} filter={[
  "capture.microphone.enabled",
  "capture.microphone.device",
]} comments={false} />

- <Def id="microphone.enabled" /> is a boolean value that determines whether the microphone capture is enabled or not.
- <Def id="microphone.device" /> is the name of the [pulseaudio device](https://wiki.archlinux.org/title/PulseAudio/Examples) that will be used as a virtual microphone.
