package com.preuni.shared.data.network

import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.http.headersOf
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

/**
 * Tests for request telemetry and PII redaction in the network layer.
 */
class ApiClientTelemetryTest {

    private fun buildClient(engine: MockEngine, accessToken: String? = null): HttpClient =
        HttpClient(engine) {
            install(ContentNegotiation) {
                json(Json { ignoreUnknownKeys = true; isLenient = true })
            }
            defaultRequest {
                contentType(ContentType.Application.Json)
                url("http://localhost/")
                accessToken?.let { headers.append("Authorization", "Bearer $it") }
            }
        }

    private val validStudentJson = """
        {"id":"u1","display_name":"A","username":"a","email":"sensitive@user.com","xp_total":0,"streak_count":0,"readiness_score":0.0,"onboarding_completed":false}
    """.trimIndent()

    // ── PII redaction ─────────────────────────────────────────────────────────

    @Test
    fun `telemetry does not include Authorization header value`() {
        val logLines = mutableListOf<String>()
        val engine = MockEngine { _ ->
            respond(validStudentJson, HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")))
        }
        val client = buildClient(engine, accessToken = "super-secret-token")
        client.attachTelemetry { telemetry -> logLines.add(telemetry.toString()) }

        kotlinx.coroutines.test.runTest {
            com.preuni.shared.data.user.UserApiClient(client).getMe()
        }

        assertTrue(logLines.isNotEmpty(), "Telemetry must be emitted")
        logLines.forEach { line ->
            assertFalse(
                line.contains("super-secret-token"),
                "Telemetry must not log the access token value",
            )
        }
    }

    @Test
    fun `telemetry does not include response body content`() {
        val logLines = mutableListOf<String>()
        val engine = MockEngine { _ ->
            respond(validStudentJson, HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")))
        }
        val client = buildClient(engine)
        client.attachTelemetry { telemetry -> logLines.add(telemetry.toString()) }

        kotlinx.coroutines.test.runTest {
            com.preuni.shared.data.user.UserApiClient(client).getMe()
        }

        logLines.forEach { line ->
            assertFalse(
                line.contains("sensitive@user.com"),
                "Telemetry must not include email from response body",
            )
        }
    }

    // ── Telemetry fields ─────────────────────────────────────────────────────

    @Test
    fun `successful request emits telemetry with all required fields`() {
        var captured: RequestTelemetry? = null
        val engine = MockEngine { _ ->
            respond(validStudentJson, HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")))
        }
        val client = buildClient(engine)
        client.attachTelemetry { captured = it }

        kotlinx.coroutines.test.runTest {
            com.preuni.shared.data.user.UserApiClient(client).getMe()
        }

        assertNotNull(captured, "RequestTelemetry must be emitted for a successful call")
        assertNotNull(captured!!.route)
        assertNotNull(captured!!.method)
        assertNotNull(captured!!.statusCode)
        assertTrue(captured!!.durationMs >= 0)
        assertTrue(captured!!.outcome == RequestOutcome.SUCCESS)
    }

    @Test
    fun `401 response emits telemetry with NON_TRANSIENT_FAILURE outcome`() {
        var captured: RequestTelemetry? = null
        val engine = MockEngine { _ ->
            respond(
                """{"error":{"code":"UNAUTHORIZED","message":"expired"}}""",
                HttpStatusCode.Unauthorized,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = buildClient(engine)
        client.attachTelemetry { captured = it }

        kotlinx.coroutines.test.runTest {
            com.preuni.shared.data.user.UserApiClient(client).getMe()
        }

        assertNotNull(captured)
        assertTrue(
            captured!!.outcome == RequestOutcome.NON_TRANSIENT_FAILURE,
            "401 should produce NON_TRANSIENT_FAILURE, got: ${captured!!.outcome}",
        )
    }

    @Test
    fun `5xx response emits telemetry with TRANSIENT_FAILURE outcome`() {
        var captured: RequestTelemetry? = null
        val engine = MockEngine { _ ->
            respond(
                """{"error":{"code":"INTERNAL_ERROR","message":"oops"}}""",
                HttpStatusCode.InternalServerError,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = buildClient(engine)
        client.attachTelemetry { captured = it }

        kotlinx.coroutines.test.runTest {
            com.preuni.shared.data.user.UserApiClient(client).getMe()
        }

        assertNotNull(captured)
        assertTrue(
            captured!!.outcome == RequestOutcome.TRANSIENT_FAILURE,
            "5xx should produce TRANSIENT_FAILURE, got: ${captured!!.outcome}",
        )
    }
}
