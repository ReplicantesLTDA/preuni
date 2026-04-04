package com.preuni.web

import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.window.CanvasBasedWindow
import com.arkivanov.decompose.DefaultComponentContext
import com.arkivanov.essenty.lifecycle.LifecycleRegistry
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.PreuniApp
import com.preuni.shared.data.auth.AuthApiClient
import com.preuni.shared.data.auth.AuthRepositoryImpl
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.data.network.buildHttpClient
import com.preuni.shared.data.user.UserApiClient
import com.preuni.shared.data.user.UserRepositoryImpl
import com.preuni.shared.presentation.RootComponent

private fun browserOrigin(): String = js("window.location.origin")

@OptIn(ExperimentalComposeUiApi::class)
fun main() {
    val baseUrl = browserOrigin()
    val httpClient = buildHttpClient(baseUrl)

    val secureStorage = SecureStorage()
    val tokenStore = TokenStore(secureStorage)
    val authApiClient = AuthApiClient(httpClient)
    val authRepository = AuthRepositoryImpl(authApiClient, tokenStore)

    val userApiClient = UserApiClient(httpClient)
    val userRepository = UserRepositoryImpl(userApiClient)

    val lifecycle = LifecycleRegistry()
    val componentContext = DefaultComponentContext(lifecycle = lifecycle)
    val storeFactory = DefaultStoreFactory()

    val root = RootComponent(
        componentContext = componentContext,
        storeFactory = storeFactory,
        tokenStore = tokenStore,
        authRepository = authRepository,
        userRepository = userRepository,
    )

    CanvasBasedWindow("Preuni") {
        PreuniApp(root)
    }
}
