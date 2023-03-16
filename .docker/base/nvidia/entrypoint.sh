#!/bin/bash -e

# Add VirtualGL directories to path
export PATH="${PATH}:/opt/VirtualGL/bin"

# Use VirtualGL to run wine with OpenGL if the GPU is available, otherwise use barebone wine
if [ -n "$(nvidia-smi --query-gpu=uuid --format=csv | sed -n 2p)" ]; then
    exec vglrun "$@"
else
    echo "No GPU detected"
    exec "$@"
fi
