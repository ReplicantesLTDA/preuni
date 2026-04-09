package com.preuni.shared.data.auth

import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession

class AuthRepositoryImpl(
    private val apiClient: AuthApiClient,
    private val tokenStore: TokenStore,
) : AuthRepository {

    override suspend fun verifyEmail(email: String, otp: String): Result<Unit> =
        apiClient.verifyEmail(email, otp)

    override suspend fun register(email: String, password: String, displayName: String): Result<AuthSession> {
        val result = apiClient.register(email, password, displayName)
        result.onSuccess { session -> tokenStore.save(session) }
        return result
    }

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

    override suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> =
        apiClient.changePassword(currentPassword, newPassword)

    override suspend fun changeEmailRequest(newEmail: String): Result<Unit> =
        apiClient.changeEmailRequest(newEmail)

    override suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> =
        apiClient.changeEmailConfirm(newEmail, otp)

    override suspend fun deleteAccount(): Result<Unit> =
        apiClient.deleteAccount()
}
