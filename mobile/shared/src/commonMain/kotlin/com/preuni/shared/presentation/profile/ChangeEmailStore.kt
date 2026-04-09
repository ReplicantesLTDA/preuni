package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

interface ChangeEmailStore : Store<ChangeEmailStore.Intent, ChangeEmailStore.State, ChangeEmailStore.Label> {

    enum class Phase { REQUEST, CONFIRM }

    data class State(
        val newEmail: String = "",
        val phase: Phase = Phase.REQUEST,
        val isLoading: Boolean = false,
        val error: String? = null,
        val resendCooldown: Int = 0,
    )

    sealed interface Intent {
        data class RequestCode(val newEmail: String) : Intent
        data class ConfirmCode(val otp: String) : Intent
        data object ResendCode : Intent
        data object ClearError : Intent
    }

    sealed interface Label {
        data class EmailChanged(val newEmail: String) : Label
    }
}

class ChangeEmailStoreFactory(
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
) {

    fun create(): ChangeEmailStore =
        object : ChangeEmailStore, Store<ChangeEmailStore.Intent, ChangeEmailStore.State, ChangeEmailStore.Label>
        by storeFactory.create(
            name = "ChangeEmailStore",
            initialState = ChangeEmailStore.State(),
            executorFactory = { Executor(authRepository) },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data object Loading : Msg
        data object DoneLoading : Msg
        data class OtpSent(val email: String) : Msg
        data class ErrorReceived(val message: String) : Msg
        data object ErrorCleared : Msg
        data class CooldownTick(val remaining: Int) : Msg
    }

    private inner class Executor(private val repo: AuthRepository) :
        CoroutineExecutor<ChangeEmailStore.Intent, Nothing, ChangeEmailStore.State, Msg, ChangeEmailStore.Label>() {

        override fun executeIntent(intent: ChangeEmailStore.Intent) {
            when (intent) {
                is ChangeEmailStore.Intent.RequestCode -> requestCode(intent.newEmail)
                is ChangeEmailStore.Intent.ConfirmCode -> confirmCode(intent.otp)
                ChangeEmailStore.Intent.ResendCode -> {
                    val email = state().newEmail
                    if (email.isNotBlank()) requestCode(email)
                }
                ChangeEmailStore.Intent.ClearError -> dispatch(Msg.ErrorCleared)
            }
        }

        private fun requestCode(email: String) {
            dispatch(Msg.Loading)
            scope.launch {
                repo.changeEmailRequest(email).fold(
                    onSuccess = {
                        dispatch(Msg.OtpSent(email))
                        startCooldown()
                    },
                    onFailure = { e ->
                        val message = when (e) {
                            is AppError.Conflict -> e.message
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

        private fun confirmCode(otp: String) {
            dispatch(Msg.Loading)
            scope.launch {
                repo.changeEmailConfirm(state().newEmail, otp).fold(
                    onSuccess = { publish(ChangeEmailStore.Label.EmailChanged(state().newEmail)) },
                    onFailure = { e ->
                        val message = when (e) {
                            is AppError.Validation -> e.message
                            is AppError.NetworkError -> "Sem conexão com a internet."
                            else -> "Código inválido ou expirado."
                        }
                        dispatch(Msg.ErrorReceived(message))
                    },
                )
                dispatch(Msg.DoneLoading)
            }
        }

        private fun startCooldown() {
            scope.launch {
                var remaining = 60
                while (remaining > 0) {
                    delay(1_000)
                    remaining--
                    dispatch(Msg.CooldownTick(remaining))
                }
            }
        }
    }

    private object ReducerImpl : Reducer<ChangeEmailStore.State, Msg> {
        override fun ChangeEmailStore.State.reduce(msg: Msg): ChangeEmailStore.State = when (msg) {
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            is Msg.OtpSent -> copy(newEmail = msg.email, phase = ChangeEmailStore.Phase.CONFIRM, resendCooldown = 60)
            is Msg.ErrorReceived -> copy(error = msg.message, isLoading = false)
            Msg.ErrorCleared -> copy(error = null)
            is Msg.CooldownTick -> copy(resendCooldown = msg.remaining)
        }
    }
}
