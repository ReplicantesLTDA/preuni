package com.preuni.shared.domain.auth

interface AuthRepository {
    suspend fun verifyEmail(email: String, otp: String): Result<Unit>
    suspend fun register(email: String, password: String, displayName: String): Result<AuthSession>
    suspend fun login(emailOrUsername: String, password: String): Result<AuthSession>
    suspend fun logout()
    fun isLoggedIn(): Boolean
}
