package com.preuni.shared.data.auth

import com.preuni.shared.domain.auth.AuthSession

private const val KEY_USER_ID = "user_id"
private const val KEY_WELCOME_SEEN = "welcome_seen"
private const val KEY_USER_EMAIL = "user_email"
private const val KEY_ACCESS_TOKEN = "access_token"
private const val KEY_REFRESH_TOKEN = "refresh_token"

class TokenStore(private val storage: SecureStorage) {

    fun save(session: AuthSession) {
        storage.put(KEY_USER_ID, session.userId)
        storage.put(KEY_USER_EMAIL, session.userEmail)
        storage.put(KEY_ACCESS_TOKEN, session.accessToken)
        storage.put(KEY_REFRESH_TOKEN, session.refreshToken)
    }

    fun load(): AuthSession? {
        val userId = storage.get(KEY_USER_ID) ?: return null
        val userEmail = storage.get(KEY_USER_EMAIL) ?: return null
        val accessToken = storage.get(KEY_ACCESS_TOKEN) ?: return null
        val refreshToken = storage.get(KEY_REFRESH_TOKEN) ?: return null
        return AuthSession(userId, userEmail, accessToken, refreshToken)
    }

    fun clear() = storage.clear()

    fun isLoggedIn(): Boolean = storage.get(KEY_ACCESS_TOKEN) != null

    fun welcomeSeen(): Boolean = storage.get(KEY_WELCOME_SEEN) != null
    fun markWelcomeSeen() = storage.put(KEY_WELCOME_SEEN, "true")

    fun accessToken(): String? = storage.get(KEY_ACCESS_TOKEN)

    fun refreshToken(): String? = storage.get(KEY_REFRESH_TOKEN)

    fun updateTokens(accessToken: String, refreshToken: String) {
        storage.put(KEY_ACCESS_TOKEN, accessToken)
        storage.put(KEY_REFRESH_TOKEN, refreshToken)
    }
}
