package com.preuni.shared.data.auth

import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession

class AuthRepositoryImpl(
    private val apiClient: AuthApiClient,
    private val tokenStore: TokenStore,
) : AuthRepository {

    override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> {
        val result = apiClient.login(emailOrUsername, password)
        result.onSuccess { session -> tokenStore.save(session) }
        return result
    }

    override suspend fun logout() {
        val refreshToken = tokenStore.refreshToken()
        if (refreshToken != null) {
            apiClient.logout(refreshToken) // best-effort; ignore failure
        }
        tokenStore.clear()
    }

    override fun isLoggedIn(): Boolean = tokenStore.isLoggedIn()
}
