package com.preuni.shared.presentation.profile

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.decompose.router.stack.ChildStack
import com.arkivanov.decompose.router.stack.StackNavigation
import com.arkivanov.decompose.router.stack.childStack
import com.arkivanov.decompose.router.stack.pop
import com.arkivanov.decompose.router.stack.push
import com.arkivanov.decompose.value.Value
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.user.UserRepository
import kotlinx.serialization.Serializable

class ProfileComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val userRepository: UserRepository,
    private val onLogout: () -> Unit,
) : ComponentContext by componentContext {

    private val navigation = StackNavigation<Config>()

    val childStack: Value<ChildStack<*, Child>> =
        childStack(
            source = navigation,
            serializer = Config.serializer(),
            initialConfiguration = Config.Profile,
            handleBackButton = true,
            childFactory = ::createChild,
        )

    private fun createChild(config: Config, context: ComponentContext): Child =
        when (config) {
            Config.Profile -> Child.Profile(
                ProfileStoreFactory(storeFactory, userRepository).create()
            )
            Config.EditUsername -> Child.EditUsername
            Config.EditPassword -> Child.EditPassword
            Config.ChangeEmail -> Child.ChangeEmail
            Config.ConfirmNewEmail -> Child.ConfirmNewEmail
            Config.DeleteAccount -> Child.DeleteAccount
        }

    fun navigateToEditUsername() = navigation.push(Config.EditUsername)
    fun navigateToEditPassword() = navigation.push(Config.EditPassword)
    fun navigateToChangeEmail() = navigation.push(Config.ChangeEmail)
    fun navigateToConfirmNewEmail() = navigation.push(Config.ConfirmNewEmail)
    fun navigateToDeleteAccount() = navigation.push(Config.DeleteAccount)
    fun navigateBack() = navigation.pop()

    @Serializable
    sealed interface Config {
        @Serializable data object Profile : Config
        @Serializable data object EditUsername : Config
        @Serializable data object EditPassword : Config
        @Serializable data object ChangeEmail : Config
        @Serializable data object ConfirmNewEmail : Config
        @Serializable data object DeleteAccount : Config
    }

    sealed interface Child {
        data class Profile(val store: ProfileStore) : Child
        data object EditUsername : Child
        data object EditPassword : Child
        data object ChangeEmail : Child
        data object ConfirmNewEmail : Child
        data object DeleteAccount : Child
    }
}
