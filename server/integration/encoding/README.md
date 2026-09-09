# Encoder capability matrix

`matrix.sh` verifies the encoder elements used by the Chromium M1 profiles. It
checks both GStreamer registry visibility and a short `videotestsrc` pipeline
that must reach `PLAYING`. The probe does not require an X display.

The matrix covers:

- software VP8, H.264, H.265, AV1 and SVT-AV1;
- VAAPI H.264, H.265 and AV1, including low-power H.264/H.265 variants;
- NVIDIA NVENC H.264, H.265 and AV1, including auto-GPU and CUDA variants.

Run it inside the same runtime image and with the same GPU device mappings used
by Neko:

```bash
# CPU/software and registry checks
server/integration/encoding/matrix.sh

# Intel/AMD VAAPI
docker run --rm --device /dev/dri \
  -v "$PWD/server/integration/encoding:/matrix:ro" \
  <neko-image> /matrix/matrix.sh

# NVIDIA NVENC/AV1
docker run --rm --gpus all \
  -v "$PWD/server/integration/encoding:/matrix:ro" \
  <neko-nvidia-image> /matrix/matrix.sh
```

Missing optional elements are reported as `unavailable` and do not fail the
default run. Set `NEKO_ENCODER_MATRIX_STRICT=1` to require every row, which is
appropriate for a hardware-specific release job. Set
`NEKO_ENCODER_MATRIX_OUTPUT=/path/result.tsv` to archive the result.
