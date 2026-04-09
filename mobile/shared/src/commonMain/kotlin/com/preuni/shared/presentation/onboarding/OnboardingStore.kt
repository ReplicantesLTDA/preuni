package com.preuni.shared.presentation.onboarding

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch


interface OnboardingStore : Store<OnboardingStore.Intent, OnboardingStore.State, OnboardingStore.Label> {

    data class State(
        val selectedTrackId: String? = null,
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data class SelectTrack(val trackId: String) : Intent
        data object Complete : Intent
    }

    sealed interface Label {
        data object Completed : Label
        data class ValidationError(val message: String) : Label
    }
}

class OnboardingStoreFactory(
    private val storeFactory: StoreFactory,
    private val setActiveTrackId: (String) -> Unit,
    private val completeOnboarding: suspend (trackIds: List<String>) -> Result<Unit>,
) {

    fun create(): OnboardingStore =
        object : OnboardingStore, Store<OnboardingStore.Intent, OnboardingStore.State, OnboardingStore.Label>
        by storeFactory.create(
            name = "OnboardingStore",
            initialState = OnboardingStore.State(),
            executorFactory = { Executor() },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data class TrackSelected(val trackId: String) : Msg
    }

    private inner class Executor :
        CoroutineExecutor<OnboardingStore.Intent, Nothing, OnboardingStore.State, Msg, OnboardingStore.Label>() {

        override fun executeIntent(intent: OnboardingStore.Intent) {
            when (intent) {
                is OnboardingStore.Intent.SelectTrack -> dispatch(Msg.TrackSelected(intent.trackId))
                OnboardingStore.Intent.Complete -> complete()
            }
        }

        private fun complete() {
            val trackId = state().selectedTrackId
            if (trackId == null) {
                publish(OnboardingStore.Label.ValidationError("Selecione uma matéria para continuar"))
                return
            }
            // Save locally and navigate immediately — resilient to network errors
            setActiveTrackId(trackId)
            publish(OnboardingStore.Label.Completed)
            // Fire-and-forget backend sync
            scope.launch {
                completeOnboarding(listOf(trackId))
            }
        }
    }

    private object ReducerImpl : Reducer<OnboardingStore.State, Msg> {
        override fun OnboardingStore.State.reduce(msg: Msg): OnboardingStore.State = when (msg) {
            is Msg.TrackSelected -> copy(selectedTrackId = msg.trackId)
        }
    }
}
