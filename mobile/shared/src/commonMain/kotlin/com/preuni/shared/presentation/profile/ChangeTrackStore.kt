package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.launch

interface ChangeTrackStore : Store<ChangeTrackStore.Intent, ChangeTrackStore.State, ChangeTrackStore.Label> {

    data class State(
        val selectedTrackId: String? = null,
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data class Load(val initialTrackId: String?) : Intent
        data class SelectTrack(val trackId: String) : Intent
        data object Save : Intent
    }

    sealed interface Label {
        data object Saved : Label
        data class ValidationError(val message: String) : Label
    }
}

class ChangeTrackStoreFactory(
    private val storeFactory: StoreFactory,
    private val setActiveTrackId: (String) -> Unit,
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
        data class TrackLoaded(val trackId: String?) : Msg
        data class TrackSelected(val trackId: String) : Msg
    }

    private inner class Executor :
        CoroutineExecutor<ChangeTrackStore.Intent, Nothing, ChangeTrackStore.State, Msg, ChangeTrackStore.Label>() {

        override fun executeIntent(intent: ChangeTrackStore.Intent) {
            when (intent) {
                is ChangeTrackStore.Intent.Load -> dispatch(Msg.TrackLoaded(intent.initialTrackId))
                is ChangeTrackStore.Intent.SelectTrack -> dispatch(Msg.TrackSelected(intent.trackId))
                ChangeTrackStore.Intent.Save -> save()
            }
        }

        private fun save() {
            val trackId = state().selectedTrackId
            if (trackId == null) {
                publish(ChangeTrackStore.Label.ValidationError("Selecione uma matéria para continuar."))
                return
            }
            // Save locally and navigate immediately — fire-and-forget backend sync
            setActiveTrackId(trackId)
            publish(ChangeTrackStore.Label.Saved)
            scope.launch {
                updateTracks(listOf(trackId))
            }
        }
    }

    private object ReducerImpl : Reducer<ChangeTrackStore.State, Msg> {
        override fun ChangeTrackStore.State.reduce(msg: Msg): ChangeTrackStore.State = when (msg) {
            is Msg.TrackLoaded -> copy(selectedTrackId = msg.trackId)
            is Msg.TrackSelected -> copy(selectedTrackId = msg.trackId)
        }
    }
}
