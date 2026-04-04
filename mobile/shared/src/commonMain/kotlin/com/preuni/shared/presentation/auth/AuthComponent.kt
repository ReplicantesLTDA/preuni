package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.decompose.router.stack.ChildStack
import com.arkivanov.decompose.router.stack.StackNavigation
import com.arkivanov.decompose.router.stack.childStack
import com.arkivanov.decompose.router.stack.pop
import com.arkivanov.decompose.router.stack.push
import com.arkivanov.decompose.value.Value
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.auth.AuthRepository
import kotlinx.serialization.Serializable

class AuthComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
    private val onLoggedIn: () -> Unit,
) : ComponentContext by componentContext {

    private val navigation = StackNavigation<Config>()

    val childStack: Value<ChildStack<*, Child>> =
        childStack(
            source = navigation,
            serializer = Config.serializer(),
            initialConfiguration = Config.Login,
            handleBackButton = true,
            childFactory = ::createChild,
        )

    private fun createChild(config: Config, context: ComponentContext): Child =
        when (config) {
            Config.Login -> Child.Login(
                LoginStoreFactory(storeFactory, authRepository).create().also { store ->
                    // Observe the LoggedIn label to propagate navigation
                    // Note: label subscription is wired in the parent composable
                }
            )
            Config.Register -> Child.Register(
                RegisterComponent(context, storeFactory, authRepository) {
                    navigation.pop()
                }
            )
            Config.VerifyEmail -> Child.VerifyEmail(
                VerifyEmailComponent(context, storeFactory) {
                    // On verified, navigate back to login
                    navigation.pop()
                }
            )
            Config.OtpLogin -> Child.OtpLogin(
                OtpLoginComponent(context, storeFactory, authRepository) {
                    onLoggedIn()
                }
            )
        }

    fun onLoginSuccess() = onLoggedIn()
    fun navigateToRegister() = navigation.push(Config.Register)
    fun navigateToOtpLogin() = navigation.push(Config.OtpLogin)
    fun navigateToVerifyEmail() = navigation.push(Config.VerifyEmail)
    fun navigateBack() = navigation.pop()

    @Serializable
    sealed interface Config {
        @Serializable data object Login : Config
        @Serializable data object Register : Config
        @Serializable data object VerifyEmail : Config
        @Serializable data object OtpLogin : Config
    }

    sealed interface Child {
        data class Login(val store: LoginStore) : Child
        data class Register(val component: RegisterComponent) : Child
        data class VerifyEmail(val component: VerifyEmailComponent) : Child
        data class OtpLogin(val component: OtpLoginComponent) : Child
    }
}
