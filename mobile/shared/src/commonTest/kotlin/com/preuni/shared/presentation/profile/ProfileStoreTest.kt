package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.domain.user.AvatarUploadUrl
import com.preuni.shared.domain.user.Student
import com.preuni.shared.domain.user.UserRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
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
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class ProfileStoreTest {

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
        override suspend fun updateTracks(trackIds: List<String>): Result<Unit> = Result.success(Unit)
        override suspend fun updateProfile(displayName: String?, username: String?) = updateResult
        override suspend fun getAvatarUploadUrl() = Result.success(AvatarUploadUrl("url", "key"))
        override suspend fun confirmAvatarUpload(objectKey: String) = Result.success(fakeStudent)
        override suspend fun anonymize() = anonymizeResult
    }

    private fun fakeAuthRepo(
        deleteResult: Result<Unit> = Result.success(Unit),
    ): AuthRepository = object : AuthRepository {
        override suspend fun verifyEmail(email: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun register(email: String, password: String, displayName: String): Result<AuthSession> =
            Result.failure(NotImplementedError())
        override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> =
            Result.failure(NotImplementedError())
        override suspend fun logout() {}
        override fun isLoggedIn(): Boolean = true
        override suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> = Result.success(Unit)
        override suspend fun changeEmailRequest(newEmail: String): Result<Unit> = Result.success(Unit)
        override suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun deleteAccount(): Result<Unit> = deleteResult
    }

    private fun buildStore(repo: UserRepository, authRepo: AuthRepository = fakeAuthRepo()) =
        ProfileStoreFactory(DefaultStoreFactory(), repo, authRepo).create()

    // ── Initial load ───────────────────────────────────────────────────────────

    @Test
    fun `LoadProfile intent emits student data on success`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(ProfileStore.Intent.LoadProfile)
        advanceUntilIdle()
        val state = store.stateFlow.first()
        assertEquals(fakeStudent, state.student)
        assertFalse(state.isLoading)
    }

    @Test
    fun `LoadProfile intent emits error on network failure`() = runTest {
        val store = buildStore(fakeRepo(getMeResult = Result.failure(AppError.NetworkError())))
        store.accept(ProfileStore.Intent.LoadProfile)
        advanceUntilIdle()
        val state = store.stateFlow.first()
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
        advanceUntilIdle()
        val state = store.stateFlow.first()
        assertFalse(repoCalled)
        assertIs<AppError.Validation>(state.error)
    }

    // ── Delete account confirmation ───────────────────────────────────────────

    @Test
    fun `DeleteAccount intent shows confirmation state`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(ProfileStore.Intent.DeleteAccount)
        val state = store.stateFlow.first()
        assertTrue(state.showDeleteConfirmation)
    }

    @Test
    fun `DismissDeleteConfirmation hides confirmation state`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(ProfileStore.Intent.DeleteAccount)
        store.accept(ProfileStore.Intent.DismissDeleteConfirmation)
        val state = store.stateFlow.first()
        assertFalse(state.showDeleteConfirmation)
    }

    // ── Account deletion uses authRepository ─────────────────────────────────

    @Test
    fun `ConfirmDeleteAccount calls authRepository deleteAccount not userRepository anonymize`() = runTest {
        var authRepoCalled = false
        var userRepoCalled = false
        val authRepo = object : AuthRepository by fakeAuthRepo() {
            override suspend fun deleteAccount(): Result<Unit> {
                authRepoCalled = true
                return Result.success(Unit)
            }
        }
        val userRepo = object : UserRepository by fakeRepo() {
            override suspend fun anonymize(): Result<Unit> {
                userRepoCalled = true
                return Result.success(Unit)
            }
        }
        val store = buildStore(userRepo, authRepo)
        store.accept(ProfileStore.Intent.ConfirmDeleteAccount)
        advanceUntilIdle()

        assertTrue(authRepoCalled, "Expected authRepository.deleteAccount() to be called")
        assertFalse(userRepoCalled, "Expected userRepository.anonymize() NOT to be called")
    }

    @Test
    fun `ConfirmDeleteAccount on success publishes Label AccountDeleted`() = runTest {
        val store = buildStore(fakeRepo())
        var labelReceived = false
        val job = launch(Dispatchers.Main) {
            store.labels.collect { if (it is ProfileStore.Label.AccountDeleted) labelReceived = true }
        }
        store.accept(ProfileStore.Intent.ConfirmDeleteAccount)
        advanceUntilIdle()

        assertTrue(labelReceived, "Expected Label.AccountDeleted to be published")
        job.cancel()
    }

    // ── Username saved label ─────────────────────────────────────────────────

    @Test
    fun `UpdateUsername on success publishes Label UsernameSaved`() = runTest {
        val store = buildStore(fakeRepo())
        var labelReceived = false
        val job = launch(Dispatchers.Main) {
            store.labels.collect { if (it is ProfileStore.Label.UsernameSaved) labelReceived = true }
        }
        store.accept(ProfileStore.Intent.UpdateUsername("validname"))
        advanceUntilIdle()

        assertTrue(labelReceived, "Expected Label.UsernameSaved to be published")
        job.cancel()
    }
}
