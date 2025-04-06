package capture

import (
	"fmt"
	"strings"

	"m1k1o/neko/internal/capture/gst"
	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/types/codec"
)

/*
	apt-get install \
		libgstreamer1.0-0 \
		gstreamer1.0-plugins-base \
		gstreamer1.0-plugins-good \
		gstreamer1.0-plugins-bad \
		gstreamer1.0-plugins-ugly\
		gstreamer1.0-libav \
		gstreamer1.0-doc \
		gstreamer1.0-tools \
		gstreamer1.0-x \
		gstreamer1.0-alsa \
    gstreamer1.0-pulseaudio

    gst-inspect-1.0 --version
    gst-inspect-1.0 plugin
    gst-launch-1.0 ximagesrc show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! autovideosink
    gst-launch-1.0 pulsesrc ! audioconvert ! opusenc ! autoaudiosink
*/

const (
	videoSrc = "ximagesrc display-name=%s show-pointer=true use-damage=false ! video/x-raw,framerate=%d/1 ! videoconvert ! queue ! "
	audioSrc = "pulsesrc device=%s ! audio/x-raw,channels=2 ! audioconvert ! "
)

func NewBroadcastPipeline(device string, display string, pipelineSrc string, url string) (string, error) {
	video := fmt.Sprintf(videoSrc, display, 25)
	audio := fmt.Sprintf(audioSrc, device)

	var pipelineStr string
	if pipelineSrc != "" {
		// replace RTMP url
		pipelineStr = strings.Replace(pipelineSrc, "{url}", url, -1)
		// replace audio device
		pipelineStr = strings.Replace(pipelineStr, "{device}", device, -1)
		// replace display
		pipelineStr = strings.Replace(pipelineStr, "{display}", display, -1)
	} else {
		pipelineStr = fmt.Sprintf("flvmux name=mux ! rtmpsink location='%s live=1' %s audio/x-raw,channels=2 ! audioconvert ! voaacenc ! mux. %s x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! mux.", url, audio, video)
	}

	return pipelineStr, nil
}

func NewVideoPipeline(rtpCodec codec.RTPCodec, display string, pipelineSrc string, fps int16, bitrate uint, hwenc config.HwEnc) (string, error) {
	pipelineStr := " ! appsink name=appsinkvideo"

	// if using custom pipeline
	if pipelineSrc != "" {
		pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, display)
		return pipelineStr, nil
	}

	// use default fps if not set
	if fps == 0 {
		fps = 25
	}

	switch rtpCodec.Name {
	case codec.VP8().Name:
		if hwenc == config.HwEncVAAPI {
			if err := gst.CheckPlugins([]string{"ximagesrc", "vaapi"}); err != nil {
				return "", err
			}
			// vp8 encode is missing from gstreamer.freedesktop.org/documentation
			// note that it was removed from some recent intel CPUs: https://trac.ffmpeg.org/wiki/Hardware/QuickSync
			// https://gstreamer.freedesktop.org/data/doc/gstreamer/head/gstreamer-vaapi-plugins/html/gstreamer-vaapi-plugins-vaapivp8enc.html
			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! vaapivp8enc rate-control=vbr bitrate=%d keyframe-period=180"+pipelineStr, display, fps, bitrate)
		} else {
			// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html?gi-language=c
			// gstreamer1.0-plugins-good
			// vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1
			if err := gst.CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
				return "", err
			}

			pipelineStr = strings.Join([]string{
				fmt.Sprintf(videoSrc, display, fps),
				"vp8enc",
				fmt.Sprintf("target-bitrate=%d", bitrate*650),
				"cpu-used=4",
				"end-usage=cbr",
				"threads=4",
				"deadline=1",
				"undershoot=95",
				fmt.Sprintf("buffer-size=%d", bitrate*4),
				fmt.Sprintf("buffer-initial-size=%d", bitrate*2),
				fmt.Sprintf("buffer-optimal-size=%d", bitrate*3),
				"keyframe-max-dist=25",
				"min-quantizer=4",
				"max-quantizer=20",
				pipelineStr,
			}, " ")
		}
	case codec.VP9().Name:
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp9enc
		if err := gst.CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return "", err
		}

		pipelineStr = fmt.Sprintf(videoSrc+"vp9enc target-bitrate=%d cpu-used=-5 threads=4 deadline=1 keyframe-max-dist=30 auto-alt-ref=true"+pipelineStr, display, fps, bitrate*1000)
	case codec.AV1().Name:
		// https://gstreamer.freedesktop.org/documentation/aom/av1enc.html?gi-language=c
		// gstreamer1.0-plugins-bad
		// av1enc usage-profile=1
		// TODO: check for plugin.
		if err := gst.CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return "", err
		}

		pipelineStr = strings.Join([]string{
			fmt.Sprintf(videoSrc, display, fps),
			"av1enc",
			fmt.Sprintf("target-bitrate=%d", bitrate*650),
			"cpu-used=4",
			"end-usage=cbr",
			// "usage-profile=realtime",
			"undershoot=95",
			"keyframe-max-dist=25",
			"min-quantizer=4",
			"max-quantizer=20",
			pipelineStr,
		}, " ")
	case codec.H264().Name:
		if err := gst.CheckPlugins([]string{"ximagesrc"}); err != nil {
			return "", err
		}

		vbvbuf := uint(1000)
		if bitrate > 1000 {
			vbvbuf = bitrate
		}

		if hwenc == config.HwEncVAAPI {
			if err := gst.CheckPlugins([]string{"vaapi"}); err != nil {
				return "", err
			}

			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! vaapih264enc rate-control=vbr bitrate=%d keyframe-period=180 quality-level=7 ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline"+pipelineStr, display, fps, bitrate)
		} else if hwenc == config.HwEncNVENC {
			if err := gst.CheckPlugins([]string{"nvcodec"}); err != nil {
				return "", err
			}

			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! nvh264enc name=encoder preset=2 gop-size=25 spatial-aq=true temporal-aq=true bitrate=%d vbv-buffer-size=%d rc-mode=6 ! h264parse config-interval=-1 ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline"+pipelineStr, display, fps, bitrate, vbvbuf)
		} else {
			// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html?gi-language=c#openh264enc
			// gstreamer1.0-plugins-bad
			// openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000
			if err := gst.CheckPlugins([]string{"openh264"}); err == nil {
				pipelineStr = fmt.Sprintf(videoSrc+"openh264enc multi-thread=4 complexity=high bitrate=%d max-bitrate=%d ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline"+pipelineStr, display, fps, bitrate*1000, (bitrate+1024)*1000)
				break
			}

			// https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c
			// gstreamer1.0-plugins-ugly
			// video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline
			if err := gst.CheckPlugins([]string{"x264"}); err != nil {
				return "", err
			}

			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! x264enc threads=4 bitrate=%d key-int-max=60 vbv-buf-capacity=%d byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline"+pipelineStr, display, fps, bitrate, vbvbuf)
		}
	default:
		return "", fmt.Errorf("unknown codec %s", rtpCodec.Name)
	}

	return pipelineStr, nil
}

func NewAudioPipeline(rtpCodec codec.RTPCodec, device string, pipelineSrc string, bitrate uint) (string, error) {
	pipelineStr := " ! appsink name=appsinkaudio"

	// if using custom pipeline
	if pipelineSrc != "" {
		pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, device)
		return pipelineStr, nil
	}

	switch rtpCodec.Name {
	case codec.Opus().Name:
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		// opusenc
		if err := gst.CheckPlugins([]string{"pulseaudio", "opus"}); err != nil {
			return "", err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"opusenc inband-fec=true bitrate=%d"+pipelineStr, device, bitrate*1000)
	case codec.G722().Name:
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html?gi-language=c
		// gstreamer1.0-libav
		// avenc_g722
		if err := gst.CheckPlugins([]string{"pulseaudio", "libav"}); err != nil {
			return "", err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"avenc_g722 bitrate=%d"+pipelineStr, device, bitrate*1000)
	case codec.PCMU().Name:
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! mulawenc
		if err := gst.CheckPlugins([]string{"pulseaudio", "mulaw"}); err != nil {
			return "", err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! mulawenc"+pipelineStr, device)
	case codec.PCMA().Name:
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! alawenc
		if err := gst.CheckPlugins([]string{"pulseaudio", "alaw"}); err != nil {
			return "", err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! alawenc"+pipelineStr, device)
	default:
		return "", fmt.Errorf("unknown codec %s", rtpCodec.Name)
	}

	return pipelineStr, nil
}

func NewWebSocketPipeline(videoCodec string, audioCodec string, display string, device string, pipelineSrc string, fps int16, videoBitrate uint, audioBitrate uint, hwenc config.HwEnc, fragmentDuration uint) (string, error) {
	// Define the appsink for the combined fMP4 stream
	appSink := fmt.Sprintf("appsink name=appsinkws emit-signals=true max-buffers=10 drop=true sync=false")

	// --- Use custom pipeline if provided --- (Note: Needs careful construction by user)
	if pipelineSrc != "" {
		// Replace placeholders in the custom pipeline
		pipelineStr := strings.Replace(pipelineSrc, "{display}", display, -1)
		pipelineStr = strings.Replace(pipelineStr, "{device}", device, -1)
		pipelineStr = strings.Replace(pipelineStr, "{appsink}", appSink, -1)
		// TODO: Add checks for essential elements if possible, or rely on user correctness
		return pipelineStr, nil
	}

	// --- Default Pipeline Construction ---

	// use default fps if not set
	if fps == 0 {
		fps = 25
	}

	// --- Check Core Plugins ---
	if err := gst.CheckPlugins([]string{"ximagesrc", "pulseaudio", "mp4mux", "h264parse", "aacparse", "voaacenc"}); err != nil {
		return "", fmt.Errorf("missing core plugins for WebSocket streaming: %w", err)
	}

	// --- Video Source and Encoding --- (H.264 only for fMP4/MSE compatibility)
	if videoCodec != "h264" {
		return "", fmt.Errorf("unsupported video codec for WebSocket streaming: %s. Only H.264 is supported", videoCodec)
	}

	videoSrcSegment := fmt.Sprintf(videoSrc, display, fps)
	var videoEncSegment string

	vbvbuf := uint(1000) // vbv-buf-capacity for x264enc
	if videoBitrate > 1000 {
		vbvbuf = videoBitrate
	}

	switch hwenc {
	case config.HwEncVAAPI:
		if err := gst.CheckPlugins([]string{"vaapi"}); err != nil {
			return "", fmt.Errorf("VAAPI requested but plugin not found: %w", err)
		}
		// vaapih264enc: rate-control=cbr might be better for streaming
		videoEncSegment = fmt.Sprintf("video/x-raw,format=NV12 ! vaapih264enc rate-control=cbr bitrate=%d keyframe-period=%d quality-level=7 ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline", videoBitrate, fps*2) // keyframe every 2s
	case config.HwEncNVENC:
		if err := gst.CheckPlugins([]string{"nvcodec"}); err != nil {
			return "", fmt.Errorf("NVENC requested but plugin not found: %w", err)
		}
		videoEncSegment = fmt.Sprintf("video/x-raw,format=NV12 ! nvh264enc name=encoder preset=llhq gop-size=%d spatial-aq=true temporal-aq=true bitrate=%d vbv-buffer-size=%d rc-mode=cbr ! h264parse config-interval=-1 ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline", fps*2, videoBitrate, vbvbuf)
	case config.HwEncNone: // Fallback to software encoding
		if err := gst.CheckPlugins([]string{"openh264"}); err == nil {
			videoEncSegment = fmt.Sprintf("openh264enc multi-thread=4 complexity=realtime bitrate=%d max-bitrate=%d ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline", videoBitrate*1000, (videoBitrate+1024)*1000)
		} else if err := gst.CheckPlugins([]string{"x264"}); err == nil {
			videoEncSegment = fmt.Sprintf("video/x-raw,format=NV12 ! x264enc threads=4 bitrate=%d key-int-max=%d vbv-buf-capacity=%d byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline", videoBitrate, fps*2, vbvbuf)
		} else {
			return "", fmt.Errorf("no suitable software H.264 encoder found (tried openh264, x264)")
		}
	default:
		return "", fmt.Errorf("unknown hardware acceleration config: %s", hwenc)
	}

	// --- Audio Source and Encoding --- (AAC only for fMP4/MSE compatibility)
	if audioCodec != "aac" {
		return "", fmt.Errorf("unsupported audio codec for WebSocket streaming: %s. Only AAC is supported", audioCodec)
	}
	// Using voaacenc as it's often available and used in broadcast pipeline
	audioSrcSegment := fmt.Sprintf(audioSrc, device)
	audioEncSegment := fmt.Sprintf("voaacenc bitrate=%d", audioBitrate*1000)

	// --- Muxing --- 
	// mp4mux name=mux fragment-duration=100 streamable=true presentation-time=true
	muxerSegment := fmt.Sprintf("mp4mux name=mux fragment-duration=%d streamable=true presentation-time=true ! %s", fragmentDuration, appSink)

	// --- Assemble the pipeline --- 
	pipelineStr := fmt.Sprintf(
		"%s %s ! h264parse config-interval=-1 ! mux.  %s %s ! aacparse ! mux.  %s",
		videoSrcSegment, // ximagesrc ! videoconvert ! queue
		videoEncSegment, // encoder ! caps
		audioSrcSegment, // pulsesrc ! audioconvert
		audioEncSegment, // voaacenc
		muxerSegment,    // mp4mux ! appsink
	)

	return pipelineStr, nil
}
