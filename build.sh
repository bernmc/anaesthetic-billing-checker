#!/bin/bash
# build.sh – Cross-compile the Anaesthetic Billing Checker for all platforms.
#
# Produces binaries in the dist/ directory:
#   billing-checker-darwin-arm64   (macOS Apple Silicon)
#   billing-checker-darwin-amd64   (macOS Intel)
#   billing-checker-linux-amd64    (Linux x86_64)
#   billing-checker-windows-amd64.exe  (Windows x86_64)
#
# Copyright (c) 2026 Bernard McClement
# Licensed under the MIT License. See LICENSE file in the project root.

set -e
cd "$(dirname "$0")"

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS="-s -w"

echo "=== Building Anaesthetic Billing Checker (${VERSION}) ==="
echo ""

mkdir -p dist

targets=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "windows/amd64"
)

for target in "${targets[@]}"; do
  os="${target%/*}"
  arch="${target#*/}"
  out="dist/billing-checker-${os}-${arch}"
  if [ "$os" = "windows" ]; then
    out="${out}.exe"
  fi
  echo "  Building ${os}/${arch} → ${out}"
  GOOS="$os" GOARCH="$arch" go build -ldflags="$LDFLAGS" -o "$out" .
done

echo ""
echo "Done. Binaries in dist/:"
ls -lh dist/
echo ""
echo "To distribute: copy the correct binary into the project root as"
echo "  'billing-checker' (macOS/Linux) or 'billing-checker.exe' (Windows)"
echo "along with the app/ folder and the start/stop scripts."
