package com.preuni.shared.data.user

import com.preuni.shared.domain.error.AppError
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
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

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
    {
      "id": "uid-1",
      "display_name": "Ana Lima",
      "username": "ana01",
      "email": "ana@example.com",
      "avatar_url": null,
      "xp_total": 120,
      "streak_count": 4,
      "readiness_score": 0.63,
      "onboarding_completed": true
    }
""".trimIndent()

class UserApiClientTest {

    // ── Auth header contract ──────────────────────────────────────────────────

    @Test
    fun `getMe sends Authorization header when token is set`() = runTest {
        var capturedAuth: String? = null
        val engine = MockEngine { req ->
            capturedAuth = req.headers["Authorization"]
            respond(validStudentJson, HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")))
        }
        val client = UserApiClient(buildClient(engine, accessToken = "my-token"))
        client.getMe()

        assertNotNull(capturedAuth)
        assertTrue(capturedAuth!!.startsWith("Bearer "), "Expected Bearer scheme")
        assertTrue(capturedAuth!!.contains("my-token"))
    }

    @Test
    fun `getMe returns 401 as AppError Unauthorized`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"error":{"code":"UNAUTHORIZED","message":"token expired"}}""",
                HttpStatusCode.Unauthorized,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).getMe()
        assertIs<AppError.Unauthorized>(result.exceptionOrNull())
    }

    // ── Success schema mapping ─────────────────────────────────────────────────

    @Test
    fun `getMe success maps all required fields without loss`() = runTest {
        val engine = MockEngine { _ ->
            respond(validStudentJson, HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")))
        }
        val result = UserApiClient(buildClient(engine)).getMe()

        assertTrue(result.isSuccess)
        val student = result.getOrThrow()
        assertEquals("uid-1", student.id)
        assertEquals("Ana Lima", student.displayName)
        assertEquals("ana01", student.username)
        assertEquals("ana@example.com", student.email)
        assertNull(student.avatarUrl)
        assertEquals(120L, student.xpTotal)
        assertEquals(4, student.streakCount)
        assertEquals(0.63, student.readinessScore)
        assertTrue(student.onboardingCompleted)
    }

    // ── Schema validation: missing required fields ────────────────────────────

    @Test
    fun `getMe rejects payload with blank id`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"id":"","display_name":"Ana","username":"ana","email":"a@b.com","xp_total":0,"streak_count":0,"readiness_score":0.5,"onboarding_completed":false}""",
                HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).getMe()
        assertTrue(result.isFailure, "Blank id should cause validation failure")
        assertIs<AppError>(result.exceptionOrNull())
    }

    @Test
    fun `getMe rejects payload with blank display_name`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"id":"u1","display_name":"","username":"ana","email":"a@b.com","xp_total":0,"streak_count":0,"readiness_score":0.5,"onboarding_completed":false}""",
                HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).getMe()
        assertTrue(result.isFailure, "Blank display_name should cause validation failure")
    }

    @Test
    fun `getMe rejects payload with negative xp_total`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"id":"u1","display_name":"A","username":"a","email":"a@b.com","xp_total":-1,"streak_count":0,"readiness_score":0.5,"onboarding_completed":false}""",
                HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).getMe()
        assertTrue(result.isFailure, "Negative xp_total should cause validation failure")
    }

    @Test
    fun `getMe rejects payload with readiness_score out of range`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"id":"u1","display_name":"A","username":"a","email":"a@b.com","xp_total":0,"streak_count":0,"readiness_score":1.5,"onboarding_completed":false}""",
                HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).getMe()
        assertTrue(result.isFailure, "readiness_score > 1.0 should cause validation failure")
    }

    // ── Network/transport errors ───────────────────────────────────────────────

    @Test
    fun `getMe maps 5xx to AppError`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"error":{"code":"INTERNAL_ERROR","message":"oops"}}""",
                HttpStatusCode.InternalServerError,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).getMe()
        assertTrue(result.isFailure)
        assertIs<AppError>(result.exceptionOrNull())
    }

    // ── Profile update schema contract ────────────────────────────────────────

    @Test
    fun `updateProfile success maps full response schema`() = runTest {
        val engine = MockEngine { _ ->
            respond(validStudentJson, HttpStatusCode.OK,
                headersOf("Content-Type" to listOf("application/json")))
        }
        val result = UserApiClient(buildClient(engine)).updateProfile("Ana Lima", "ana01")
        assertTrue(result.isSuccess)
        assertEquals("uid-1", result.getOrThrow().id)
    }

    @Test
    fun `updateProfile validation failure returns AppError Validation`() = runTest {
        val engine = MockEngine { _ ->
            respond(
                """{"error":{"code":"VALIDATION_ERROR","message":"username taken","field":"username"}}""",
                HttpStatusCode.UnprocessableEntity,
                headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val result = UserApiClient(buildClient(engine)).updateProfile(null, "taken")
        assertIs<AppError.Validation>(result.exceptionOrNull())
        assertEquals("username", (result.exceptionOrNull() as AppError.Validation).field)
    }
}
