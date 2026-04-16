package com.preuni.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import com.arkivanov.decompose.DefaultComponentContext
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.PreuniApp
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.network.NetworkStackFactory
import com.preuni.shared.presentation.RootComponent

class MainActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val baseUrl = "http://10.0.2.2:8080" // Android emulator → host loopback

        val stack = NetworkStackFactory.create(
            baseUrl = baseUrl,
            secureStorage = SecureStorage(),
        )

        val storeFactory = DefaultStoreFactory()
        val root = RootComponent(
            componentContext = DefaultComponentContext(lifecycle),
            storeFactory = storeFactory,
            tokenStore = stack.tokenStore,
            authRepository = stack.authRepository,
            userRepository = stack.userRepository,
            contentRepository = stack.contentRepository,
        )

        setContent {
            PreuniApp(root)
        }
    }
}
