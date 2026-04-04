package com.preuni.shared.data.auth

import com.preuni.shared.data.network.toAppError
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.error.AppError
import io.ktor.client.HttpClient
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.http.isSuccess
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

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

class AuthApiClient(private val httpClient: HttpClient) {

    suspend fun login(email: String, password: String): Result<AuthSession> {
        return runCatching {
            val response: HttpResponse = httpClient.post("v1/auth/login") {
                setBody(LoginRequest(email, password))
            }
            if (!response.status.isSuccess()) {
                throw response.toAppError()
            }
            val body = io.ktor.client.call.body<AuthResponse>(response)
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
            val body = io.ktor.client.call.body<AuthResponse>(response)
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
}

private fun <T> Result<T>.mapFailure(): Result<T> = this.recoverCatching { e ->
    throw if (e is AppError) e else AppError.NetworkError(e.message ?: "Network error")
}
