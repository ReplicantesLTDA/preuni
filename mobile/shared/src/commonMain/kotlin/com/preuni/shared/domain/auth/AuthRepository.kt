package com.preuni.shared.domain.auth

interface AuthRepository {
    suspend fun login(emailOrUsername: String, password: String): Result<AuthSession>
    suspend fun logout()
    fun isLoggedIn(): Boolean
}
