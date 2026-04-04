package com.preuni.shared.presentation.auth

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthValidator
import com.preuni.shared.domain.auth.ValidationResult
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface RegisterStore : Store<RegisterStore.Intent, RegisterStore.State, RegisterStore.Label> {

    data class State(
        val displayName: String = "",
        val email: String = "",
        val username: String = "",
        val password: String = "",
        val confirmPassword: String = "",
        val emailError: String? = null,
        val passwordError: String? = null,
        val confirmPasswordError: String? = null,
        val globalError: AppError? = null,
        val isLoading: Boolean = false,
    )

    sealed interface Intent {
        data class UpdateDisplayName(val value: String) : Intent
        data class UpdateEmail(val value: String) : Intent
        data class UpdateUsername(val value: String) : Intent
        data class UpdatePassword(val value: String) : Intent
        data class UpdateConfirmPassword(val value: String) : Intent
        data object Submit : Intent
    }

    sealed interface Label {
        data object Registered : Label
    }
}

class RegisterStoreFactory(
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
) {
    fun create(): RegisterStore =
        object : RegisterStore, Store<RegisterStore.Intent, RegisterStore.State, RegisterStore.Label>
        by storeFactory.create(
            name = "RegisterStore",
            initialState = RegisterStore.State(),
            executorFactory = { Executor(authRepository) },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data class DisplayNameChanged(val v: String) : Msg
        data class EmailChanged(val v: String) : Msg
        data class UsernameChanged(val v: String) : Msg
        data class PasswordChanged(val v: String) : Msg
        data class ConfirmPasswordChanged(val v: String) : Msg
        data class EmailError(val msg: String?) : Msg
        data class PasswordError(val msg: String?) : Msg
        data class ConfirmError(val msg: String?) : Msg
        data class GlobalError(val err: AppError?) : Msg
        data object Loading : Msg
        data object DoneLoading : Msg
    }

    private inner class Executor(private val repo: AuthRepository) :
        CoroutineExecutor<RegisterStore.Intent, Nothing, RegisterStore.State, Msg, RegisterStore.Label>() {

        override fun executeIntent(intent: RegisterStore.Intent) {
            when (intent) {
                is RegisterStore.Intent.UpdateDisplayName -> dispatch(Msg.DisplayNameChanged(intent.value))
                is RegisterStore.Intent.UpdateEmail -> dispatch(Msg.EmailChanged(intent.value))
                is RegisterStore.Intent.UpdateUsername -> dispatch(Msg.UsernameChanged(intent.value))
                is RegisterStore.Intent.UpdatePassword -> dispatch(Msg.PasswordChanged(intent.value))
                is RegisterStore.Intent.UpdateConfirmPassword -> dispatch(Msg.ConfirmPasswordChanged(intent.value))
                RegisterStore.Intent.Submit -> submit()
            }
        }

        private fun submit() {
            val s = state()
            var hasError = false

            val emailResult = AuthValidator.validateEmail(s.email)
            if (emailResult is ValidationResult.Invalid) {
                dispatch(Msg.EmailError(emailResult.message)); hasError = true
            }
            val pwResult = AuthValidator.validatePassword(s.password)
            if (pwResult is ValidationResult.Invalid) {
                dispatch(Msg.PasswordError(pwResult.message)); hasError = true
            }
            val matchResult = AuthValidator.validatePasswordsMatch(s.password, s.confirmPassword)
            if (matchResult is ValidationResult.Invalid) {
                dispatch(Msg.ConfirmError(matchResult.message)); hasError = true
            }
            if (hasError) return

            dispatch(Msg.Loading)
            scope.launch {
                // Registration API call will be wired once AuthApiClient.register() is added
                dispatch(Msg.DoneLoading)
                publish(RegisterStore.Label.Registered)
            }
        }
    }

    private object ReducerImpl : Reducer<RegisterStore.State, Msg> {
        override fun RegisterStore.State.reduce(msg: Msg): RegisterStore.State = when (msg) {
            is Msg.DisplayNameChanged -> copy(displayName = msg.v)
            is Msg.EmailChanged -> copy(email = msg.v, emailError = null)
            is Msg.UsernameChanged -> copy(username = msg.v)
            is Msg.PasswordChanged -> copy(password = msg.v, passwordError = null)
            is Msg.ConfirmPasswordChanged -> copy(confirmPassword = msg.v, confirmPasswordError = null)
            is Msg.EmailError -> copy(emailError = msg.msg)
            is Msg.PasswordError -> copy(passwordError = msg.msg)
            is Msg.ConfirmError -> copy(confirmPasswordError = msg.msg)
            is Msg.GlobalError -> copy(globalError = msg.err, isLoading = false)
            Msg.Loading -> copy(isLoading = true, globalError = null)
            Msg.DoneLoading -> copy(isLoading = false)
        }
    }
}
