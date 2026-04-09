package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface ChangePasswordStore : Store<ChangePasswordStore.Intent, ChangePasswordStore.State, ChangePasswordStore.Label> {

    data class State(
        val isLoading: Boolean = false,
        val error: String? = null,
    )

    sealed interface Intent {
        data class Submit(val currentPassword: String, val newPassword: String) : Intent
        data object ClearError : Intent
    }

    sealed interface Label {
        data object Saved : Label
    }
}

class ChangePasswordStoreFactory(
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
) {

    fun create(): ChangePasswordStore =
        object : ChangePasswordStore, Store<ChangePasswordStore.Intent, ChangePasswordStore.State, ChangePasswordStore.Label>
        by storeFactory.create(
            name = "ChangePasswordStore",
            initialState = ChangePasswordStore.State(),
            executorFactory = { Executor(authRepository) },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data object Loading : Msg
        data class ErrorReceived(val message: String) : Msg
        data object ErrorCleared : Msg
        data object DoneLoading : Msg
    }

    private inner class Executor(private val repo: AuthRepository) :
        CoroutineExecutor<ChangePasswordStore.Intent, Nothing, ChangePasswordStore.State, Msg, ChangePasswordStore.Label>() {

        override fun executeIntent(intent: ChangePasswordStore.Intent) {
            when (intent) {
                is ChangePasswordStore.Intent.Submit -> changePassword(intent.currentPassword, intent.newPassword)
                ChangePasswordStore.Intent.ClearError -> dispatch(Msg.ErrorCleared)
            }
        }

        private fun changePassword(current: String, new: String) {
            dispatch(Msg.Loading)
            scope.launch {
                repo.changePassword(current, new).fold(
                    onSuccess = { publish(ChangePasswordStore.Label.Saved) },
                    onFailure = { e ->
                        val message = when (e) {
                            is AppError.Validation -> e.message
                            is AppError.NetworkError -> "Sem conexão com a internet."
                            else -> "Algo deu errado. Tente novamente."
                        }
                        dispatch(Msg.ErrorReceived(message))
                    },
                )
                dispatch(Msg.DoneLoading)
            }
        }
    }

    private object ReducerImpl : Reducer<ChangePasswordStore.State, Msg> {
        override fun ChangePasswordStore.State.reduce(msg: Msg): ChangePasswordStore.State = when (msg) {
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            is Msg.ErrorReceived -> copy(error = msg.message, isLoading = false)
            Msg.ErrorCleared -> copy(error = null)
        }
    }
}
