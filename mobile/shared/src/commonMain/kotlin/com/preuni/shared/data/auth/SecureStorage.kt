package com.preuni.shared.data.auth

/** Platform-specific secure key-value storage (Keychain on iOS, EncryptedSharedPreferences on Android, localStorage on Web). */
expect class SecureStorage() {
    fun put(key: String, value: String)
    fun get(key: String): String?
    fun remove(key: String)
    fun clear()
}
