#!/bin/sh
set -e
# Fix ownership of the HLS volume mount before dropping privileges.
# Docker creates named volumes as root; this runs once at container start.
chown appuser:appuser /app/hls
exec su-exec appuser "$@"
