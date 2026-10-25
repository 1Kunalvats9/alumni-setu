#!/bin/bash
set -e

cd "$(dirname "$0")"
ROOT="$PWD"
VERSION="1.0.0"
APP="dist/AlumniSetu.app"
DMG="dist/AlumniSetu-${VERSION}.dmg"

rm -rf dist
mkdir -p dist

echo "==> building the Go server (universal)"
./build-server.sh >/dev/null
echo "    $(lipo -archs server/bin/alumnisetu-server) $(du -h server/bin/alumnisetu-server | cut -f1)"

echo "==> compiling the native shell (universal)"
swiftc -O -target arm64-apple-macos11  -o dist/shell-arm64 native/main.swift
swiftc -O -target x86_64-apple-macos11 -o dist/shell-x64   native/main.swift
lipo -create -output dist/AlumniSetu-shell dist/shell-arm64 dist/shell-x64
rm -f dist/shell-arm64 dist/shell-x64
echo "    $(lipo -archs dist/AlumniSetu-shell) $(du -h dist/AlumniSetu-shell | cut -f1)"

echo "==> assembling the app bundle"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources/server" "$APP/Contents/Resources/renderer"

mv dist/AlumniSetu-shell "$APP/Contents/MacOS/AlumniSetu"
cp server/bin/alumnisetu-server "$APP/Contents/Resources/server/alumnisetu-server"
chmod +x "$APP/Contents/MacOS/AlumniSetu" "$APP/Contents/Resources/server/alumnisetu-server"

cp -R desktop/renderer/. "$APP/Contents/Resources/renderer/"
cp desktop/build/icon.icns "$APP/Contents/Resources/AlumniSetu.icns"

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>                  <string>AlumniSetu</string>
  <key>CFBundleDisplayName</key>           <string>AlumniSetu</string>
  <key>CFBundleExecutable</key>            <string>AlumniSetu</string>
  <key>CFBundleIdentifier</key>            <string>in.gov.rajasthan.alumnisetu</string>
  <key>CFBundleIconFile</key>              <string>AlumniSetu</string>
  <key>CFBundlePackageType</key>           <string>APPL</string>
  <key>CFBundleShortVersionString</key>    <string>${VERSION}</string>
  <key>CFBundleVersion</key>               <string>${VERSION}</string>
  <key>LSMinimumSystemVersion</key>        <string>11.0</string>
  <key>LSApplicationCategoryType</key>     <string>public.app-category.education</string>
  <key>NSHighResolutionCapable</key>       <true/>
  <key>NSHumanReadableCopyright</key>      <string>Technical Education Department, Govt. of Rajasthan</string>
  <key>NSAppTransportSecurity</key>
  <dict>
    <key>NSAllowsLocalNetworking</key>     <true/>
  </dict>
</dict>
</plist>
PLIST

codesign --force --deep --sign - "$APP" 2>/dev/null || echo "    (ad-hoc signing skipped)"
echo "    app size: $(du -sh "$APP" | cut -f1)"

echo "==> building the dmg"
STAGE="dist/stage"
mkdir -p "$STAGE"
cp -R "$APP" "$STAGE/"
ln -s /Applications "$STAGE/Applications"

hdiutil create -volname "AlumniSetu ${VERSION}" -srcfolder "$STAGE" \
  -ov -format ULMO -fs APFS "$DMG" >/dev/null
rm -rf "$STAGE"

echo ""
echo "built $DMG"
ls -lh "$DMG" | awk '{print "   size:", $5}'
