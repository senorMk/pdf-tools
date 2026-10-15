#!/usr/bin/env bash
set -euo pipefail
if [[ "$(go env GOOS)" != "darwin" ]]; then
  echo "macOS packaging requires macOS and Xcode Command Line Tools." >&2
  exit 1
fi
bundle="build/bin/PDF Tools.app"
iconset="build/bin/PDF Tools.iconset"
mkdir -p "$bundle/Contents/MacOS" "$bundle/Contents/Resources"
mkdir -p "$iconset"
for size in 16 32 128 256 512; do
  sips -z "$size" "$size" build/appicon.png --out "$iconset/icon_${size}x${size}.png" >/dev/null
  retina_size=$((size * 2))
  sips -z "$retina_size" "$retina_size" build/appicon.png --out "$iconset/icon_${size}x${size}@2x.png" >/dev/null
done
go run scripts/iconpack/main.go "$iconset" "$bundle/Contents/Resources/appicon.icns"
go build -tags desktop,production -ldflags '-s -w' -o "$bundle/Contents/MacOS/pdf-tools" .
cp build/Info.plist "$bundle/Contents/Info.plist"
codesign --force --sign "${PDFTOOLS_SIGN_IDENTITY:--}" --identifier com.senormk.pdf-tools "$bundle"
codesign --verify --deep --strict "$bundle"
echo "Built $bundle"
