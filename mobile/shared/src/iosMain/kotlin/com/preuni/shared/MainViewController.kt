package com.preuni.shared

import androidx.compose.ui.window.ComposeUIViewController
import com.arkivanov.decompose.DefaultComponentContext
import com.arkivanov.essenty.lifecycle.LifecycleRegistry
import com.arkivanov.essenty.lifecycle.resume
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.data.auth.AuthApiClient
import com.preuni.shared.data.auth.AuthRepositoryImpl
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.data.content.ContentRepositoryImpl
import com.preuni.shared.data.network.buildHttpClient
import com.preuni.shared.data.user.UserApiClient
import com.preuni.shared.data.user.UserRepositoryImpl
import com.preuni.shared.presentation.RootComponent
import platform.UIKit.UIViewController

private const val IOS_BASE_URL = "http://localhost:8080"

fun MainViewController(): UIViewController {
    val httpClient = buildHttpClient(IOS_BASE_URL)

    val secureStorage = SecureStorage()
    val tokenStore = TokenStore(secureStorage)
    val authApiClient = AuthApiClient(httpClient)
    val authRepository = AuthRepositoryImpl(authApiClient, tokenStore)

    val userApiClient = UserApiClient(httpClient)
    val userRepository = UserRepositoryImpl(userApiClient)
    val contentRepository = ContentRepositoryImpl(httpClient)

    val lifecycle = LifecycleRegistry()
    lifecycle.resume()
    val componentContext = DefaultComponentContext(lifecycle = lifecycle)
    val storeFactory = DefaultStoreFactory()

    val root = RootComponent(
        componentContext = componentContext,
        storeFactory = storeFactory,
        tokenStore = tokenStore,
        authRepository = authRepository,
        userRepository = userRepository,
        contentRepository = contentRepository,
    )

    return ComposeUIViewController {
        PreuniApp(root)
    }
}
