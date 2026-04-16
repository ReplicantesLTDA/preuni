package com.preuni.shared.presentation.home

import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.user.AvatarUploadUrl
import com.preuni.shared.domain.user.Student
import com.preuni.shared.domain.user.UserRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import kotlin.test.AfterTest
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertNull

@OptIn(ExperimentalCoroutinesApi::class)
class HomeStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private val fakeStudent = Student(
        id = "uid-1",
        displayName = "Ana Lima",
        username = "ana01",
        email = "ana@example.com",
        avatarUrl = null,
        xpTotal = 50,
        streakCount = 3,
        readinessScore = 0.4,
        onboardingCompleted = true,
    )

    private fun fakeRepo(
        getMeResult: Result<Student> = Result.success(fakeStudent),
    ): UserRepository = object : UserRepository {
        override suspend fun getMe() = getMeResult
        override suspend fun updateTracks(trackIds: List<String>): Result<Unit> = Result.success(Unit)
        override suspend fun updateProfile(displayName: String?, username: String?) = Result.success(fakeStudent)
        override suspend fun getAvatarUploadUrl() = Result.success(AvatarUploadUrl("url", "key"))
        override suspend fun confirmAvatarUpload(objectKey: String) = Result.success(fakeStudent)
        override suspend fun anonymize() = Result.success(Unit)
    }

    private fun buildStore(repo: UserRepository) =
        HomeStoreFactory(DefaultStoreFactory(), repo).create()

    // ── Load ──────────────────────────────────────────────────────────────────

    @Test
    fun `Load intent emits student data on success`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()
        val state = store.stateFlow.first()
        assertEquals(fakeStudent, state.student)
        assertFalse(state.isLoading)
        assertNull(state.error)
    }

    // ── Network error ─────────────────────────────────────────────────────────

    @Test
    fun `Load intent emits error state on network failure`() = runTest {
        val store = buildStore(fakeRepo(getMeResult = Result.failure(AppError.NetworkError())))
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()
        val state = store.stateFlow.first()
        assertIs<AppError.NetworkError>(state.error)
        assertFalse(state.isLoading)
    }

    // ── Retry ─────────────────────────────────────────────────────────────────

    @Test
    fun `Retry intent re-fetches student data after error`() = runTest {
        var callCount = 0
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return if (callCount == 1) Result.failure(AppError.NetworkError())
                else Result.success(fakeStudent)
            }
        }
        val store = buildStore(repo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        store.accept(HomeStore.Intent.Retry)
        advanceUntilIdle()
        val state = store.stateFlow.first()
        assertEquals(fakeStudent, state.student)
    }

    // ── Auto-retry ────────────────────────────────────────────────────────────

    @Test
    fun `load_transientFailure_retriesAndSucceeds`() = runTest {
        var callCount = 0
        val repo = fakeRepo(getMeResult = Result.failure(AppError.NetworkError()))
        val transientRepo = object : UserRepository by repo {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return if (callCount < 3) Result.failure(AppError.NetworkError())
                else Result.success(fakeStudent)
            }
        }
        val store = buildStore(transientRepo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        assertNull(store.stateFlow.first().error)
        assertEquals(fakeStudent, store.stateFlow.first().student)
    }

    @Test
    fun `load_persistentFailure_setsErrorAfterThreeAttempts`() = runTest {
        var callCount = 0
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return Result.failure(AppError.NetworkError())
            }
        }
        val store = buildStore(repo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        assertIs<AppError>(store.stateFlow.first().error)
        assertNull(store.stateFlow.first().student)
        assertEquals(3, callCount)
    }

    // ── Non-transient errors do not retry ─────────────────────────────────────

    @Test
    fun `load_nonTransientError_stopsImmediatelyWithoutRetry`() = runTest {
        var callCount = 0
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return Result.failure(AppError.Unauthorized())
            }
        }
        val store = buildStore(repo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        assertIs<AppError.Unauthorized>(store.stateFlow.first().error)
        assertEquals(1, callCount, "Non-transient error must not be retried")
    }

    @Test
    fun `load_validationError_stopsImmediatelyWithoutRetry`() = runTest {
        var callCount = 0
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return Result.failure(AppError.Validation("field", "bad"))
            }
        }
        val store = buildStore(repo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        assertEquals(1, callCount, "Validation error must not be retried")
        assertIs<AppError.Validation>(store.stateFlow.first().error)
    }

    // ── Timeout classification ─────────────────────────────────────────────────

    @Test
    fun `load_networkErrorIsTransient_retriesUpToMaxAttempts`() = runTest {
        var callCount = 0
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return Result.failure(AppError.NetworkError())
            }
        }
        val store = buildStore(repo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        // NetworkError is transient → should retry until MAX_ATTEMPTS (3)
        assertEquals(3, callCount, "NetworkError should be retried up to MAX_ATTEMPTS")
        assertIs<AppError.NetworkError>(store.stateFlow.first().error)
    }

    // ── Manual Retry re-requests backend ──────────────────────────────────────

    @Test
    fun `Retry_afterNonTransientError_sendsNewRequest`() = runTest {
        var callCount = 0
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun getMe(): Result<Student> {
                callCount++
                return if (callCount == 1) Result.failure(AppError.Unauthorized())
                else Result.success(fakeStudent)
            }
        }
        val store = buildStore(repo)
        store.accept(HomeStore.Intent.Load)
        advanceUntilIdle()

        // First attempt hit Unauthorized (not retried)
        assertEquals(1, callCount)
        assertIs<AppError.Unauthorized>(store.stateFlow.first().error)

        // Manual Retry should trigger a fresh request
        store.accept(HomeStore.Intent.Retry)
        advanceUntilIdle()

        assertEquals(2, callCount)
        assertEquals(fakeStudent, store.stateFlow.first().student)
        assertNull(store.stateFlow.first().error)
    }
}
