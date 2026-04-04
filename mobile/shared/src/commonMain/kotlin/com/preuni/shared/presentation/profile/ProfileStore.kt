package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.auth.AuthValidator
import com.preuni.shared.domain.auth.ValidationResult
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.user.Student
import com.preuni.shared.domain.user.UserRepository
import kotlinx.coroutines.launch

interface ProfileStore : Store<ProfileStore.Intent, ProfileStore.State, ProfileStore.Label> {

    data class State(
        val student: Student? = null,
        val isLoading: Boolean = false,
        val error: AppError? = null,
        val showDeleteConfirmation: Boolean = false,
    )

    sealed interface Intent {
        data object LoadProfile : Intent
        data class UpdateDisplayName(val value: String) : Intent
        data class UpdateUsername(val value: String) : Intent
        data object UploadAvatar : Intent
        data class ConfirmAvatarUpload(val objectKey: String) : Intent
        data object DeleteAccount : Intent
        data object ConfirmDeleteAccount : Intent
        data object DismissDeleteConfirmation : Intent
        data object ClearError : Intent
    }

    sealed interface Label {
        data object NavigateToEditUsername : Label
        data object NavigateToEditPassword : Label
        data object NavigateToChangeEmail : Label
        data object AccountDeleted : Label
        data class AvatarUploadReady(val uploadUrl: String, val objectKey: String) : Label
    }
}

class ProfileStoreFactory(
    private val storeFactory: StoreFactory,
    private val userRepository: UserRepository,
) {

    fun create(): ProfileStore =
        object : ProfileStore, Store<ProfileStore.Intent, ProfileStore.State, ProfileStore.Label>
        by storeFactory.create(
            name = "ProfileStore",
            initialState = ProfileStore.State(),
            executorFactory = { Executor(userRepository) },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data object Loading : Msg
        data object DoneLoading : Msg
        data class StudentLoaded(val student: Student) : Msg
        data class ErrorReceived(val error: AppError) : Msg
        data object ErrorCleared : Msg
        data object ShowDeleteConfirmation : Msg
        data object HideDeleteConfirmation : Msg
    }

    private inner class Executor(private val repo: UserRepository) :
        CoroutineExecutor<ProfileStore.Intent, Nothing, ProfileStore.State, Msg, ProfileStore.Label>() {

        override fun executeIntent(intent: ProfileStore.Intent) {
            when (intent) {
                ProfileStore.Intent.LoadProfile -> load()
                is ProfileStore.Intent.UpdateUsername -> {
                    val result = AuthValidator.validateUsername(intent.value)
                    if (result is ValidationResult.Invalid) {
                        dispatch(Msg.ErrorReceived(AppError.Validation("username", result.message)))
                    } else {
                        updateProfile(username = intent.value)
                    }
                }
                is ProfileStore.Intent.UpdateDisplayName -> updateProfile(displayName = intent.value)
                ProfileStore.Intent.UploadAvatar -> uploadAvatar()
                is ProfileStore.Intent.ConfirmAvatarUpload -> confirmAvatar(intent.objectKey)
                ProfileStore.Intent.DeleteAccount -> dispatch(Msg.ShowDeleteConfirmation)
                ProfileStore.Intent.DismissDeleteConfirmation -> dispatch(Msg.HideDeleteConfirmation)
                ProfileStore.Intent.ConfirmDeleteAccount -> deleteAccount()
                ProfileStore.Intent.ClearError -> dispatch(Msg.ErrorCleared)
            }
        }

        private fun load() {
            dispatch(Msg.Loading)
            scope.launch {
                repo.getMe().fold(
                    onSuccess = { dispatch(Msg.StudentLoaded(it)) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
                dispatch(Msg.DoneLoading)
            }
        }

        private fun updateProfile(displayName: String? = null, username: String? = null) {
            dispatch(Msg.Loading)
            scope.launch {
                repo.updateProfile(displayName, username).fold(
                    onSuccess = { dispatch(Msg.StudentLoaded(it)) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
                dispatch(Msg.DoneLoading)
            }
        }

        private fun uploadAvatar() {
            scope.launch {
                repo.getAvatarUploadUrl().fold(
                    onSuccess = { publish(ProfileStore.Label.AvatarUploadReady(it.uploadUrl, it.objectKey)) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
            }
        }

        private fun confirmAvatar(objectKey: String) {
            scope.launch {
                repo.confirmAvatarUpload(objectKey).fold(
                    onSuccess = { dispatch(Msg.StudentLoaded(it)) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
            }
        }

        private fun deleteAccount() {
            dispatch(Msg.Loading)
            scope.launch {
                repo.anonymize().fold(
                    onSuccess = { publish(ProfileStore.Label.AccountDeleted) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
                dispatch(Msg.DoneLoading)
            }
        }
    }

    private object ReducerImpl : Reducer<ProfileStore.State, Msg> {
        override fun ProfileStore.State.reduce(msg: Msg): ProfileStore.State = when (msg) {
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            is Msg.StudentLoaded -> copy(student = msg.student, isLoading = false)
            is Msg.ErrorReceived -> copy(error = msg.error, isLoading = false)
            Msg.ErrorCleared -> copy(error = null)
            Msg.ShowDeleteConfirmation -> copy(showDeleteConfirmation = true)
            Msg.HideDeleteConfirmation -> copy(showDeleteConfirmation = false)
        }
    }
}
