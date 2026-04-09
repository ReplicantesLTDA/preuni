package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.decompose.router.stack.ChildStack
import com.arkivanov.decompose.router.stack.StackNavigation
import com.arkivanov.decompose.router.stack.childStack
import com.arkivanov.decompose.router.stack.pop
import com.arkivanov.decompose.router.stack.push
import com.arkivanov.decompose.router.stack.replaceAll
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
                LoginStoreFactory(storeFactory, authRepository).create()
            )
            Config.Register -> Child.Register(
                RegisterComponent(context, storeFactory, authRepository) {
                    navigation.pop()
                }
            )
            is Config.VerifyEmail -> Child.VerifyEmail(
                VerifyEmailComponent(
                    componentContext = context,
                    storeFactory = storeFactory,
                    authRepository = authRepository,
                    email = config.email,
                    onVerified = { navigation.replaceAll(Config.Login) },
                )
            )
            Config.OtpLogin -> Child.OtpLogin(
                OtpLoginComponent(context, storeFactory, authRepository) {
                    onLoggedIn()
                }
            )
        }

    fun onLoginSuccess() = onLoggedIn()
    fun navigateToRegister() = navigation.push(Config.Register)
    fun navigateToVerifyEmail(email: String) = navigation.push(Config.VerifyEmail(email))
    fun navigateToOtpLogin() = navigation.push(Config.OtpLogin)
    fun navigateBack() = navigation.pop()

    @Serializable
    sealed interface Config {
        @Serializable data object Login : Config
        @Serializable data object Register : Config
        @Serializable data class VerifyEmail(val email: String) : Config
        @Serializable data object OtpLogin : Config
    }

    sealed interface Child {
        data class Login(val store: LoginStore) : Child
        data class Register(val component: RegisterComponent) : Child
        data class VerifyEmail(val component: VerifyEmailComponent) : Child
        data class OtpLogin(val component: OtpLoginComponent) : Child
    }
}
