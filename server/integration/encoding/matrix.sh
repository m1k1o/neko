#!/usr/bin/env bash

set -euo pipefail

gst_inspect="${GST_INSPECT:-gst-inspect-1.0}"
gst_launch="${GST_LAUNCH:-gst-launch-1.0}"
output="${NEKO_ENCODER_MATRIX_OUTPUT:-}"
strict="${NEKO_ENCODER_MATRIX_STRICT:-0}"
family="${NEKO_ENCODER_MATRIX_FAMILY:-all}"

case "$family" in
	all|software|vaapi|nvenc)
		;;
	*)
		echo "NEKO_ENCODER_MATRIX_FAMILY must be all, software, vaapi, or nvenc" >&2
		exit 2
		;;
esac

if ! command -v "$gst_inspect" >/dev/null 2>&1 || ! command -v "$gst_launch" >/dev/null 2>&1; then
	echo "GStreamer tools are required: $gst_inspect and $gst_launch" >&2
	exit 2
fi

if [[ -n "$output" ]]; then
	mkdir -p "$(dirname "$output")"
	exec > >(tee "$output")
fi

printf 'codec\tencoder\telement\tstatus\tdiagnostic\n'

matrix=(
	"vp8|software|vp8enc|"
	"h264|software|x264enc|h264parse"
	"h264|vaapi|vah264enc|h264parse"
	"h264|vaapi|vah264lpenc|h264parse"
	"h264|nvenc|nvautogpuh264enc|h264parse"
	"h264|nvenc|nvh264enc|h264parse"
	"h265|software|x265enc|h265parse"
	"h265|vaapi|vah265enc|h265parse"
	"h265|vaapi|vah265lpenc|h265parse"
	"h265|nvenc|nvautogpuh265enc|h265parse"
	"h265|nvenc|nvh265enc|h265parse"
	"av1|software|av1enc|"
	"av1|software|svtav1enc|"
	"av1|vaapi|vaav1enc|"
	"av1|nvenc|nvautogpuav1enc|"
	"av1|nvenc|nvav1enc|"
)

failed=0
unavailable=0
matrix_log="$(mktemp)"
trap 'rm -f "$matrix_log"' EXIT

for row in "${matrix[@]}"; do
	IFS='|' read -r codec encoder element parser <<< "$row"
	if [[ "$family" != "all" && "$encoder" != "$family" ]]; then
		continue
	fi

	if ! "$gst_inspect" "$element" >/dev/null 2>&1; then
		printf '%s\t%s\t%s\tunavailable\tencoder element not installed\n' "$codec" "$encoder" "$element"
		unavailable=$((unavailable + 1))
		continue
	fi
	if [[ -n "$parser" ]] && ! "$gst_inspect" "$parser" >/dev/null 2>&1; then
		printf '%s\t%s\t%s\tunavailable\tparser element %s not installed\n' "$codec" "$encoder" "$element" "$parser"
		unavailable=$((unavailable + 1))
		continue
	fi

	pipeline=(
		videotestsrc is-live=true num-buffers=60 pattern=black
		! videoconvert
		! video/x-raw,width=1280,height=720,framerate=30/1
		! "$element"
	)
	case "$codec" in
		h264)
			pipeline+=( ! "$parser" ! video/x-h264,stream-format=byte-stream )
			;;
		h265)
			pipeline+=( ! "$parser" ! video/x-h265,stream-format=byte-stream )
			;;
		av1)
			pipeline+=( ! video/x-av1,stream-format=obu-stream,alignment=tu )
			;;
		vp8)
			;;
		*)
			printf '%s\t%s\t%s\tfail\tunknown matrix codec\n' "$codec" "$encoder" "$element"
			failed=$((failed + 1))
			continue
			;;
	esac
	pipeline+=( ! fakesink sync=false )

	if "$gst_launch" -q -e "${pipeline[@]}" >"$matrix_log" 2>&1; then
		printf '%s\t%s\t%s\tpass\truntime pipeline reached PLAYING\n' "$codec" "$encoder" "$element"
	else
		diagnostic="$(tail -n 1 "$matrix_log" 2>/dev/null || true)"
		printf '%s\t%s\t%s\tfail\t%s\n' "$codec" "$encoder" "$element" "${diagnostic:-runtime pipeline failed}"
		failed=$((failed + 1))
	fi
done

printf '\nMatrix summary: failed=%d unavailable=%d\n' "$failed" "$unavailable"
if [[ "$failed" -gt 0 ]] || [[ "$strict" == "1" && "$unavailable" -gt 0 ]]; then
	exit 1
fi
