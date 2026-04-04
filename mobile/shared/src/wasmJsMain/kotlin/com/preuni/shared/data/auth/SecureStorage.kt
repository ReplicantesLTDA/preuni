package com.preuni.shared.data.auth

// NOTE: Web uses sessionStorage (tab-scoped) as a best-effort; for production
// consider a SubtleCrypto-backed store. Tokens are never written to localStorage
// to reduce XSS exposure surface.
actual class SecureStorage actual constructor() {

    actual fun put(key: String, value: String) {
        sessionStoragePut(key, value)
    }

    actual fun get(key: String): String? = sessionStorageGet(key)

    actual fun remove(key: String) {
        sessionStorageRemove(key)
    }

    actual fun clear() {
        sessionStorageClear()
    }
}

private fun sessionStoragePut(key: String, value: String): Unit =
    js("window.sessionStorage.setItem(key, value)")

private fun sessionStorageGet(key: String): String? =
    js("window.sessionStorage.getItem(key)")

private fun sessionStorageRemove(key: String): Unit =
    js("window.sessionStorage.removeItem(key)")

private fun sessionStorageClear(): Unit =
    js("window.sessionStorage.clear()")
