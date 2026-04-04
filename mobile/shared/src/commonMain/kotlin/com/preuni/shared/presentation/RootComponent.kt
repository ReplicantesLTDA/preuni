package com.preuni.shared.presentation

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.decompose.router.stack.ChildStack
import com.arkivanov.decompose.router.stack.StackNavigation
import com.arkivanov.decompose.router.stack.childStack
import com.arkivanov.decompose.router.stack.replaceAll
import com.arkivanov.decompose.value.Value
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.user.UserRepository
import com.preuni.shared.presentation.auth.AuthComponent
import com.preuni.shared.presentation.main.MainComponent
import com.preuni.shared.presentation.onboarding.OnboardingComponent
import kotlinx.serialization.Serializable

class RootComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val tokenStore: TokenStore,
    private val authRepository: AuthRepository,
    private val userRepository: UserRepository,
) : ComponentContext by componentContext {

    private val navigation = StackNavigation<Config>()

    private val initialConfig: Config
        get() {
            val session = tokenStore.load() ?: return Config.Auth
            // If logged in: check onboarding flag (stored after first onboarding complete)
            // For now we route to Onboarding — MainComponent sets the flag
            return if (onboardingCompleted()) Config.Main else Config.Onboarding
        }

    val childStack: Value<ChildStack<*, Child>> =
        childStack(
            source = navigation,
            serializer = Config.serializer(),
            initialConfiguration = initialConfig,
            handleBackButton = false,
            childFactory = ::createChild,
        )

    private fun createChild(config: Config, context: ComponentContext): Child =
        when (config) {
            Config.Auth -> Child.Auth(
                AuthComponent(context, storeFactory, authRepository) {
                    navigation.replaceAll(Config.Onboarding)
                }
            )
            Config.Onboarding -> Child.Onboarding(
                OnboardingComponent(context, storeFactory) {
                    navigation.replaceAll(Config.Main)
                }
            )
            Config.Main -> Child.Main(
                MainComponent(context, storeFactory, authRepository, tokenStore, userRepository) {
                    // On logout
                    navigation.replaceAll(Config.Auth)
                }
            )
        }

    private fun onboardingCompleted(): Boolean =
        tokenStore.load() != null // Will be refined when OnboardingStore persists its flag

    @Serializable
    sealed interface Config {
        @Serializable data object Auth : Config
        @Serializable data object Onboarding : Config
        @Serializable data object Main : Config
    }

    sealed interface Child {
        data class Auth(val component: AuthComponent) : Child
        data class Onboarding(val component: OnboardingComponent) : Child
        data class Main(val component: MainComponent) : Child
    }
}
