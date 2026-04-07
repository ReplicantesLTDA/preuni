package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface ChangeTrackStore : Store<ChangeTrackStore.Intent, ChangeTrackStore.State, ChangeTrackStore.Label> {

    data class State(
        val selectedTrackIds: Set<String> = emptySet(),
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data class Load(val initialIds: Set<String>) : Intent
        data class ToggleTrack(val trackId: String) : Intent
        data object Save : Intent
    }

    sealed interface Label {
        data object Saved : Label
        data class ValidationError(val message: String) : Label
    }
}

class ChangeTrackStoreFactory(
    private val storeFactory: StoreFactory,
    private val updateTracks: suspend (List<String>) -> Result<Unit>,
) {
    fun create(): ChangeTrackStore =
        object : ChangeTrackStore, Store<ChangeTrackStore.Intent, ChangeTrackStore.State, ChangeTrackStore.Label>
        by storeFactory.create(
            name = "ChangeTrackStore",
            initialState = ChangeTrackStore.State(),
            executorFactory = { Executor() },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data class TracksLoaded(val ids: Set<String>) : Msg
        data class TrackToggled(val trackId: String) : Msg
        data object Loading : Msg
        data object DoneLoading : Msg
        data class ErrorReceived(val error: AppError) : Msg
    }

    private inner class Executor :
        CoroutineExecutor<ChangeTrackStore.Intent, Nothing, ChangeTrackStore.State, Msg, ChangeTrackStore.Label>() {

        override fun executeIntent(intent: ChangeTrackStore.Intent) {
            when (intent) {
                is ChangeTrackStore.Intent.Load -> dispatch(Msg.TracksLoaded(intent.initialIds))
                is ChangeTrackStore.Intent.ToggleTrack -> dispatch(Msg.TrackToggled(intent.trackId))
                ChangeTrackStore.Intent.Save -> save()
            }
        }

        private fun save() {
            val ids = state().selectedTrackIds
            if (ids.isEmpty()) {
                publish(ChangeTrackStore.Label.ValidationError("Selecione ao menos uma matéria."))
                return
            }
            dispatch(Msg.Loading)
            scope.launch {
                updateTracks(ids.toList()).fold(
                    onSuccess = { publish(ChangeTrackStore.Label.Saved) },
                    onFailure = { dispatch(Msg.ErrorReceived(it as? AppError ?: AppError.Unknown())) },
                )
                dispatch(Msg.DoneLoading)
            }
        }
    }

    private object ReducerImpl : Reducer<ChangeTrackStore.State, Msg> {
        override fun ChangeTrackStore.State.reduce(msg: Msg): ChangeTrackStore.State = when (msg) {
            is Msg.TracksLoaded -> copy(selectedTrackIds = msg.ids)
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
