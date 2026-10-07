#!/usr/bin/env bash
set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

GRADLE_BIN="/Users/dodecaneser/.gradle/wrapper/dists/gradle-8.11.1-bin/bpt9gzteqjrbo1mjrsomdt32c/gradle-8.11.1"
GRADLE_LIBS="${GRADLE_BIN}/lib/*"
KOTLIN_STDLIB="${GRADLE_BIN}/lib/kotlin-stdlib-2.0.20.jar"
BUILD_DIR="${PROJECT_ROOT}/build/test-classes"

mkdir -p "$BUILD_DIR"

echo "[*] Compiling E2E Acceptance Test Suites (Kotlin 2.0.20 Target 17)..."
java -cp "$GRADLE_LIBS" org.jetbrains.kotlin.cli.jvm.K2JVMCompiler \
  -cp "$KOTLIN_STDLIB" \
  -d "$BUILD_DIR" \
  app/src/test/java/com/dodecaneser/rastreador/e2e/model/OpaqueModels.kt \
  app/src/test/java/com/dodecaneser/rastreador/e2e/model/ReferenceOracle.kt \
  app/src/test/java/com/dodecaneser/rastreador/e2e/Tier1FeatureCoverageTest.kt \
  app/src/test/java/com/dodecaneser/rastreador/e2e/Tier2BoundaryCornerCaseTest.kt \
  app/src/test/java/com/dodecaneser/rastreador/e2e/Tier3CrossFeaturePairwiseTest.kt \
  app/src/test/java/com/dodecaneser/rastreador/e2e/Tier4RealWorldScenariosTest.kt \
  app/src/test/java/com/dodecaneser/rastreador/e2e/TestRunner.kt

echo "[*] Launching Test Runner..."
java -cp "${BUILD_DIR}:${GRADLE_LIBS}" com.dodecaneser.rastreador.e2e.TestRunnerKt
