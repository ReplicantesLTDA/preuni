package com.preuni.shared.presentation.home

import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.user.AvatarUploadUrl
import com.preuni.shared.domain.user.Student
import com.preuni.shared.domain.user.UserRepository
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertNull

@OptIn(ExperimentalCoroutinesApi::class)
class HomeStoreTest {

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
        val state = store.stateFlow.first { it.student != null }
        assertEquals(fakeStudent, state.student)
        assertFalse(state.isLoading)
        assertNull(state.error)
    }

    // ── Network error ─────────────────────────────────────────────────────────

    @Test
    fun `Load intent emits error state on network failure`() = runTest {
        val store = buildStore(fakeRepo(getMeResult = Result.failure(AppError.NetworkError())))
        store.accept(HomeStore.Intent.Load)
        val state = store.stateFlow.first { it.error != null }
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
        store.stateFlow.first { it.error != null }

        store.accept(HomeStore.Intent.Retry)
        val state = store.stateFlow.first { it.student != null }
        assertEquals(fakeStudent, state.student)
    }
}
