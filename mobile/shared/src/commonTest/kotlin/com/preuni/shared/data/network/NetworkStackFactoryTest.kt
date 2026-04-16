package com.preuni.shared.data.network

import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.data.user.UserApiClient
import com.preuni.shared.data.user.UserRepositoryImpl
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
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

/**
 * Tests for NetworkStackFactory — verifies that the shared authenticated client
 * bootstrap correctly injects bearer tokens on every request.
 */
class NetworkStackFactoryTest {

    /** Builds a test HttpClient backed by MockEngine using the same config as the factory. */
    private fun buildTestHttpClient(
        engine: MockEngine,
        getAccessToken: (() -> String?)? = null,
    ): HttpClient = HttpClient(engine) {
        install(ContentNegotiation) {
            json(Json { ignoreUnknownKeys = true; isLenient = true })
        }
        defaultRequest {
            contentType(ContentType.Application.Json)
            url("http://localhost/")
            getAccessToken?.invoke()?.let { token ->
                headers.append("Authorization", "Bearer $token")
            }
        }
    }

    // ── Token injection ────────────────────────────────────────────────────────

    @Test
    fun `authenticated client injects bearer token on requests`() {
        var capturedAuthHeader: String? = null
        val engine = MockEngine { req ->
            capturedAuthHeader = req.headers["Authorization"]
            respond(
                content = """{"id":"u1","display_name":"A","username":"a","email":"a@b.com","xp_total":0,"streak_count":0,"readiness_score":0.0,"onboarding_completed":false}""",
                status = HttpStatusCode.OK,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }

        val storage = SecureStorage()
        val tokenStore = TokenStore(storage)
        tokenStore.updateTokens("test-access-token", "refresh")

        val client = buildTestHttpClient(engine, getAccessToken = { tokenStore.accessToken() })
        val repo = UserRepositoryImpl(UserApiClient(client))

        kotlinx.coroutines.test.runTest {
            repo.getMe()
        }

        assertNotNull(capturedAuthHeader, "Authorization header must be set")
        assertTrue(
            capturedAuthHeader!!.startsWith("Bearer "),
            "Authorization must use Bearer scheme, got: $capturedAuthHeader",
        )
        assertTrue(
            capturedAuthHeader!!.contains("test-access-token"),
            "Bearer token must match stored access token",
        )
    }

    @Test
    fun `unauthenticated client omits authorization header`() {
        var capturedAuthHeader: String? = "sentinel"
        val engine = MockEngine { req ->
            capturedAuthHeader = req.headers["Authorization"]
            respond(
                content = """{"id":"u1","display_name":"A","username":"a","email":"a@b.com","xp_total":0,"streak_count":0,"readiness_score":0.0,"onboarding_completed":false}""",
                status = HttpStatusCode.OK,
                headers = headersOf("Content-Type" to listOf("application/json")),
            )
        }

        // No token stored — getAccessToken returns null
        val client = buildTestHttpClient(engine, getAccessToken = { null })
        val repo = UserRepositoryImpl(UserApiClient(client))

        kotlinx.coroutines.test.runTest {
            repo.getMe()
        }

        assertTrue(
            capturedAuthHeader == null,
            "Authorization header must be absent when no token is stored",
        )
    }

    // ── NetworkStackFactory ────────────────────────────────────────────────────

    @Test
    fun `NetworkStackFactory create returns non-null dependencies`() {
        val storage = SecureStorage()
        val stack = NetworkStackFactory.create(
            baseUrl = "http://localhost/",
            secureStorage = storage,
        )

        assertNotNull(stack.tokenStore)
        assertNotNull(stack.authRepository)
        assertNotNull(stack.userRepository)
        assertNotNull(stack.contentRepository)
    }
}
