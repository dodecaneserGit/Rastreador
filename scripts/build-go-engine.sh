#!/usr/bin/env bash
# ==============================================================================
# Rastreador Mobile — Native Go Core Engine Automated Build Script
# Targets: Android arm64-v8a & x86_64
# Supported Modes:
#   --aar  : Gomobile bind producing app/libs/rastreador-core.aar (default)
#   --so   : Direct Cgo C-shared producing app/src/main/jniLibs/{abi}/librastreador.so
#   --all  : Builds both --aar and --so targets
# ==============================================================================
set -euo pipefail

MODE="${1:---aar}"
if [ "$MODE" = "--help" ] || [ "$MODE" = "-h" ]; then
    echo "Usage: $0 [--aar | --so | --all]"
    echo "  --aar : Compile Go engine into app/libs/rastreador-core.aar (default)"
    echo "  --so  : Compile Go engine into app/src/main/jniLibs/{abi}/librastreador.so"
    echo "  --all : Compile both --aar and --so targets"
    exit 0
fi

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_ENGINE_DIR="/Volumes/SSD/Proyectos/Rastreador"
ANDROID_SDK="${ANDROID_HOME:-/Users/dodecaneser/Library/Android/sdk}"

# Ensure Go and Gomobile binaries in user home are discoverable in non-interactive shells
export PATH="$HOME/go/bin:/usr/local/go/bin:$PATH"

echo "================================================================================"
echo "  Rastreador Mobile — Native Engine Build Toolchain"
echo "  Target Mode: $MODE"
echo "================================================================================"

# ------------------------------------------------------------------------------
# 1. Environment & SDK Validation
# ------------------------------------------------------------------------------
echo ">>> [1/5] Validating Build Environment..."

# 1.1 Verify Go installation
if ! command -v go &>/dev/null; then
    echo "ERROR: Go toolchain not found in PATH ($PATH)." >&2
    exit 1
fi
echo "    Found Go: $(go version)"

# 1.2 Verify Android SDK
if [ ! -d "$ANDROID_SDK" ]; then
    echo "ERROR: Android SDK not found at $ANDROID_SDK. Set ANDROID_HOME." >&2
    exit 1
fi
export ANDROID_HOME="$ANDROID_SDK"
echo "    Found Android SDK: $ANDROID_SDK"

# 1.3 Autodetect or Install Android NDK
NDK_DIR="${ANDROID_NDK_HOME:-}"
if [ -z "$NDK_DIR" ] || [ ! -d "$NDK_DIR" ]; then
    if [ -d "$ANDROID_SDK/ndk" ]; then
        NDK_DIR=$(ls -d "$ANDROID_SDK/ndk"/* 2>/dev/null | sort -V | tail -n 1 || true)
    elif [ -d "/opt/homebrew/share/android-ndk" ]; then
        NDK_DIR="/opt/homebrew/share/android-ndk"
    fi
fi

if [ -z "$NDK_DIR" ] || [ ! -d "$NDK_DIR" ]; then
    SETUP_NDK_SCRIPT="$PROJECT_ROOT/scripts/setup-ndk.sh"
    if [ -f "$SETUP_NDK_SCRIPT" ]; then
        echo ">>> Android NDK not found. Invoking automated installer: $SETUP_NDK_SCRIPT"
        bash "$SETUP_NDK_SCRIPT"
        NDK_DIR=$(ls -d "$ANDROID_SDK/ndk"/* 2>/dev/null | sort -V | tail -n 1 || true)
    fi
fi

if [ -z "$NDK_DIR" ] || [ ! -d "$NDK_DIR" ]; then
    echo "ERROR: Android NDK not found in $ANDROID_SDK/ndk or /opt/homebrew/share/android-ndk." >&2
    echo "Please install Android NDK by running: bash scripts/setup-ndk.sh" >&2
    exit 1
fi
export ANDROID_NDK_HOME="$NDK_DIR"
echo "    Found Android NDK: $NDK_DIR"

# 1.4 Clean AppleDouble metadata files on external volume
dot_clean "$GO_ENGINE_DIR" 2>/dev/null || true
dot_clean "$PROJECT_ROOT" 2>/dev/null || true

# 1.5 Verify / Auto-configure Go Module Toolchain Dependencies
if ! grep -q "golang.org/x/mobile" "$GO_ENGINE_DIR/go.mod" 2>/dev/null; then
    echo ">>> [!] golang.org/x/mobile not found in $GO_ENGINE_DIR/go.mod. Registering tool directive..."
    (cd "$GO_ENGINE_DIR" && go get -tool golang.org/x/mobile/cmd/gobind)
fi

cd "$GO_ENGINE_DIR"

# ------------------------------------------------------------------------------
# 2. Build Functions
# ------------------------------------------------------------------------------
build_aar() {
    echo ">>> [2/5] Building Android AAR via gomobile bind..."

    if ! command -v gomobile &>/dev/null; then
        echo "ERROR: gomobile CLI not found. Install via: go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init" >&2
        exit 1
    fi

    mkdir -p "$PROJECT_ROOT/app/libs"
    local output_aar="$PROJECT_ROOT/app/libs/rastreador-core.aar"

    echo "    Compiling Go package './mobile' -> '$output_aar' (arm64, amd64)..."
    gomobile bind \
        -target=android/arm64,android/amd64 \
        -javapkg=com.dodecaneser.rastreador.bridge \
        -o "$output_aar" \
        ./mobile

    if [ ! -f "$output_aar" ] || [ ! -s "$output_aar" ]; then
        echo "ERROR: AAR compilation failed; output file missing or empty." >&2
        exit 1
    fi

    echo ">>> [SUCCESS] Generated AAR: $output_aar ($(du -h "$output_aar" | cut -f1))"
}

build_so() {
    echo ">>> [2/5] Cross-compiling Cgo JNI Shared Libraries (.so)..."

    # Identify Cgo main package entry point
    local cgo_pkg=""
    if [ -d "$GO_ENGINE_DIR/cmd/librastreador" ]; then
        cgo_pkg="./cmd/librastreador"
    elif [ -d "$GO_ENGINE_DIR/jni" ]; then
        cgo_pkg="./jni"
    elif [ -d "$GO_ENGINE_DIR/mobile/cgo" ]; then
        cgo_pkg="./mobile/cgo"
    else
        echo "ERROR: JNI Cgo main package not found in cmd/librastreador, jni, or mobile/cgo." >&2
        echo "Cgo -buildmode=c-shared requires a package main exporting JNI functions." >&2
        exit 1
    fi
    echo "    Using Cgo main package: $cgo_pkg"

    # Locate LLVM toolchain
    local toolchain
    toolchain=$(ls -d "$NDK_DIR/toolchains/llvm/prebuilt"/darwin-* 2>/dev/null | head -n 1 || true)
    if [ -z "$toolchain" ] || [ ! -d "$toolchain" ]; then
        echo "ERROR: LLVM toolchain not found under $NDK_DIR/toolchains/llvm/prebuilt/." >&2
        exit 1
    fi

    local clang_arm64="$toolchain/bin/aarch64-linux-android26-clang"
    local clang_x86_64="$toolchain/bin/x86_64-linux-android26-clang"

    if [ ! -x "$clang_arm64" ] || [ ! -x "$clang_x86_64" ]; then
        echo "ERROR: NDK clang compilers not found in $toolchain/bin." >&2
        exit 1
    fi

    local out_arm64="$PROJECT_ROOT/app/src/main/jniLibs/arm64-v8a"
    local out_x86_64="$PROJECT_ROOT/app/src/main/jniLibs/x86_64"
    mkdir -p "$out_arm64" "$out_x86_64"

    echo "    Compiling arm64-v8a (librastreador.so)..."
    CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC="$clang_arm64" \
        go build -buildmode=c-shared -ldflags="-s -w" \
        -o "$out_arm64/librastreador.so" "$cgo_pkg"

    echo "    Compiling x86_64 (librastreador.so)..."
    CGO_ENABLED=1 GOOS=android GOARCH=amd64 CC="$clang_x86_64" \
        go build -buildmode=c-shared -ldflags="-s -w" \
        -o "$out_x86_64/librastreador.so" "$cgo_pkg"

    # Remove temporary C header artifacts generated beside .so binaries
    rm -f "$out_arm64/librastreador.h" "$out_x86_64/librastreador.h"

    if [ ! -s "$out_arm64/librastreador.so" ] || [ ! -s "$out_x86_64/librastreador.so" ]; then
        echo "ERROR: Shared library compilation failed or generated empty binaries." >&2
        exit 1
    fi

    echo ">>> [SUCCESS] Generated native shared libraries in app/src/main/jniLibs/:"
    ls -lh "$out_arm64/librastreador.so" "$out_x86_64/librastreador.so"
}

# ------------------------------------------------------------------------------
# 3. Execution Dispatcher
# ------------------------------------------------------------------------------
case "$MODE" in
    --aar)
        build_aar
        ;;
    --so)
        build_so
        ;;
    --all)
        build_aar
        build_so
        ;;
    --help|-h)
        echo "Usage: $0 [--aar | --so | --all]"
        echo "  --aar : Compile Go engine into app/libs/rastreador-core.aar (default)"
        echo "  --so  : Compile Go engine into app/src/main/jniLibs/{abi}/librastreador.so"
        echo "  --all : Compile both --aar and --so targets"
        exit 0
        ;;
    *)
        echo "ERROR: Unknown mode '$MODE'. Use --aar, --so, or --all." >&2
        exit 1
        ;;
esac

# ------------------------------------------------------------------------------
# 4. Gradle Packaging Verification
# ------------------------------------------------------------------------------
echo ">>> [3/5] Triggering Gradle packaging verification..."
cd "$PROJECT_ROOT"
./gradlew assembleDebug --no-daemon

echo "================================================================================"
echo ">>> [COMPLETE] rastreador-android successfully linked with Go core engine!"
echo "================================================================================"
