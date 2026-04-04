package com.preuni.shared.data.network

import com.preuni.shared.data.auth.AuthApiClient
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.domain.error.AppError
import io.ktor.client.plugins.HttpSend
import io.ktor.client.plugins.plugin
import io.ktor.client.statement.HttpResponse
import io.ktor.http.HttpStatusCode
import io.ktor.client.HttpClient

/**
 * Installs a token-refresh interceptor on the given [HttpClient].
 *
 * On a 401 response:
 *  1. Calls POST /v1/auth/refresh with the stored refresh token.
 *  2. If successful, updates [TokenStore] and retries the original request once.
 *  3. If the refresh fails, clears [TokenStore] and emits [AppError.Unauthorized].
 */
fun HttpClient.installTokenRefreshInterceptor(
    tokenStore: TokenStore,
    authApiClient: AuthApiClient,
    onSessionExpired: () -> Unit,
) {
    plugin(HttpSend).intercept { request ->
        val response = execute(request)

        if (response.response.status != HttpStatusCode.Unauthorized) {
            return@intercept response
        }

        // Attempt silent refresh
        val refreshToken = tokenStore.refreshToken()
        if (refreshToken == null) {
            tokenStore.clear()
            onSessionExpired()
            return@intercept response
        }

        val refreshResult = authApiClient.refresh(refreshToken)
        if (refreshResult.isFailure) {
            tokenStore.clear()
            onSessionExpired()
            return@intercept response
        }

        val (newAccess, newRefresh) = refreshResult.getOrThrow()
        tokenStore.updateTokens(newAccess, newRefresh)

        // Retry original request with new access token
        request.headers.remove("Authorization")
        request.headers.append("Authorization", "Bearer $newAccess")
        execute(request)
    }
}
