package com.preuni.shared.presentation.onboarding

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface OnboardingStore : Store<OnboardingStore.Intent, OnboardingStore.State, OnboardingStore.Label> {

    data class State(
        val pageIndex: Int = 0,
        val selectedTrackIds: Set<String> = emptySet(),
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data object NextPage : Intent
        data object PreviousPage : Intent
        data class ToggleTrack(val trackId: String) : Intent
        data object Complete : Intent
    }

    sealed interface Label {
        data object Completed : Label
        data class ValidationError(val message: String) : Label
    }
}

class OnboardingStoreFactory(
    private val storeFactory: StoreFactory,
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
        data object NextPage : Msg
        data object PreviousPage : Msg
        data class TrackToggled(val trackId: String) : Msg
        data object Loading : Msg
        data object DoneLoading : Msg
        data class ErrorReceived(val error: AppError) : Msg
    }

    private inner class Executor :
        CoroutineExecutor<OnboardingStore.Intent, Nothing, OnboardingStore.State, Msg, OnboardingStore.Label>() {

        override fun executeIntent(intent: OnboardingStore.Intent) {
            when (intent) {
                OnboardingStore.Intent.NextPage -> dispatch(Msg.NextPage)
                OnboardingStore.Intent.PreviousPage -> dispatch(Msg.PreviousPage)
                is OnboardingStore.Intent.ToggleTrack -> dispatch(Msg.TrackToggled(intent.trackId))
                OnboardingStore.Intent.Complete -> complete()
            }
        }

        private fun complete() {
            val s = state()
            // Page 3 is track selection — require at least 1 track
            if (s.selectedTrackIds.isEmpty()) {
                publish(OnboardingStore.Label.ValidationError("Select at least one subject track to continue"))
                return
            }
            dispatch(Msg.Loading)
            scope.launch {
                completeOnboarding(s.selectedTrackIds.toList()).fold(
                    onSuccess = { publish(OnboardingStore.Label.Completed) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
                dispatch(Msg.DoneLoading)
            }
        }
    }

    private object ReducerImpl : Reducer<OnboardingStore.State, Msg> {
        private const val TOTAL_PAGES = 4

        override fun OnboardingStore.State.reduce(msg: Msg): OnboardingStore.State = when (msg) {
            Msg.NextPage -> copy(pageIndex = (pageIndex + 1).coerceAtMost(TOTAL_PAGES - 1))
            Msg.PreviousPage -> copy(pageIndex = (pageIndex - 1).coerceAtLeast(0))
            is Msg.TrackToggled -> copy(
                selectedTrackIds = if (msg.trackId in selectedTrackIds)
                    selectedTrackIds - msg.trackId
                else
                    selectedTrackIds + msg.trackId
            )
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            is Msg.ErrorReceived -> copy(error = msg.error, isLoading = false)
        }
    }
}
