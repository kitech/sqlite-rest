#!/bin/bash
set -euo pipefail

NDK=/opt/android-ndk
TOOLCHAIN=$NDK/toolchains/llvm/prebuilt/linux-x86_64
CC=$TOOLCHAIN/bin/armv7a-linux-androideabi21-clang

CGO_ENABLED=1 GOOS=android GOARCH=arm CC=$CC \
  go build -trimpath -ldflags="-s -w" -o sqlite-rest-arm32 .
