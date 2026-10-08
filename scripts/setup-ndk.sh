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

if [ -d "$TARGET_DIR" ] && [ -x "$TARGET_DIR/toolchains/llvm/prebuilt/darwin-x86_64/bin/clang" ]; then
    echo ">>> Android NDK $NDK_RELEASE ($NDK_VERSION) is already installed at $TARGET_DIR"
    exit 0
fi

echo ">>> Preparing to install Android NDK $NDK_RELEASE into $TARGET_DIR..."
TMP_ZIP="/tmp/android-ndk-${NDK_RELEASE}-darwin.zip"

if [ ! -f "$TMP_ZIP" ]; then
    echo ">>> Downloading Android NDK from Google CDN..."
    curl -fSL --progress-bar -o "$TMP_ZIP" "https://dl.google.com/android/repository/android-ndk-${NDK_RELEASE}-darwin.zip"
else
    echo ">>> Found existing archive in $TMP_ZIP. Using cached file..."
fi

echo ">>> Extracting NDK archive to $SDK_DIR/ndk/..."
mkdir -p "$SDK_DIR/ndk"
unzip -qo "$TMP_ZIP" -d "$SDK_DIR/ndk/"

if [ -d "$SDK_DIR/ndk/android-ndk-${NDK_RELEASE}" ]; then
    rm -rf "$TARGET_DIR"
    mv "$SDK_DIR/ndk/android-ndk-${NDK_RELEASE}" "$TARGET_DIR"
fi

# Clear quarantine attributes on macOS to avoid Gatekeeper popups on NDK clang binaries
echo ">>> Clearing macOS quarantine flags..."
xattr -cr "$TARGET_DIR" 2>/dev/null || true

rm -f "$TMP_ZIP"
echo ">>> [SUCCESS] Android NDK $NDK_RELEASE successfully installed to $TARGET_DIR"
