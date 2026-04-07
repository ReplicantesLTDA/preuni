package com.preuni.shared.data.auth

/** JVM/Desktop stub — backed by an in-memory map. Used only in tests. */
actual class SecureStorage actual constructor() {
    private val store = mutableMapOf<String, String>()
    actual fun put(key: String, value: String) { store[key] = value }
    actual fun get(key: String): String? = store[key]
    actual fun remove(key: String) { store.remove(key) }
    actual fun clear() { store.clear() }
}
