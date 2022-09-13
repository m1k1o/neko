package capture

import (
	"fmt"
	"m1k1o/neko/internal/capture/gst"
	"strings"
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

// CreateRTMPPipeline creates a GStreamer Pipeline
func CreateRTMPPipeline(pipelineDevice string, pipelineDisplay string, pipelineSrc string, pipelineRTMP string) (*gst.Pipeline, error) {
	video := fmt.Sprintf(videoSrc, pipelineDisplay, 25)
	audio := fmt.Sprintf(audioSrc, pipelineDevice)

	var pipelineStr string
	if pipelineSrc != "" {
		// replace RTMP url
		pipelineStr = strings.Replace(pipelineSrc, "{url}", pipelineRTMP, -1)
		// replace audio device
		pipelineStr = strings.Replace(pipelineStr, "{device}", pipelineDevice, -1)
		// replace display
		pipelineStr = strings.Replace(pipelineStr, "{display}", pipelineDisplay, -1)
	} else {
		pipelineStr = fmt.Sprintf("flvmux name=mux ! rtmpsink location='%s live=1' %s audio/x-raw,channels=2 ! audioconvert ! voaacenc ! mux. %s x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! mux.", pipelineRTMP, audio, video)
	}

	return gst.CreatePipeline(pipelineStr)
}

// CreateAppPipeline creates a GStreamer Pipeline
func CreateAppPipeline(codecName string, pipelineDevice string, pipelineSrc string, fps int16, bitrate uint, hwenc string) (*gst.Pipeline, error) {
	pipelineStr := " ! appsink name=appsink"

	// if using custom pipeline
	if pipelineSrc != "" {
		pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		return gst.CreatePipeline(pipelineStr)
	}

	switch codecName {
	case "VP8":
		if hwenc == "VAAPI" {
			if err := gst.CheckPlugins([]string{"ximagesrc", "vaapi"}); err != nil {
				return nil, err
			}
			// vp8 encode is missing from gstreamer.freedesktop.org/documentation
			// note that it was removed from some recent intel CPUs: https://trac.ffmpeg.org/wiki/Hardware/QuickSync
			// https://gstreamer.freedesktop.org/data/doc/gstreamer/head/gstreamer-vaapi-plugins/html/gstreamer-vaapi-plugins-vaapivp8enc.html
			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! vaapivp8enc rate-control=vbr bitrate=%d keyframe-period=180"+pipelineStr, pipelineDevice, fps, bitrate)
		} else {
			// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html?gi-language=c
			// gstreamer1.0-plugins-good
			// vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1
			if err := gst.CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
				return nil, err
			}

			pipelineStr = strings.Join([]string{
				fmt.Sprintf(videoSrc, pipelineDevice, fps),
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
	case "VP9":
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp9enc
		if err := gst.CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(videoSrc+"vp9enc target-bitrate=%d cpu-used=-5 threads=4 deadline=1 keyframe-max-dist=30 auto-alt-ref=true"+pipelineStr, pipelineDevice, fps, bitrate*1000)
	case "H264":
		if err := gst.CheckPlugins([]string{"ximagesrc"}); err != nil {
			return nil, err
		}

		if hwenc == "VAAPI" {
			if err := gst.CheckPlugins([]string{"vaapi"}); err != nil {
				return nil, err
			}

			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! vaapih264enc rate-control=vbr bitrate=%d keyframe-period=180 quality-level=7 ! video/x-h264,stream-format=byte-stream"+pipelineStr, pipelineDevice, fps, bitrate)

		} else {
			// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html?gi-language=c#openh264enc
			// gstreamer1.0-plugins-bad
			// openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000
			if err := gst.CheckPlugins([]string{"openh264"}); err == nil {
				pipelineStr = fmt.Sprintf(videoSrc+"openh264enc multi-thread=4 complexity=high bitrate=%d max-bitrate=%d ! video/x-h264,stream-format=byte-stream"+pipelineStr, pipelineDevice, fps, bitrate*1000, (bitrate+1024)*1000)
				break
			}

			// https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c
			// gstreamer1.0-plugins-ugly
			// video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream
			if err := gst.CheckPlugins([]string{"x264"}); err != nil {
				return nil, err
			}

			vbvbuf := uint(1000)
			if bitrate > 1000 {
				vbvbuf = bitrate
			}

			pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=NV12 ! x264enc threads=4 bitrate=%d key-int-max=60 vbv-buf-capacity=%d byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream"+pipelineStr, pipelineDevice, fps, bitrate, vbvbuf)
		}
	case "Opus":
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		// opusenc
		if err := gst.CheckPlugins([]string{"pulseaudio", "opus"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"opusenc inband-fec=true bitrate=%d"+pipelineStr, pipelineDevice, bitrate*1000)
	case "G722":
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html?gi-language=c
		// gstreamer1.0-libav
		// avenc_g722
		if err := gst.CheckPlugins([]string{"pulseaudio", "libav"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"avenc_g722 bitrate=%d"+pipelineStr, pipelineDevice, bitrate*1000)
	case "PCMU":
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! mulawenc
		if err := gst.CheckPlugins([]string{"pulseaudio", "mulaw"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! mulawenc"+pipelineStr, pipelineDevice)
	case "PCMA":
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! alawenc
		if err := gst.CheckPlugins([]string{"pulseaudio", "alaw"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! alawenc"+pipelineStr, pipelineDevice)
	default:
		return nil, fmt.Errorf("unknown codec %s", codecName)
	}

	return gst.CreatePipeline(pipelineStr)
}
