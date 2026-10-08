package main

/*
#include "jni.h"
#include <stdlib.h>

static const char* jni_GetStringUTFChars(JNIEnv *env, jstring str) {
    if (!env || !str) return NULL;
    return (*env)->GetStringUTFChars(env, str, NULL);
}

static void jni_ReleaseStringUTFChars(JNIEnv *env, jstring str, const char *chars) {
    if (env && str && chars) {
        (*env)->ReleaseStringUTFChars(env, str, chars);
    }
}

static jstring jni_NewStringUTF(JNIEnv *env, const char *chars) {
    if (!env || !chars) return NULL;
    return (*env)->NewStringUTF(env, chars);
}
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/dodecaneser/rastreador/mobile"
)

// Helper: safely convert JNI jstring to Go string and release memory
func jstringToString(env *C.JNIEnv, jStr C.jstring) string {
	if jStr == 0 || env == nil {
		return ""
	}
	cStr := C.jni_GetStringUTFChars(env, jStr)
	if cStr == nil {
		return ""
	}
	defer C.jni_ReleaseStringUTFChars(env, jStr, cStr)
	return C.GoString(cStr)
}

// Helper: convert Go string to JNI jstring with proper memory cleanup
func stringToJstring(env *C.JNIEnv, goStr string) C.jstring {
	if env == nil {
		return 0
	}
	cStr := C.CString(goStr)
	defer C.free(unsafe.Pointer(cStr))
	return C.jni_NewStringUTF(env, cStr)
}

// Helper: two-layer panic recovery wrapper for string-based operations
func safeStringCall(env *C.JNIEnv, jsonArgs C.jstring, fn func(req string) string) (ret C.jstring) {
	defer func() {
		if r := recover(); r != nil {
			errEnvelope := fmt.Sprintf(`{"success":false,"error":"jni panic recovered: %v"}`, r)
			ret = stringToJstring(env, errEnvelope)
		}
	}()

	req := jstringToString(env, jsonArgs)
	res := fn(req)
	ret = stringToJstring(env, res)
	return ret
}

// -----------------------------------------------------------------------------
// JNI Exported Functions matching JniNativeBridgeDriver.kt
// Package: com.dodecaneser.rastreador.core
// Class:   JniNativeBridgeDriver
// -----------------------------------------------------------------------------

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeQueryBgpAsn
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeQueryBgpAsn(
	env *C.JNIEnv,
	thiz C.jobject,
	jsonArgs C.jstring,
) C.jstring {
	return safeStringCall(env, jsonArgs, mobile.QueryBgpAsn)
}

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativePerformMultilateration
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativePerformMultilateration(
	env *C.JNIEnv,
	thiz C.jobject,
	jsonArgs C.jstring,
) C.jstring {
	return safeStringCall(env, jsonArgs, mobile.PerformMultilateration)
}

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeTriangulateWiFi
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeTriangulateWiFi(
	env *C.JNIEnv,
	thiz C.jobject,
	jsonArgs C.jstring,
) C.jstring {
	return safeStringCall(env, jsonArgs, mobile.TriangulateWiFi)
}

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeAnalyzeIpId
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeAnalyzeIpId(
	env *C.JNIEnv,
	thiz C.jobject,
	jsonArgs C.jstring,
) C.jstring {
	return safeStringCall(env, jsonArgs, mobile.AnalyzeIpId)
}

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeAnalyzeTunnel
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeAnalyzeTunnel(
	env *C.JNIEnv,
	thiz C.jobject,
	jsonArgs C.jstring,
) C.jstring {
	return safeStringCall(env, jsonArgs, mobile.AnalyzeTunnel)
}

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeCalculateHaversine
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeCalculateHaversine(
	env *C.JNIEnv,
	thiz C.jobject,
	lat1 C.jdouble,
	lon1 C.jdouble,
	lat2 C.jdouble,
	lon2 C.jdouble,
) (ret C.jdouble) {
	defer func() {
		if r := recover(); r != nil {
			ret = -1.0
		}
	}()

	dist := mobile.CalculateHaversineDistance(
		float64(lat1),
		float64(lon1),
		float64(lat2),
		float64(lon2),
	)
	return C.jdouble(dist)
}

//export Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeInvokeJson
func Java_com_dodecaneser_rastreador_core_JniNativeBridgeDriver_nativeInvokeJson(
	env *C.JNIEnv,
	thiz C.jobject,
	action C.jstring,
	payloadJSON C.jstring,
) (ret C.jstring) {
	defer func() {
		if r := recover(); r != nil {
			errEnvelope := fmt.Sprintf(`{"success":false,"error":"jni panic recovered: %v"}`, r)
			ret = stringToJstring(env, errEnvelope)
		}
	}()

	act := jstringToString(env, action)
	payload := jstringToString(env, payloadJSON)
	res := mobile.InvokeJson(act, payload)
	return stringToJstring(env, res)
}

func main() {}
