package com.preuni.shared.presentation.home

import com.arkivanov.mvikotlin.core.store.Reducer
import com.arkivanov.mvikotlin.core.store.Store
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.arkivanov.mvikotlin.extensions.coroutines.CoroutineExecutor
import com.preuni.shared.data.network.RetryPolicy
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.user.Student
import com.preuni.shared.domain.user.UserRepository
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

interface HomeStore : Store<HomeStore.Intent, HomeStore.State, Nothing> {

    data class State(
        val student: Student? = null,
        val isLoading: Boolean = false,
        val error: AppError? = null,
    )

    sealed interface Intent {
        data object Load : Intent
        data object Retry : Intent
    }
}

class HomeStoreFactory(
    private val storeFactory: StoreFactory,
    private val userRepository: UserRepository,
) {
    fun create(): HomeStore =
        object : HomeStore, Store<HomeStore.Intent, HomeStore.State, Nothing>
        by storeFactory.create(
            name = "HomeStore",
            initialState = HomeStore.State(),
            executorFactory = { Executor(userRepository) },
            reducer = ReducerImpl,
        ) {}

    private sealed interface Msg {
        data object Loading : Msg
        data object DoneLoading : Msg
        data class StudentLoaded(val student: Student) : Msg
        data class ErrorReceived(val error: AppError) : Msg
    }

    private inner class Executor(private val repo: UserRepository) :
        CoroutineExecutor<HomeStore.Intent, Nothing, HomeStore.State, Msg, Nothing>() {

        override fun executeIntent(intent: HomeStore.Intent) {
            when (intent) {
                HomeStore.Intent.Load, HomeStore.Intent.Retry -> load()
            }
        }

        private fun load() {
            dispatch(Msg.Loading)
            scope.launch {
                var lastError: AppError = AppError.Unknown()
                for (attempt in 0 until RetryPolicy.MAX_ATTEMPTS) {
                    if (attempt > 0) delay(RetryPolicy.delayMillis(attempt))
                    val result = repo.getMe()
                    if (result.isSuccess) {
                        dispatch(Msg.StudentLoaded(result.getOrThrow()))
                        dispatch(Msg.DoneLoading)
                        return@launch
                    }
                    lastError = result.exceptionOrNull() as? AppError ?: AppError.Unknown()
                    // Non-transient errors (auth, validation, etc.) must not be retried.
                    if (!RetryPolicy.isTransient(lastError)) break
                }
                dispatch(Msg.ErrorReceived(lastError))
                dispatch(Msg.DoneLoading)
            }
        }
    }

    private object ReducerImpl : Reducer<HomeStore.State, Msg> {
        override fun HomeStore.State.reduce(msg: Msg): HomeStore.State = when (msg) {
            Msg.Loading -> copy(isLoading = true, error = null)
            Msg.DoneLoading -> copy(isLoading = false)
            is Msg.StudentLoaded -> copy(student = msg.student, isLoading = false)
            is Msg.ErrorReceived -> copy(error = msg.error, isLoading = false)
        }
    }
}
