package com.preuni.shared.presentation.auth

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.auth.AuthValidator
import com.preuni.shared.domain.auth.ValidationResult
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface LoginStore : Store<LoginStore.Intent, LoginStore.State, LoginStore.Label> {

    data class State(
        val emailOrUsername: String = "",
        val password: String = "",
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data class UpdateEmailOrUsername(val value: String) : Intent
        data class UpdatePassword(val value: String) : Intent
        data object Submit : Intent
        data object ClearError : Intent
    }

    sealed interface Label {
        data class LoggedIn(val session: AuthSession) : Label
    }
}

class LoginStoreFactory(
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
) {

    fun create(): LoginStore =
        object : LoginStore, Store<LoginStore.Intent, LoginStore.State, LoginStore.Label>
        by storeFactory.create(
            name = "LoginStore",
            initialState = LoginStore.State(),
            executorFactory = { Executor(authRepository) },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data class EmailOrUsernameChanged(val value: String) : Msg
        data class PasswordChanged(val value: String) : Msg
        data object LoadingStarted : Msg
        data object LoadingFinished : Msg
        data class ErrorReceived(val error: AppError) : Msg
        data object ErrorCleared : Msg
    }

    private inner class Executor(private val repo: AuthRepository) :
        CoroutineExecutor<LoginStore.Intent, Nothing, LoginStore.State, Msg, LoginStore.Label>() {

        override fun executeIntent(intent: LoginStore.Intent) {
            when (intent) {
                is LoginStore.Intent.UpdateEmailOrUsername ->
                    dispatch(Msg.EmailOrUsernameChanged(intent.value))

                is LoginStore.Intent.UpdatePassword ->
                    dispatch(Msg.PasswordChanged(intent.value))

                is LoginStore.Intent.ClearError ->
                    dispatch(Msg.ErrorCleared)

                LoginStore.Intent.Submit -> submit()
            }
        }

        private fun submit() {
            val state = state()

            // Client-side guard: do not hit network with empty fields
            if (state.emailOrUsername.isBlank()) {
                dispatch(Msg.ErrorReceived(AppError.Validation("emailOrUsername", "Email or username is required")))
                return
            }
            val pwResult = AuthValidator.validatePassword(state.password)
            if (pwResult is ValidationResult.Invalid) {
                dispatch(Msg.ErrorReceived(AppError.Validation("password", pwResult.message)))
                return
            }

            dispatch(Msg.LoadingStarted)
            scope.launch {
                val result = repo.login(state.emailOrUsername, state.password)
                dispatch(Msg.LoadingFinished)
                result.fold(
                    onSuccess = { session -> publish(LoginStore.Label.LoggedIn(session)) },
                    onFailure = { e ->
                        dispatch(Msg.ErrorReceived(e as? AppError ?: AppError.Unknown()))
                    },
                )
            }
        }
    }

    private object ReducerImpl : Reducer<LoginStore.State, Msg> {
        override fun LoginStore.State.reduce(msg: Msg): LoginStore.State =
            when (msg) {
                is Msg.EmailOrUsernameChanged -> copy(emailOrUsername = msg.value, error = null)
                is Msg.PasswordChanged -> copy(password = msg.value, error = null)
                Msg.LoadingStarted -> copy(isLoading = true, error = null)
                Msg.LoadingFinished -> copy(isLoading = false)
                is Msg.ErrorReceived -> copy(error = msg.error, isLoading = false)
                Msg.ErrorCleared -> copy(error = null)
            }
    }
}
