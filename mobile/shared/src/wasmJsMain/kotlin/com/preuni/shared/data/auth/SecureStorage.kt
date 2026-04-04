package com.preuni.shared.data.auth

import kotlinx.browser.window

// Web uses sessionStorage (tab-scoped) as a best-effort; for production
// consider a SubtleCrypto-backed store. Tokens are never written to localStorage
// to reduce XSS exposure surface.
actual class SecureStorage actual constructor() {

    actual fun put(key: String, value: String) {
        window.sessionStorage.setItem(key, value)
    }

    actual fun get(key: String): String? = window.sessionStorage.getItem(key)

    actual fun remove(key: String) {
        window.sessionStorage.removeItem(key)
    }

    actual fun clear() {
        window.sessionStorage.clear()
    }
}
