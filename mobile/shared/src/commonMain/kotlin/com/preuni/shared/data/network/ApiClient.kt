package com.preuni.shared.data.network

import com.preuni.shared.domain.error.AppError
import io.ktor.client.HttpClient
import io.ktor.client.plugins.HttpSend
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.plugins.plugin
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlin.time.measureTime

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
        // Suppress header/body logging to prevent Authorization values and response PII
        // from appearing in logcat/console. Structured request telemetry is provided via
        // attachTelemetry() which explicitly excludes sensitive fields.
        level = LogLevel.NONE
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

// ── Telemetry ─────────────────────────────────────────────────────────────────

/** Outcome category for a single request attempt. */
enum class RequestOutcome {
    SUCCESS,
    TRANSIENT_FAILURE,
    NON_TRANSIENT_FAILURE,
    TIMEOUT,
}

/**
 * Per-attempt request metrics. Contains no PII — Authorization header values and
 * response bodies are never included.
 */
data class RequestTelemetry(
    val route: String,
    val method: String,
    /** HTTP status code, or null when a transport-level failure prevented any response. */
    val statusCode: Int?,
    val attempt: Int,
    val durationMs: Long,
    val outcome: RequestOutcome,
) {
    override fun toString(): String =
        "Telemetry[method=$method route=$route status=$statusCode attempt=${attempt} duration=${durationMs}ms outcome=$outcome]"
}

/**
 * Installs a lightweight telemetry interceptor on this [HttpClient].
 *
 * For each completed request attempt the interceptor records route, method, HTTP
 * status, elapsed time, and outcome category. No Authorization header values or
 * response body content are ever captured, ensuring PII stays out of logs.
 *
 * Must be called after client construction (like [installTokenRefreshInterceptor]).
 *
 * @param onTelemetry Callback invoked after each request attempt. Defaults to a
 *                    structured `println` for debug builds. Use a no-op in production
 *                    until an analytics backend is wired.
 */
fun HttpClient.attachTelemetry(onTelemetry: (RequestTelemetry) -> Unit = ::logTelemetry) {
    plugin(HttpSend).intercept { request ->
        var call: io.ktor.client.call.HttpClientCall
        var statusCode: Int? = null
        var outcome: RequestOutcome

        val duration = measureTime {
            call = try {
                execute(request).also { c ->
                    statusCode = c.response.status.value
                }
            } catch (e: Exception) {
                outcome = RequestOutcome.TIMEOUT
                throw e
            }
        }

        outcome = when {
            statusCode == null -> RequestOutcome.TIMEOUT
            statusCode!! in 200..299 -> RequestOutcome.SUCCESS
            statusCode!! in 500..599 -> RequestOutcome.TRANSIENT_FAILURE
            else -> RequestOutcome.NON_TRANSIENT_FAILURE
        }

        onTelemetry(
            RequestTelemetry(
                route = request.url.build().encodedPath,
                method = request.method.value,
                statusCode = statusCode,
                attempt = 0,
                durationMs = duration.inWholeMilliseconds,
                outcome = outcome,
            )
        )
        call
    }
}

private fun logTelemetry(telemetry: RequestTelemetry) {
    println(telemetry.toString())
}
