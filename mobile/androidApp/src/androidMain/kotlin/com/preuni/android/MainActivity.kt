package com.preuni.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import com.arkivanov.decompose.DefaultComponentContext
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.PreuniApp
import com.preuni.shared.data.auth.AuthApiClient
import com.preuni.shared.data.auth.AuthRepositoryImpl
import com.preuni.shared.data.auth.SecureStorage
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.data.content.ContentRepositoryImpl
import com.preuni.shared.data.network.buildHttpClient
import com.preuni.shared.data.user.UserApiClient
import com.preuni.shared.data.user.UserRepositoryImpl
import com.preuni.shared.presentation.RootComponent

class MainActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val baseUrl = "http://10.0.2.2:8080" // Android emulator → host loopback
        val httpClient = buildHttpClient(baseUrl)

        val secureStorage = SecureStorage()
        val tokenStore = TokenStore(secureStorage)
        val authApiClient = AuthApiClient(httpClient)
        val authRepository = AuthRepositoryImpl(authApiClient, tokenStore)

        val userApiClient = UserApiClient(httpClient)
        val userRepository = UserRepositoryImpl(userApiClient)
        val contentRepository = ContentRepositoryImpl(httpClient)

        val storeFactory = DefaultStoreFactory()
        val root = RootComponent(
            componentContext = DefaultComponentContext(lifecycle),
            storeFactory = storeFactory,
            tokenStore = tokenStore,
            authRepository = authRepository,
            userRepository = userRepository,
            contentRepository = contentRepository,
        )

        setContent {
            PreuniApp(root)
        }
    }
}
