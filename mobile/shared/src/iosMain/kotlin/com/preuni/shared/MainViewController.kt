package com.preuni.shared

import androidx.compose.ui.window.ComposeUIViewController
import com.arkivanov.decompose.DefaultComponentContext
import com.arkivanov.essenty.lifecycle.LifecycleRegistry
import com.arkivanov.essenty.lifecycle.resume
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.network.NetworkStackFactory
import com.preuni.shared.presentation.RootComponent
import platform.UIKit.UIViewController

private const val IOS_BASE_URL = "http://localhost:8080"

fun MainViewController(): UIViewController {
    val stack = NetworkStackFactory.create(
        baseUrl = IOS_BASE_URL,
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

    return ComposeUIViewController {
        PreuniApp(root)
    }
}
