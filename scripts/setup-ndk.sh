#!/usr/bin/env bash
# ==============================================================================
# Rastreador Mobile — Automated Android NDK Installer
# Downloads and installs Android NDK r26d into the local Android SDK.
# ==============================================================================
set -euo pipefail

NDK_VERSION="26.3.11579264"
NDK_RELEASE="r26d"
SDK_DIR="${ANDROID_HOME:-/Users/dodecaneser/Library/Android/sdk}"
TARGET_DIR="$SDK_DIR/ndk/$NDK_VERSION"

if [ -d "$TARGET_DIR" ]; then
    echo ">>> Android NDK $NDK_RELEASE ($NDK_VERSION) is already installed at $TARGET_DIR"
    exit 0
fi

echo ">>> Preparing to install Android NDK $NDK_RELEASE into $TARGET_DIR..."
TMP_ZIP="/tmp/android-ndk-${NDK_RELEASE}-darwin.zip"

echo ">>> Downloading Android NDK from Google CDN..."
curl -fSL --progress-bar -o "$TMP_ZIP" "https://dl.google.com/android/repository/android-ndk-${NDK_RELEASE}-darwin.zip"

echo ">>> Extracting NDK archive to $SDK_DIR/ndk/..."
mkdir -p "$SDK_DIR/ndk"
unzip -q "$TMP_ZIP" -d "$SDK_DIR/ndk/"

if [ -d "$SDK_DIR/ndk/android-ndk-${NDK_RELEASE}" ]; then
    mv "$SDK_DIR/ndk/android-ndk-${NDK_RELEASE}" "$TARGET_DIR"
fi

rm -f "$TMP_ZIP"
echo ">>> [SUCCESS] Android NDK $NDK_RELEASE installed to $TARGET_DIR"
