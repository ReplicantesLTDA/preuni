package com.preuni.shared.presentation.profile

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.decompose.DelicateDecomposeApi
import com.arkivanov.decompose.router.stack.ChildStack
import com.arkivanov.decompose.router.stack.StackNavigation
import com.arkivanov.decompose.router.stack.childStack
import com.arkivanov.decompose.router.stack.pop
import com.arkivanov.decompose.router.stack.push
import com.arkivanov.decompose.router.stack.replaceAll
import com.arkivanov.decompose.value.Value
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.content.ContentRepository
import com.preuni.shared.domain.user.UserRepository
import kotlinx.serialization.Serializable

class ProfileComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val userRepository: UserRepository,
    private val contentRepository: ContentRepository,
    private val authRepository: AuthRepository,
    private val onLogout: () -> Unit,
) : ComponentContext by componentContext {

    private val profileStore = ProfileStoreFactory(storeFactory, userRepository, authRepository).create()
    private val changePasswordStore = ChangePasswordStoreFactory(storeFactory, authRepository).create()
    private val changeEmailStore = ChangeEmailStoreFactory(storeFactory, authRepository).create()

    private val navigation = StackNavigation<Config>()

    val childStack: Value<ChildStack<*, Child>> =
        childStack(
            source = navigation,
            serializer = Config.serializer(),
            initialConfiguration = Config.Profile,
            handleBackButton = true,
            childFactory = ::createChild,
        )

    @OptIn(DelicateDecomposeApi::class)
    private fun createChild(config: Config, context: ComponentContext): Child =
        when (config) {
            Config.Profile -> Child.Profile(profileStore)
            Config.EditUsername -> Child.EditUsername(profileStore)
            Config.EditPassword -> Child.EditPassword(changePasswordStore)
            Config.ChangeEmail -> Child.ChangeEmail(changeEmailStore)
            Config.DeleteAccount -> Child.DeleteAccount(profileStore)
            is Config.ChangeTrack -> Child.ChangeTrack(
                store = ChangeTrackStoreFactory(storeFactory) { ids ->
                    userRepository.updateTracks(ids)
                }.create().also { store ->
                    store.accept(ChangeTrackStore.Intent.Load(config.initialIds))
                },
                contentRepository = contentRepository,
            )
        }

    fun navigateToEditUsername() = navigation.push(Config.EditUsername)
    fun navigateToEditPassword() = navigation.push(Config.EditPassword)
    fun navigateToChangeEmail() = navigation.push(Config.ChangeEmail)
    fun navigateToDeleteAccount() = navigation.push(Config.DeleteAccount)
    fun navigateToChangeTrack(initialIds: Set<String>) = navigation.push(Config.ChangeTrack(initialIds))
    fun navigateBack() = navigation.pop()
    fun resetToRoot() = navigation.replaceAll(Config.Profile)
    fun logout() = onLogout()

    @Serializable
    sealed interface Config {
        @Serializable data object Profile : Config
        @Serializable data object EditUsername : Config
        @Serializable data object EditPassword : Config
        @Serializable data object ChangeEmail : Config
        @Serializable data object DeleteAccount : Config
        @Serializable data class ChangeTrack(val initialIds: Set<String>) : Config
    }

    sealed interface Child {
        data class Profile(val store: ProfileStore) : Child
        data class EditUsername(val store: ProfileStore) : Child
        data class EditPassword(val store: ChangePasswordStore) : Child
        data class ChangeEmail(val store: ChangeEmailStore) : Child
        data class DeleteAccount(val store: ProfileStore) : Child
        data class ChangeTrack(val store: ChangeTrackStore, val contentRepository: ContentRepository) : Child
    }
}
