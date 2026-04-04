package com.preuni.shared.presentation.profile

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
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class ProfileStoreTest {

    private val fakeStudent = Student(
        id = "uid-1",
        displayName = "Ana Lima",
        username = "ana01",
        email = "ana@example.com",
        avatarUrl = null,
        xpTotal = 100,
        streakCount = 5,
        readinessScore = 0.6,
        onboardingCompleted = true,
    )

    private fun fakeRepo(
        getMeResult: Result<Student> = Result.success(fakeStudent),
        updateResult: Result<Student> = Result.success(fakeStudent),
        anonymizeResult: Result<Unit> = Result.success(Unit),
    ): UserRepository = object : UserRepository {
        override suspend fun getMe() = getMeResult
        override suspend fun updateProfile(displayName: String?, username: String?) = updateResult
        override suspend fun getAvatarUploadUrl() = Result.success(AvatarUploadUrl("url", "key"))
        override suspend fun confirmAvatarUpload(objectKey: String) = Result.success(fakeStudent)
        override suspend fun anonymize() = anonymizeResult
    }

    private fun buildStore(repo: UserRepository) =
        ProfileStoreFactory(DefaultStoreFactory(), repo).create()

    // ── Initial load ───────────────────────────────────────────────────────────

    @Test
    fun `LoadProfile intent emits student data on success`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(ProfileStore.Intent.LoadProfile)
        val state = store.stateFlow.first { it.student != null }
        assertEquals(fakeStudent, state.student)
        assertFalse(state.isLoading)
    }

    @Test
    fun `LoadProfile intent emits error on network failure`() = runTest {
        val store = buildStore(fakeRepo(getMeResult = Result.failure(AppError.NetworkError())))
        store.accept(ProfileStore.Intent.LoadProfile)
        val state = store.stateFlow.first { it.error != null }
        assertIs<AppError.NetworkError>(state.error)
    }

    // ── Username update ───────────────────────────────────────────────────────

    @Test
    fun `UpdateUsername with invalid username sets validation error without calling repo`() = runTest {
        var repoCalled = false
        val repo = object : UserRepository by fakeRepo() {
            override suspend fun updateProfile(displayName: String?, username: String?): Result<Student> {
                repoCalled = true
                return Result.success(fakeStudent)
            }
        }
        val store = buildStore(repo)
        store.accept(ProfileStore.Intent.UpdateUsername("UPPERCASE")) // invalid
        val state = store.stateFlow.first { it.error != null }
        assertFalse(repoCalled)
        assertIs<AppError.Validation>(state.error)
    }

    // ── Delete account confirmation ───────────────────────────────────────────

    @Test
    fun `DeleteAccount intent shows confirmation state`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(ProfileStore.Intent.DeleteAccount)
        val state = store.stateFlow.first { it.showDeleteConfirmation }
        assertTrue(state.showDeleteConfirmation)
    }

    @Test
    fun `DismissDeleteConfirmation hides confirmation state`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(ProfileStore.Intent.DeleteAccount)
        store.stateFlow.first { it.showDeleteConfirmation }
        store.accept(ProfileStore.Intent.DismissDeleteConfirmation)
        val state = store.stateFlow.first { !it.showDeleteConfirmation }
        assertFalse(state.showDeleteConfirmation)
    }
}
