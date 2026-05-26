package com.preuni.shared.domain.auth

interface AuthRepository {
    suspend fun verifyEmail(email: String, otp: String): Result<Unit>
    suspend fun resendVerificationOtp(email: String): Result<Unit>
    suspend fun register(email: String, password: String, displayName: String): Result<AuthSession>
    suspend fun login(emailOrUsername: String, password: String): Result<AuthSession>
    suspend fun logout()
    fun isLoggedIn(): Boolean
    suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit>
    suspend fun changeEmailRequest(newEmail: String): Result<Unit>
    suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit>
    suspend fun deleteAccount(): Result<Unit>
}
