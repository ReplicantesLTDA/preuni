package com.preuni.shared.data.auth

import com.preuni.shared.data.network.toAppError
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.error.AppError
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.delete
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.http.isSuccess
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class VerifyEmailRequest(
    val email: String,
    val otp: String,
)

@Serializable
data class ResendVerificationRequest(
    val email: String,
)

@Serializable
data class RegisterRequest(
    val email: String,
    val password: String,
    @SerialName("display_name") val displayName: String,
)

@Serializable
data class LoginRequest(
    val email: String,
    val password: String,
)

@Serializable
data class AuthResponse(
    @SerialName("student_id") val studentId: String,
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("expires_in") val expiresIn: Int,
)

@Serializable
data class RefreshRequest(
    @SerialName("refresh_token") val refreshToken: String,
)

@Serializable
data class ChangePasswordRequest(
    @SerialName("current_password") val currentPassword: String,
    @SerialName("new_password") val newPassword: String,
)

@Serializable
data class ChangeEmailRequestBody(
    @SerialName("new_email") val newEmail: String,
)

@Serializable
data class ChangeEmailConfirmRequest(
    @SerialName("new_email") val newEmail: String,
    val otp: String,
)

class AuthApiClient(private val httpClient: HttpClient) {

    suspend fun verifyEmail(email: String, otp: String): Result<Unit> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/email/verify") {
                setBody(VerifyEmailRequest(email, otp))
            }
            if (!response.status.isSuccess()) {
                throw response.toAppError()
            }
        }.mapFailure()
    }

    suspend fun resendVerificationOtp(email: String): Result<Unit> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/email/verify-resend") {
                setBody(ResendVerificationRequest(email))
            }
            if (!response.status.isSuccess()) {
                throw response.toAppError()
            }
        }.mapFailure()
    }

    suspend fun register(email: String, password: String, displayName: String): Result<AuthSession> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/register") {
                setBody(RegisterRequest(email, password, displayName))
            }
            if (!response.status.isSuccess()) {
                throw response.toAppError()
            }
            val body: AuthResponse = response.body()
            AuthSession(
                userId = body.studentId,
                userEmail = email,
                accessToken = body.accessToken,
                refreshToken = body.refreshToken,
            )
        }.mapFailure()
    }

    suspend fun login(email: String, password: String): Result<AuthSession> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/login") {
                setBody(LoginRequest(email, password))
            }
            if (!response.status.isSuccess()) {
                throw response.toAppError()
            }
            val body: AuthResponse = response.body()
            AuthSession(
                userId = body.studentId,
                userEmail = email,
                accessToken = body.accessToken,
                refreshToken = body.refreshToken,
            )
        }.mapFailure()
    }

    suspend fun refresh(refreshToken: String): Result<Pair<String, String>> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/refresh") {
                setBody(RefreshRequest(refreshToken))
            }
            if (!response.status.isSuccess()) {
                throw response.toAppError()
            }
            val body: AuthResponse = response.body()
            Pair(body.accessToken, body.refreshToken)
        }.mapFailure()
    }

    suspend fun logout(refreshToken: String): Result<Unit> {
        return runCatching {
            httpClient.post("v1/auth/logout") {
                setBody(RefreshRequest(refreshToken))
            }
            Unit
        }.mapFailure()
    }

    suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/password/change") {
                setBody(ChangePasswordRequest(currentPassword, newPassword))
            }
            if (!response.status.isSuccess()) throw response.toAppError()
        }.mapFailure()
    }

    suspend fun changeEmailRequest(newEmail: String): Result<Unit> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/email/change/request") {
                setBody(ChangeEmailRequestBody(newEmail))
            }
            if (!response.status.isSuccess()) throw response.toAppError()
        }.mapFailure()
    }

    suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/email/change/confirm") {
                setBody(ChangeEmailConfirmRequest(newEmail, otp))
            }
            if (!response.status.isSuccess()) throw response.toAppError()
        }.mapFailure()
    }

    suspend fun deleteAccount(): Result<Unit> {
        return runCatching {
            val response: HttpResponse = httpClient.delete("v1/auth/account")
            if (!response.status.isSuccess()) throw response.toAppError()
        }.mapFailure()
    }
}

private fun <T> Result<T>.mapFailure(): Result<T> = this.recoverCatching { e ->
    throw if (e is AppError) e else AppError.NetworkError(e.message ?: "Network error")
}
