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
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class ChangeEmailStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private fun fakeAuthRepo(
        requestResult: Result<Unit> = Result.success(Unit),
        confirmResult: Result<Unit> = Result.success(Unit),
    ): AuthRepository = object : AuthRepository {
        override suspend fun verifyEmail(email: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun resendVerificationOtp(email: String): Result<Unit> = Result.success(Unit)
        override suspend fun register(email: String, password: String, displayName: String): Result<AuthSession> =
            Result.failure(NotImplementedError())
        override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> =
            Result.failure(NotImplementedError())
        override suspend fun logout() {}
        override fun isLoggedIn(): Boolean = true
        override suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> = Result.success(Unit)
        override suspend fun changeEmailRequest(newEmail: String): Result<Unit> = requestResult
        override suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> = confirmResult
        override suspend fun deleteAccount(): Result<Unit> = Result.success(Unit)
    }

    private fun buildStore(repo: AuthRepository) =
        ChangeEmailStoreFactory(DefaultStoreFactory(), repo).create()

    @Test
    fun `RequestCode with valid email transitions to CONFIRM phase`() = runTest {
        val store = buildStore(fakeAuthRepo())
        store.accept(ChangeEmailStore.Intent.RequestCode("new@example.com"))
        // UnconfinedTestDispatcher runs the request synchronously; avoid advanceUntilIdle()
        // which would also drain the 60s cooldown countdown to 0.

        val state = store.stateFlow.first()
        assertEquals(ChangeEmailStore.Phase.CONFIRM, state.phase)
        assertEquals("new@example.com", state.newEmail)
        assertEquals(60, state.resendCooldown)
        assertFalse(state.isLoading)
        assertNull(state.error)
    }

    @Test
    fun `RequestCode with conflict stays in REQUEST phase and sets error`() = runTest {
        val store = buildStore(fakeAuthRepo(
            requestResult = Result.failure(AppError.Conflict("Email already in use"))
        ))
        store.accept(ChangeEmailStore.Intent.RequestCode("taken@example.com"))
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertEquals(ChangeEmailStore.Phase.REQUEST, state.phase)
        assertNotNull(state.error)
        assertFalse(state.isLoading)
    }

    @Test
    fun `ConfirmCode with valid OTP publishes Label EmailChanged`() = runTest {
        val store = buildStore(fakeAuthRepo())
        store.accept(ChangeEmailStore.Intent.RequestCode("new@example.com"))

        var emailChangedLabel: ChangeEmailStore.Label.EmailChanged? = null
        val job = launch(Dispatchers.Main) {
            store.labels.collect { if (it is ChangeEmailStore.Label.EmailChanged) emailChangedLabel = it }
        }

        store.accept(ChangeEmailStore.Intent.ConfirmCode("123456"))
        advanceUntilIdle()

        assertNotNull(emailChangedLabel, "Expected Label.EmailChanged to be published")
        assertEquals("new@example.com", emailChangedLabel?.newEmail)
        job.cancel()
    }

    @Test
    fun `ConfirmCode with wrong OTP sets error and does not emit EmailChanged`() = runTest {
        val store = buildStore(fakeAuthRepo(
            confirmResult = Result.failure(AppError.Validation("otp", "Invalid code"))
        ))
        store.accept(ChangeEmailStore.Intent.RequestCode("new@example.com"))

        var labelReceived = false
        val job = launch(Dispatchers.Main) { store.labels.collect { labelReceived = true } }

        store.accept(ChangeEmailStore.Intent.ConfirmCode("000000"))
        advanceUntilIdle()

        assertFalse(labelReceived, "Expected no label to be published")
        assertNotNull(store.stateFlow.first().error)
        job.cancel()
    }

    @Test
    fun `ResendCode resets cooldown to 60`() = runTest {
        val store = buildStore(fakeAuthRepo())
        store.accept(ChangeEmailStore.Intent.RequestCode("new@example.com"))
        // No advanceUntilIdle() — would drain the 60s countdown to 0

        store.accept(ChangeEmailStore.Intent.ResendCode)
        // Same: state updates synchronously with UnconfinedTestDispatcher

        assertEquals(60, store.stateFlow.first().resendCooldown)
    }
}
