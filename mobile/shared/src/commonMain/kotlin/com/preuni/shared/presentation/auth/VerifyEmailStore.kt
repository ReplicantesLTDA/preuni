package com.preuni.shared.presentation.auth

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import kotlinx.coroutines.launch

interface VerifyEmailStore : Store<VerifyEmailStore.Intent, VerifyEmailStore.State, Nothing> {

    data class State(
        val code: String = "",
        val isLoading: Boolean = false,
        val error: String? = null,
        val verified: Boolean = false,
    )

    sealed interface Intent {
        data class UpdateCode(val value: String) : Intent
        data object Submit : Intent
        data object Resend : Intent
    }
}

class VerifyEmailStoreFactory(
    private val storeFactory: StoreFactory,
    private val email: String,
    private val onSubmit: suspend (email: String, code: String) -> Result<Unit>,
    private val onResend: suspend (email: String) -> Result<Unit>,
) {
    fun create(): VerifyEmailStore =
        object : VerifyEmailStore, Store<VerifyEmailStore.Intent, VerifyEmailStore.State, Nothing>
        by storeFactory.create(
            name = "VerifyEmailStore",
            initialState = VerifyEmailStore.State(),
            executorFactory = { Executor() },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data class CodeChanged(val v: String) : Msg
        data object Loading : Msg
        data object DoneLoading : Msg
        data class ErrorReceived(val msg: String) : Msg
        data object Verified : Msg
    }

    private inner class Executor :
        CoroutineExecutor<VerifyEmailStore.Intent, Nothing, VerifyEmailStore.State, Msg, Nothing>() {

        override fun executeIntent(intent: VerifyEmailStore.Intent) {
            when (intent) {
                is VerifyEmailStore.Intent.UpdateCode -> dispatch(Msg.CodeChanged(intent.value))
                VerifyEmailStore.Intent.Submit -> submit()
                VerifyEmailStore.Intent.Resend -> resend()
            }
        }

        private fun submit() {
            val code = state().code
            dispatch(Msg.Loading)
            scope.launch {
                onSubmit(email, code).fold(
                    onSuccess = { dispatch(Msg.Verified) },
                    onFailure = { dispatch(Msg.ErrorReceived(it.message ?: "Invalid or expired code")) },
                )
                dispatch(Msg.DoneLoading)
            }
        }

        private fun resend() {
            scope.launch { onResend(email) }
        }
    }

    private object ReducerImpl : Reducer<VerifyEmailStore.State, Msg> {
        override fun VerifyEmailStore.State.reduce(msg: Msg): VerifyEmailStore.State = when (msg) {
            is Msg.CodeChanged -> copy(code = msg.v, error = null)
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            is Msg.ErrorReceived -> copy(error = msg.msg, isLoading = false)
            Msg.Verified -> copy(verified = true, isLoading = false)
        }
    }
}
