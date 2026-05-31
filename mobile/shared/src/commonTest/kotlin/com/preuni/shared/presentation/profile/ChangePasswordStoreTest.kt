package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.error.AppError
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
import kotlin.test.assertNull
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class ChangePasswordStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private fun fakeAuthRepo(
        changePasswordResult: Result<Unit> = Result.success(Unit),
    ): AuthRepository = object : AuthRepository {
        override suspend fun verifyEmail(email: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun resendVerificationOtp(email: String): Result<Unit> = Result.success(Unit)
        override suspend fun register(email: String, password: String, displayName: String): Result<AuthSession> =
            Result.failure(NotImplementedError())
        override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> =
            Result.failure(NotImplementedError())
        override suspend fun logout() {}
        override fun isLoggedIn(): Boolean = true
        override suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> =
            changePasswordResult
        override suspend fun changeEmailRequest(newEmail: String): Result<Unit> = Result.success(Unit)
        override suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun deleteAccount(): Result<Unit> = Result.success(Unit)
    }

    private fun buildStore(repo: AuthRepository) =
        ChangePasswordStoreFactory(DefaultStoreFactory(), repo).create()

    @Test
    fun `Submit with valid credentials emits Label Saved and clears loading`() = runTest {
        val store = buildStore(fakeAuthRepo())
        var labelReceived = false
        val job = launch(Dispatchers.Main) { store.labels.collect { if (it is ChangePasswordStore.Label.Saved) labelReceived = true } }

        store.accept(ChangePasswordStore.Intent.Submit("oldPass", "newPass123"))
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertTrue(labelReceived, "Expected Label.Saved to be published")
        assertFalse(state.isLoading)
        assertNull(state.error)
        job.cancel()
    }

    @Test
    fun `Submit with wrong current password sets error and does not emit Saved`() = runTest {
        val store = buildStore(fakeAuthRepo(
            changePasswordResult = Result.failure(AppError.Validation("current_password", "Wrong password"))
        ))
        var labelReceived = false
        val job = launch { store.labels.collect { if (it is ChangePasswordStore.Label.Saved) labelReceived = true } }

        store.accept(ChangePasswordStore.Intent.Submit("wrongPass", "newPass123"))
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertFalse(labelReceived, "Expected Label.Saved NOT to be published")
        assertFalse(state.isLoading)
        assertTrue(state.error != null, "Expected error to be set")
        job.cancel()
    }

    @Test
    fun `Submit with network error sets NetworkError in state`() = runTest {
        val store = buildStore(fakeAuthRepo(
            changePasswordResult = Result.failure(AppError.NetworkError())
        ))

        store.accept(ChangePasswordStore.Intent.Submit("pass", "newPass123"))
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertFalse(state.isLoading)
        assertTrue(state.error != null)
    }

    @Test
    fun `ClearError resets error to null`() = runTest {
        val store = buildStore(fakeAuthRepo(
            changePasswordResult = Result.failure(AppError.NetworkError())
        ))
        store.accept(ChangePasswordStore.Intent.Submit("pass", "newPass123"))
        advanceUntilIdle()

        store.accept(ChangePasswordStore.Intent.ClearError)
        advanceUntilIdle()

        assertNull(store.stateFlow.first().error)
    }
}
