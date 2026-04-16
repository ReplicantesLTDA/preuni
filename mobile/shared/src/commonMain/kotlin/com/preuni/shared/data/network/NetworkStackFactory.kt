package com.preuni.shared.data.network

import com.preuni.shared.data.auth.AuthApiClient
import com.preuni.shared.data.auth.AuthRepositoryImpl
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.data.content.ContentRepositoryImpl
import com.preuni.shared.data.user.UserApiClient
import com.preuni.shared.data.user.UserRepositoryImpl
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.content.ContentRepository
import com.preuni.shared.domain.user.UserRepository
import io.ktor.client.HttpClient

/**
 * Holds all initialized network and repository dependencies for a single app session.
 * Produced by [NetworkStackFactory.create]; platform entrypoints consume this to build
 * [com.preuni.shared.presentation.RootComponent].
 */
data class NetworkStack(
    val httpClient: HttpClient,
    val tokenStore: TokenStore,
    val authRepository: AuthRepository,
    val userRepository: UserRepository,
    val contentRepository: ContentRepository,
)

/**
 * Single place that creates the authenticated HTTP client and wires all repositories.
 *
 * Previously each platform entrypoint (Android/iOS/Web) duplicated this bootstrap —
 * iOS and Web were missing bearer-token injection and token-refresh wiring, causing
 * Home and Profile to fail with 401 on non-Android targets.
 */
object NetworkStackFactory {

    /**
     * Creates a fully-wired [NetworkStack] for production use.
     *
     * @param baseUrl         Root URL for all API calls (e.g. `https://api.preuni.com.br`).
     * @param secureStorage   Platform-specific secure storage for session tokens.
     * @param onSessionExpired Called when the refresh token is invalid/expired so the
     *                        app can navigate back to the login screen.
     */
    fun create(
        baseUrl: String,
        secureStorage: SecureStorage,
        onSessionExpired: () -> Unit = {},
    ): NetworkStack {
        val tokenStore = TokenStore(secureStorage)
        val httpClient = buildHttpClient(baseUrl, getAccessToken = { tokenStore.accessToken() })
        val authApiClient = AuthApiClient(httpClient)
        val authRepository = AuthRepositoryImpl(authApiClient, tokenStore)

        httpClient.installTokenRefreshInterceptor(
            tokenStore = tokenStore,
            authApiClient = authApiClient,
            onSessionExpired = onSessionExpired,
        )
        httpClient.attachTelemetry()

        return NetworkStack(
            httpClient = httpClient,
            tokenStore = tokenStore,
            authRepository = authRepository,
            userRepository = UserRepositoryImpl(UserApiClient(httpClient)),
            contentRepository = ContentRepositoryImpl(httpClient),
        )
    }
}
