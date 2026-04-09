package com.preuni.shared.presentation.learn

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.domain.content.Track
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.learn.ModuleNode
import com.preuni.shared.domain.learn.NodeState
import com.preuni.shared.domain.learn.generateStubModules
import kotlinx.coroutines.launch

interface LearnStore : Store<LearnStore.Intent, LearnStore.State, LearnStore.Label> {

    data class State(
        val activeTrackId: String? = null,
        val activeTrack: Track? = null,
        val modules: List<ModuleNode> = emptyList(),
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data object Load : Intent
        data class TapNode(val nodeId: String) : Intent
        data object Retry : Intent
    }

    sealed interface Label {
        data class OpenLesson(val moduleId: String) : Label
        data object ShowLockedMessage : Label
    }
}

class LearnStoreFactory(
    private val storeFactory: StoreFactory,
    private val getActiveTrackId: () -> String?,
    private val getTracks: suspend () -> Result<List<Track>>,
) {

    fun create(): LearnStore =
        object : LearnStore, Store<LearnStore.Intent, LearnStore.State, LearnStore.Label>
        by storeFactory.create(
            name = "LearnStore",
            initialState = LearnStore.State(),
            executorFactory = { Executor() },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data object Loading : Msg
        data class Loaded(
            val activeTrackId: String?,
            val activeTrack: Track?,
            val modules: List<ModuleNode>,
        ) : Msg
    }

    private inner class Executor :
        CoroutineExecutor<LearnStore.Intent, Nothing, LearnStore.State, Msg, LearnStore.Label>() {

        override fun executeIntent(intent: LearnStore.Intent) {
            when (intent) {
                LearnStore.Intent.Load, LearnStore.Intent.Retry -> load()
                is LearnStore.Intent.TapNode -> tapNode(intent.nodeId)
            }
        }

        private fun load() {
            dispatch(Msg.Loading)
            scope.launch {
                val trackId = getActiveTrackId()
                if (trackId == null) {
                    dispatch(Msg.Loaded(activeTrackId = null, activeTrack = null, modules = emptyList()))
                    return@launch
                }
                // Backend failure is non-fatal — stub modules don't need backend data
                val track = getTracks().getOrNull()?.find { it.id == trackId }
                dispatch(Msg.Loaded(
                    activeTrackId = trackId,
                    activeTrack = track,
                    modules = generateStubModules(15),
                ))
            }
        }

        private fun tapNode(nodeId: String) {
            val module = state().modules.find { it.id == nodeId } ?: return
            when (module.state) {
                NodeState.ACTIVE -> publish(LearnStore.Label.OpenLesson(nodeId))
                NodeState.LOCKED -> publish(LearnStore.Label.ShowLockedMessage)
                NodeState.COMPLETED -> Unit
            }
        }
    }

    private object ReducerImpl : Reducer<LearnStore.State, Msg> {
        override fun LearnStore.State.reduce(msg: Msg): LearnStore.State = when (msg) {
            Msg.Loading -> copy(isLoading = true, error = null)
            is Msg.Loaded -> copy(
                isLoading = false,
                activeTrackId = msg.activeTrackId,
                activeTrack = msg.activeTrack,
                modules = msg.modules,
            )
        }
    }
}
