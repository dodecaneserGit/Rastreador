#!/usr/bin/env bash
# ==============================================================================
# Rastreador Mobile — Native Go Core Engine Automated Build Script
# Targets: Android arm64-v8a & x86_64
# Supports:
#   --aar  : Gomobile bind producing app/libs/rastreador-core.aar (default)
#   --so   : Direct Cgo C-shared producing app/src/main/jniLibs/{abi}/librastreador.so
# ==============================================================================
set -euo pipefail

MODE="${1:---aar}"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_ENGINE_DIR="/Volumes/SSD/Proyectos/Rastreador"
ANDROID_SDK="${ANDROID_HOME:-/Users/dodecaneser/Library/Android/sdk}"

echo ">>> [1/5] Validating Build Environment..."

# 1. Validate Android SDK
if [ ! -d "$ANDROID_SDK" ]; then
    echo "ERROR: Android SDK not found at $ANDROID_SDK. Set ANDROID_HOME." >&2
    exit 1
fi
export ANDROID_HOME="$ANDROID_SDK"

# 2. Autodetect Android NDK
NDK_DIR="${ANDROID_NDK_HOME:-}"
if [ -z "$NDK_DIR" ] || [ ! -d "$NDK_DIR" ]; then
    if [ -d "$ANDROID_SDK/ndk" ]; then
        NDK_DIR=$(ls -d "$ANDROID_SDK/ndk"/* 2>/dev/null | sort -V | tail -n 1 || true)
    elif [ -d "/opt/homebrew/share/android-ndk" ]; then
        NDK_DIR="/opt/homebrew/share/android-ndk"
    fi
fi

if [ -z "$NDK_DIR" ] || [ ! -d "$NDK_DIR" ]; then
    echo "WARNING: Android NDK not found in $ANDROID_SDK/ndk or /opt/homebrew/share/android-ndk." >&2
    echo "To build native .aar or .so binaries, install NDK by running:" >&2
    echo "    bash \"$PROJECT_ROOT/scripts/setup-ndk.sh\"" >&2
    exit 1
fi
export ANDROID_NDK_HOME="$NDK_DIR"
echo "    Found Android SDK: $ANDROID_SDK"
echo "    Found Android NDK: $NDK_DIR"

# 3. Clean AppleDouble files on external drive
dot_clean "$GO_ENGINE_DIR" 2>/dev/null || true

cd "$GO_ENGINE_DIR"

# 4. Compile according to mode
if [ "$MODE" = "--aar" ]; then
    echo ">>> [2/5] Building Android AAR via gomobile bind..."

    mkdir -p "$PROJECT_ROOT/app/libs"
    OUTPUT_AAR="$PROJECT_ROOT/app/libs/rastreador-core.aar"

    gomobile bind \
        -target=android/arm64,android/amd64 \
        -javapkg=com.dodecaneser.rastreador.bridge \
        -o "$OUTPUT_AAR" \
        ./mobile

    echo ">>> [SUCCESS] Generated: $OUTPUT_AAR"
    ls -lh "$OUTPUT_AAR"

elif [ "$MODE" = "--so" ]; then
    echo ">>> [2/5] Cross-compiling Cgo JNI Shared Libraries (.so)..."

    TOOLCHAIN="$NDK_DIR/toolchains/llvm/prebuilt/darwin-x86_64"
    CLANG_ARM64="$TOOLCHAIN/bin/aarch64-linux-android26-clang"
    CLANG_X86_64="$TOOLCHAIN/bin/x86_64-linux-android26-clang"

    OUT_ARM64="$PROJECT_ROOT/app/src/main/jniLibs/arm64-v8a"
    OUT_X86_64="$PROJECT_ROOT/app/src/main/jniLibs/x86_64"
    mkdir -p "$OUT_ARM64" "$OUT_X86_64"

    echo "    Compiling arm64-v8a..."
    CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC="$CLANG_ARM64" \
        go build -buildmode=c-shared -ldflags="-s -w" \
        -o "$OUT_ARM64/librastreador.so" ./mobile

    echo "    Compiling x86_64..."
    CGO_ENABLED=1 GOOS=android GOARCH=amd64 CC="$CLANG_X86_64" \
        go build -buildmode=c-shared -ldflags="-s -w" \
        -o "$OUT_X86_64/librastreador.so" ./mobile

    echo ">>> [SUCCESS] Generated native shared libraries in app/src/main/jniLibs/"
    ls -lh "$OUT_ARM64/librastreador.so" "$OUT_X86_64/librastreador.so"
else
    echo "ERROR: Unknown mode $MODE. Use --aar or --so." >&2
    exit 1
fi

echo ">>> [3/5] Triggering Gradle packaging verification..."
cd "$PROJECT_ROOT"
./gradlew assembleDebug --no-daemon
echo ">>> [COMPLETE] rastreador-android successfully linked with Go core engine!"
