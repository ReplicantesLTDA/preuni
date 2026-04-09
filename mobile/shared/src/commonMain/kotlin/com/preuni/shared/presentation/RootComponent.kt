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
import com.preuni.shared.domain.content.ContentRepository
import com.preuni.shared.domain.user.UserRepository
import com.preuni.shared.presentation.auth.AuthComponent
import com.preuni.shared.presentation.main.MainComponent
import com.preuni.shared.presentation.onboarding.OnboardingComponent
import com.preuni.shared.presentation.welcome.WelcomeComponent
import kotlinx.serialization.Serializable

class RootComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val tokenStore: TokenStore,
    private val authRepository: AuthRepository,
    private val userRepository: UserRepository,
    private val contentRepository: ContentRepository,
) : ComponentContext by componentContext {

    private val navigation = StackNavigation<Config>()

    private val initialConfig: Config
        get() = when {
            !tokenStore.welcomeSeen() -> Config.Welcome
            tokenStore.load() == null -> Config.Auth
            else -> Config.Main
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
            Config.Welcome -> Child.Welcome(
                WelcomeComponent(context, storeFactory) {
                    tokenStore.markWelcomeSeen()
                    navigation.replaceAll(Config.Auth)
                }
            )
            Config.Auth -> Child.Auth(
                AuthComponent(context, storeFactory, authRepository) {
                    navigation.replaceAll(Config.Onboarding)
                }
            )
            Config.Onboarding -> Child.Onboarding(
                OnboardingComponent(
                    componentContext = context,
                    storeFactory = storeFactory,
                    onCompleted = { navigation.replaceAll(Config.Main) },
                    setActiveTrackId = { trackId -> tokenStore.setActiveTrackId(trackId) },
                    completeOnboarding = { trackIds -> userRepository.updateTracks(trackIds) },
                )
            )
            Config.Main -> Child.Main(
                MainComponent(context, storeFactory, authRepository, tokenStore, userRepository, contentRepository) {
                    navigation.replaceAll(Config.Auth)
                }
            )
        }

    @Serializable
    sealed interface Config {
        @Serializable data object Welcome : Config
        @Serializable data object Auth : Config
        @Serializable data object Onboarding : Config
        @Serializable data object Main : Config
    }

    sealed interface Child {
        data class Welcome(val component: WelcomeComponent) : Child
        data class Auth(val component: AuthComponent) : Child
        data class Onboarding(val component: OnboardingComponent) : Child
        data class Main(val component: MainComponent) : Child
    }
}
