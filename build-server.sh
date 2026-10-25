#!/bin/bash
set -e

cd "$(dirname "$0")/server"
mkdir -p bin

echo "building the Go server as a universal macOS binary..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/server-arm64 .
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/server-amd64 .

lipo -create -output bin/alumnisetu-server bin/server-arm64 bin/server-amd64
rm -f bin/server-arm64 bin/server-amd64
chmod +x bin/alumnisetu-server

echo "done:"
lipo -archs bin/alumnisetu-server
ls -lh bin/alumnisetu-server
