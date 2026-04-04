package com.preuni.shared.presentation.auth

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface OtpLoginStore : Store<OtpLoginStore.Intent, OtpLoginStore.State, OtpLoginStore.Label> {

    data class State(
        val email: String = "",
        val otp: String = "",
        val codeSent: Boolean = false,
        val isLoading: Boolean = false,
        val error: String? = null,
    )

    sealed interface Intent {
        data class UpdateEmail(val value: String) : Intent
        data class UpdateOtp(val value: String) : Intent
        data object RequestCode : Intent
        data object VerifyCode : Intent
    }

    sealed interface Label {
        data class LoggedIn(val session: AuthSession) : Label
    }
}

class OtpLoginStoreFactory(
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
    private val requestOtp: suspend (email: String) -> Result<Unit>,
    private val verifyOtp: suspend (email: String, otp: String) -> Result<AuthSession>,
) {
    fun create(): OtpLoginStore =
        object : OtpLoginStore, Store<OtpLoginStore.Intent, OtpLoginStore.State, OtpLoginStore.Label>
        by storeFactory.create(
            name = "OtpLoginStore",
            initialState = OtpLoginStore.State(),
            executorFactory = { Executor() },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data class EmailChanged(val v: String) : Msg
        data class OtpChanged(val v: String) : Msg
        data object Loading : Msg
        data object DoneLoading : Msg
        data object CodeSent : Msg
        data class ErrorReceived(val msg: String) : Msg
    }

    private inner class Executor :
        CoroutineExecutor<OtpLoginStore.Intent, Nothing, OtpLoginStore.State, Msg, OtpLoginStore.Label>() {

        override fun executeIntent(intent: OtpLoginStore.Intent) {
            when (intent) {
                is OtpLoginStore.Intent.UpdateEmail -> dispatch(Msg.EmailChanged(intent.value))
                is OtpLoginStore.Intent.UpdateOtp -> dispatch(Msg.OtpChanged(intent.value))
                OtpLoginStore.Intent.RequestCode -> requestCode()
                OtpLoginStore.Intent.VerifyCode -> verifyCode()
            }
        }

        private fun requestCode() {
            dispatch(Msg.Loading)
            scope.launch {
                requestOtp(state().email).fold(
                    onSuccess = { dispatch(Msg.CodeSent) },
                    onFailure = { dispatch(Msg.ErrorReceived(it.message ?: "Failed to send code")) },
                )
                dispatch(Msg.DoneLoading)
            }
        }

        private fun verifyCode() {
            val s = state()
            dispatch(Msg.Loading)
            scope.launch {
                verifyOtp(s.email, s.otp).fold(
                    onSuccess = { session -> publish(OtpLoginStore.Label.LoggedIn(session)) },
                    onFailure = { e ->
                        dispatch(Msg.ErrorReceived(
                            when (e) {
                                is AppError.Validation -> e.message
                                else -> "Invalid or expired code"
                            }
                        ))
                    },
                )
                dispatch(Msg.DoneLoading)
            }
        }
    }

    private object ReducerImpl : Reducer<OtpLoginStore.State, Msg> {
        override fun OtpLoginStore.State.reduce(msg: Msg): OtpLoginStore.State = when (msg) {
            is Msg.EmailChanged -> copy(email = msg.v, error = null)
            is Msg.OtpChanged -> copy(otp = msg.v, error = null)
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            Msg.CodeSent -> copy(codeSent = true)
            is Msg.ErrorReceived -> copy(error = msg.msg, isLoading = false)
        }
    }
}
