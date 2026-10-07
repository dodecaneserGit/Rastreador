package com.dodecaneser.rastreador.e2e

import java.lang.reflect.InvocationTargetException
import java.lang.reflect.Method

/**
 * Standalone E2E Acceptance Test Suite Runner.
 * Discovers and executes all test cases across Tiers 1-4, producing a comprehensive
 * execution report and exit code semantics.
 */
object TestRunner {

    data class TierResult(
        val tierName: String,
        val total: Int,
        val passed: Int,
        val failed: Int,
        val failures: List<String>,
        val durationMs: Long
    )

    fun runAll(): Boolean {
        println("================================================================================")
        println("  RASTREADOR MOBILE — E2E OPAQUE-BOX ACCEPTANCE TEST SUITE")
        println("================================================================================")
        println("Runtime: Java ${System.getProperty("java.version")} (${System.getProperty("java.vendor")})")
        println("Target OS: Android 14/15 (API 34/35) — Kotlin 2.0.20 Target 17")
        println("--------------------------------------------------------------------------------\n")

        val results = mutableListOf<TierResult>()
        val overallStart = System.currentTimeMillis()

        results.add(executeSuite("Tier 1: Feature Coverage (F01-F17)", Tier1FeatureCoverageTest::class.java))
        results.add(executeSuite("Tier 2: Boundary & Corner Cases (F01-F17)", Tier2BoundaryCornerCaseTest::class.java))
        results.add(executeSuite("Tier 3: Cross-Feature Interactions", Tier3CrossFeaturePairwiseTest::class.java))
        results.add(executeSuite("Tier 4: Real-World Scenarios", Tier4RealWorldScenariosTest::class.java))

        val overallDuration = System.currentTimeMillis() - overallStart

        println("\n================================================================================")
        println("  TEST SUITE EXECUTION SUMMARY")
        println("================================================================================")
        var grandTotal = 0
        var grandPassed = 0
        var grandFailed = 0

        for (r in results) {
            grandTotal += r.total
            grandPassed += r.passed
            grandFailed += r.failed
            val status = if (r.failed == 0) "PASSED" else "FAILED"
            println(String.format("  %-42s : %3d / %3d %-6s (%d ms)", r.tierName, r.passed, r.total, status, r.durationMs))
        }

        println("--------------------------------------------------------------------------------")
        println(String.format("  TOTAL ACCEPTANCE TESTS EXECUTED            : %3d", grandTotal))
        println(String.format("  TOTAL PASSED                               : %3d", grandPassed))
        println(String.format("  TOTAL FAILED                               : %3d", grandFailed))
        println(String.format("  OVERALL DURATION                           : %d ms", overallDuration))
        println("================================================================================")

        if (grandFailed > 0) {
            println("\n[!] TEST SUITE FAILED WITH $grandFailed FAILURES:")
            for (r in results) {
                for (f in r.failures) {
                    println("  - [${r.tierName}] $f")
                }
            }
            return false
        } else {
            println("\n[+] 100% ACCEPTANCE TESTS PASSED SUCCESSFULLY.")
            return true
        }
    }

    private fun executeSuite(tierName: String, testClass: Class<*>): TierResult {
        print("Executing $tierName ... ")
        System.out.flush()
        val start = System.currentTimeMillis()
        val instance = testClass.getDeclaredConstructor().newInstance()
        val methods = testClass.declaredMethods
            .filter { it.name.startsWith("test_") && !it.name.contains("$") && it.parameterCount == 0 }
            .sortedBy { it.name }

        var passed = 0
        var failed = 0
        val failures = mutableListOf<String>()

        for (m in methods) {
            m.isAccessible = true
            try {
                m.invoke(instance)
                passed++
            } catch (ite: InvocationTargetException) {
                failed++
                val cause = ite.cause ?: ite
                failures.add("${m.name}: ${cause.message}")
            } catch (e: Throwable) {
                failed++
                failures.add("${m.name}: ${e.message}")
            }
        }

        val duration = System.currentTimeMillis() - start
        if (failed == 0) {
            println("OK ($passed tests in ${duration}ms)")
        } else {
            println("FAIL ($passed passed, $failed failed in ${duration}ms)")
        }

        return TierResult(tierName, methods.size, passed, failed, failures, duration)
    }
}

fun main() {
    val success = TestRunner.runAll()
    if (!success) {
        System.exit(1)
    }
}
