#!/usr/bin/env bash
# Convert the recorded webm to a GIF and write it to docs/assets/demo-frontend.gif.
# Uses ffmpeg with palette optimization for clean colors.
#
# Usage: scripts/demo-frontend.sh [INPUT_WEBM] [OUTPUT_GIF]
# Defaults:
#   INPUT_WEBM  = docs/assets/demo-frontend.webm
#   OUTPUT_GIF  = docs/assets/demo-frontend.gif

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

INPUT="${1:-$REPO_ROOT/docs/assets/demo-frontend.webm}"
OUTPUT="${2:-$REPO_ROOT/docs/assets/demo-frontend.gif}"

if [ ! -f "$INPUT" ]; then
  echo "Error: input webm not found: $INPUT" >&2
  echo "Run 'node scripts/demo-frontend.js' first to record the session." >&2
  exit 1
fi

mkdir -p "$(dirname "$OUTPUT")"

# Two-pass encode with palettegen + paletteuse for best GIF quality.
# - fps=24 matches the existing terminal demo (VHS tape uses 24)
# - scale=1200 keeps the recorded viewport width
# - flags=lanczos produces a clean downscale
# - dither=bayer:bayer_scale=5 reduces banding without bloating the file
PALETTE_FILE="$(mktemp --suffix=.png)"
trap 'rm -f "$PALETTE_FILE"' EXIT

ffmpeg -loglevel error -y -i "$INPUT" \
  -vf "fps=24,scale=1200:-1:flags=lanczos,palettegen" \
  "$PALETTE_FILE"

ffmpeg -loglevel error -y -i "$INPUT" -i "$PALETTE_FILE" \
  -filter_complex "fps=24,scale=1200:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=5" \
  -loop 0 \
  "$OUTPUT"

echo "✓ Wrote $OUTPUT"
echo "  size: $(du -h "$OUTPUT" | cut -f1)"
