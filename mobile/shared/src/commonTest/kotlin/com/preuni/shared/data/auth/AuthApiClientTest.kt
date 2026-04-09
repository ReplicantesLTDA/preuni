package com.preuni.shared.data.auth

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
import kotlin.test.assertIs
import kotlin.test.assertTrue

private fun buildClient(engine: MockEngine): HttpClient = HttpClient(engine) {
    install(ContentNegotiation) {
        json(Json { ignoreUnknownKeys = true })
    }
    defaultRequest {
        contentType(ContentType.Application.Json)
        url("http://localhost/")
    }
}

class AuthApiClientTest {

    // ── register ──────────────────────────────────────────────────────────────

    @Test
    fun register_success_returnsAuthSession() = runTest {
        val engine = MockEngine { _ ->
            respond(
                content = """{"student_id":"x","access_token":"a","refresh_token":"r","expires_in":3600}""",
                status = HttpStatusCode.Created,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = AuthApiClient(buildClient(engine))
        val result = client.register("test@example.com", "Passw0rd!", "Tester")

        assertTrue(result.isSuccess)
        val session = result.getOrThrow()
        assertTrue(session.userId == "x")
        assertTrue(session.accessToken == "a")
    }

    @Test
    fun register_conflict_returnsFailureConflict() = runTest {
        val engine = MockEngine { _ ->
            respond(
                content = """{"message":"email address is already registered"}""",
                status = HttpStatusCode.Conflict,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = AuthApiClient(buildClient(engine))
        val result = client.register("dup@example.com", "Passw0rd!", "Dup")

        assertTrue(result.isFailure)
        assertIs<AppError.Conflict>(result.exceptionOrNull())
    }

    // ── login ─────────────────────────────────────────────────────────────────

    @Test
    fun login_forbidden_returnsFailureForbidden() = runTest {
        val engine = MockEngine { _ ->
            respond(
                content = """{"message":"email address is not yet verified"}""",
                status = HttpStatusCode.Forbidden,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = AuthApiClient(buildClient(engine))
        val result = client.login("unverified@example.com", "Passw0rd!")

        assertTrue(result.isFailure)
        assertIs<AppError.Forbidden>(result.exceptionOrNull())
    }

    @Test
    fun login_unauthorized_returnsFailureUnauthorized() = runTest {
        val engine = MockEngine { _ ->
            respond(
                content = """{"message":"invalid credentials"}""",
                status = HttpStatusCode.Unauthorized,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = AuthApiClient(buildClient(engine))
        val result = client.login("user@example.com", "WrongPass!")

        assertTrue(result.isFailure)
        assertIs<AppError.Unauthorized>(result.exceptionOrNull())
    }

    // ── verifyEmail ───────────────────────────────────────────────────────────

    @Test
    fun verifyEmail_success_returnsSuccessUnit() = runTest {
        val engine = MockEngine { _ ->
            respond(
                content = "",
                status = HttpStatusCode.NoContent,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }
        val client = AuthApiClient(buildClient(engine))
        val result = client.verifyEmail("user@example.com", "123456")

        assertTrue(result.isSuccess)
    }
}
