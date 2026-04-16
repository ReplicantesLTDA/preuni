package com.preuni.web

import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.window.CanvasBasedWindow
import com.arkivanov.decompose.DefaultComponentContext
import com.arkivanov.essenty.lifecycle.LifecycleRegistry
import com.arkivanov.essenty.lifecycle.resume
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import kotlinx.browser.window
import com.preuni.shared.PreuniApp
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.network.NetworkStackFactory
import com.preuni.shared.presentation.RootComponent

private fun browserOrigin(): String = window.location.origin

@OptIn(ExperimentalComposeUiApi::class)
fun main() {
    val stack = NetworkStackFactory.create(
        baseUrl = browserOrigin(),
        secureStorage = SecureStorage(),
    )

    val lifecycle = LifecycleRegistry()
    lifecycle.resume()
    val componentContext = DefaultComponentContext(lifecycle = lifecycle)
    val storeFactory = DefaultStoreFactory()

    val root = RootComponent(
        componentContext = componentContext,
        storeFactory = storeFactory,
        tokenStore = stack.tokenStore,
        authRepository = stack.authRepository,
        userRepository = stack.userRepository,
        contentRepository = stack.contentRepository,
    )

    CanvasBasedWindow("Preuni") {
        PreuniApp(root)
    }
}
