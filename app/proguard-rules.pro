# ==============================================================================
# Rastreador Mobile — ProGuard & R8 Optimization and Keep Rules
# Target: Android 14+ (API 34/35) | Kotlin 2.0.20 | AGP 8.5+ | R8 Full Mode
# ==============================================================================

# ------------------------------------------------------------------------------
# 1. GENERAL AND RUNTIME METADATA PRESERVATION
# ------------------------------------------------------------------------------
# Preserve annotations, signatures, and inner class relationships
-keepattributes *Annotation*,InnerClasses,EnclosingMethod,Signature,SourceFile,LineNumberTable

# Preserve source file and line numbers for de-obfuscating crash stack traces
-renamesourcefileattribute SourceFile

# ------------------------------------------------------------------------------
# 2. GOMOBILE & JNI NATIVE BINDINGS CORE
# ------------------------------------------------------------------------------
# Preserve Gomobile runtime internals (Seq, Ref, error, JNI bridge)
-keep class go.** { *; }
-keep class mobile.** { *; }
-keep class rastreador.** { *; }
-keep class com.dodecaneser.rastreador.bridge.** { *; }
-keepclassmembers class com.dodecaneser.rastreador.bridge.** { *; }

# Preserve Kotlin Go Bridge interfaces and native implementations
-keep class com.dodecaneser.rastreador.core.** { *; }
-keepclassmembers class com.dodecaneser.rastreador.core.** { *; }

# Preserve all classes with native methods and keep native method names intact
-keepclasseswithmembernames class * {
    native <methods>;
}

# Suppress harmless warnings for Gomobile internal references
-dontwarn go.**
-dontwarn mobile.**
-dontwarn rastreador.**
-dontwarn com.dodecaneser.rastreador.bridge.**

# ------------------------------------------------------------------------------
# 3. KOTLINX SERIALIZATION PRESERVATION
# ------------------------------------------------------------------------------
# Suppress compiler plugin notes
-dontnote kotlinx.serialization.SerializationExtension

# Keep companion objects that hold the generated serializer instance
-keepclassmembers class * {
    *** Companion;
}

# Keep the serializer() factory function generated on @Serializable companion objects
-keepclasseswithmembers class * {
    kotlinx.serialization.KSerializer serializer(...);
}

# Preserve fields annotated with @SerialName
-keepclassmembers class * {
    @kotlinx.serialization.SerialName <fields>;
}

# Preserve all @Serializable model classes and their default constructors
-keepclassmembers @kotlinx.serialization.Serializable class * {
    *** Companion;
    public static *** serializer();
    <init>(...);
}

# Keep custom serializer implementations
-keepclassmembers class * implements kotlinx.serialization.KSerializer {
    public static *** INSTANCE;
    <init>(...);
}

# Keep polymorphic serializer configurations
-keepattributes EnclosingMethod,InnerClasses
-dontwarn kotlinx.serialization.**

# ------------------------------------------------------------------------------
# 4. ROOM DATABASE & SQLITE PRESERVATION
# ------------------------------------------------------------------------------
# Preserve RoomDatabase abstract classes and KSP generated implementations
-keep class * extends androidx.room.RoomDatabase {
    <init>(...);
}
-keep class **_Impl {
    <init>(...);
}

# Preserve all Room entities and their constructors/field accessors
-keep @androidx.room.Entity class * {
    <init>(...);
    <fields>;
}
-keepclassmembers @androidx.room.Entity class * { *; }

# Preserve all DAO interfaces and their generated implementation methods
-keep @androidx.room.Dao interface * { *; }
-keep class * implements androidx.room.Dao { *; }

# Suppress optional Room paging warnings if paging library is not linked
-dontwarn androidx.room.paging.**
-dontwarn androidx.sqlite.db.**

# ------------------------------------------------------------------------------
# 5. OSMDROID MAP ENGINE PRESERVATION
# ------------------------------------------------------------------------------
# OsmDroid uses extensive reflection to instantiate tile providers and overlays
-keep class org.osmdroid.** { *; }
-keepclassmembers class org.osmdroid.** { *; }
-keep public class * extends org.osmdroid.views.overlay.Overlay {
    public <init>(...);
}
-keep public class * implements org.osmdroid.tileprovider.tilesource.ITileSource {
    public <init>(...);
}

# Suppress warnings related to legacy Apache HTTP or deprecated Android classes in OsmDroid
-dontwarn org.osmdroid.**
-dontwarn org.apache.http.**

# ------------------------------------------------------------------------------
# 6. JETPACK COMPOSE & MATERIAL 3
# ------------------------------------------------------------------------------
# Preserve Compose runtime lambdas and stability markers
-keepclassmembers class * {
    @androidx.compose.runtime.Composable *;
    androidx.compose.runtime.RecomposeScopeImpl *;
}

# Preserve Compose tooling for debug builds
-dontwarn androidx.compose.**

# ------------------------------------------------------------------------------
# 7. KOTLIN COROUTINES & FLOW
# ------------------------------------------------------------------------------
# Preserve ServiceLoader entry points for Coroutines internal dispatchers
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}

# Preserve volatile fields required for atomic thread-safety
-keepclassmembernames class kotlinx.coroutines.** {
    volatile <fields>;
}

# Suppress internal reflection warnings
-dontwarn kotlinx.coroutines.**

# ------------------------------------------------------------------------------
# 8. HARDWARE SENSORS & PLAY SERVICES GNSS FALLBACK
# ------------------------------------------------------------------------------
# Keep Google Play Services Location API classes if present
-keep class com.google.android.gms.location.** { *; }
-keep class com.google.android.gms.common.** { *; }
-dontwarn com.google.android.gms.**

# Preserve AOSP GnssStatus callback reflection
-keep class android.location.GnssStatus$Callback { *; }
-keep class android.telephony.TelephonyCallback$* { *; }
