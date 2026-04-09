package com.preuni.shared.presentation.welcome

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor

interface WelcomeStore : Store<WelcomeStore.Intent, WelcomeStore.State, WelcomeStore.Label> {

    data class State(val pageIndex: Int = 0)

    sealed interface Intent {
        data object NextPage : Intent
        data object PreviousPage : Intent
        data object Skip : Intent
    }

    sealed interface Label {
        data object Completed : Label
    }
}

class WelcomeStoreFactory(private val storeFactory: StoreFactory) {

    fun create(): WelcomeStore =
        object : WelcomeStore, Store<WelcomeStore.Intent, WelcomeStore.State, WelcomeStore.Label>
        by storeFactory.create(
            name = "WelcomeStore",
            initialState = WelcomeStore.State(),
            executorFactory = { Executor() },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data object NextPage : Msg
        data object PreviousPage : Msg
    }

    private inner class Executor :
        CoroutineExecutor<WelcomeStore.Intent, Nothing, WelcomeStore.State, Msg, WelcomeStore.Label>() {

        override fun executeIntent(intent: WelcomeStore.Intent) {
            when (intent) {
                WelcomeStore.Intent.NextPage -> {
                    if (state().pageIndex >= TOTAL_PAGES - 1) {
                        publish(WelcomeStore.Label.Completed)
                    } else {
                        dispatch(Msg.NextPage)
                    }
                }
                WelcomeStore.Intent.PreviousPage -> dispatch(Msg.PreviousPage)
                WelcomeStore.Intent.Skip -> publish(WelcomeStore.Label.Completed)
            }
        }
    }

    private object ReducerImpl : Reducer<WelcomeStore.State, Msg> {
        override fun WelcomeStore.State.reduce(msg: Msg): WelcomeStore.State = when (msg) {
            Msg.NextPage -> copy(pageIndex = pageIndex + 1)
            Msg.PreviousPage -> copy(pageIndex = (pageIndex - 1).coerceAtLeast(0))
        }
    }

    companion object {
        const val TOTAL_PAGES = 3
    }
}
