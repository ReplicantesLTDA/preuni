package com.preuni.shared.data.network

import com.preuni.shared.domain.error.AppError
import io.ktor.client.HttpClient
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class ApiErrorBody(
    val field: String? = null,
    val message: String = "Unknown error",
)

fun buildHttpClient(baseUrl: String, getAccessToken: (() -> String?)? = null): HttpClient = HttpClient {
    install(ContentNegotiation) {
        json(Json {
            ignoreUnknownKeys = true
            isLenient = true
        })
    }

    install(HttpTimeout) {
        requestTimeoutMillis = 30_000
        connectTimeoutMillis = 10_000
        socketTimeoutMillis = 30_000
    }

    install(Logging) {
        level = LogLevel.INFO
    }

    defaultRequest {
        contentType(ContentType.Application.Json)
        url(baseUrl)
        getAccessToken?.invoke()?.let { token ->
            headers.append("Authorization", "Bearer $token")
        }
    }
}

suspend fun HttpResponse.toAppError(): AppError {
    val body = runCatching {
        val text = bodyAsText()
        Json.decodeFromString<ApiErrorBody>(text)
    }.getOrNull()

    return when (status) {
        HttpStatusCode.Unauthorized -> AppError.Unauthorized()
        HttpStatusCode.Forbidden -> AppError.Forbidden(body?.message ?: "Access denied.")
        HttpStatusCode.Conflict -> AppError.Conflict(body?.message ?: "Conflict")
        HttpStatusCode.UnprocessableEntity ->
            AppError.Validation(
                field = body?.field ?: "unknown",
                message = body?.message ?: "Validation failed",
            )
        else -> AppError.Unknown(body?.message ?: "Something went wrong. Please try again.")
    }
}
