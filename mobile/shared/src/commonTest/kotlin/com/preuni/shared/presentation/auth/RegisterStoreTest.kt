package com.preuni.shared.presentation.auth

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
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertNotNull
import kotlin.test.assertNull

@OptIn(ExperimentalCoroutinesApi::class)
class RegisterStoreTest {

    @BeforeTest
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterTest
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun fakeRepo(
        result: Result<AuthSession> = Result.success(
            AuthSession("uid-1", "user@example.com", "access-tok", "refresh-tok")
        ),
    ): AuthRepository = object : AuthRepository {
        override suspend fun register(email: String, password: String, displayName: String) = result
        override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> =
            Result.failure(AppError.Unknown())
        override suspend fun verifyEmail(email: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun resendVerificationOtp(email: String): Result<Unit> = Result.success(Unit)
        override suspend fun logout() {}
        override fun isLoggedIn() = false
        override suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> = Result.success(Unit)
        override suspend fun changeEmailRequest(newEmail: String): Result<Unit> = Result.success(Unit)
        override suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> = Result.success(Unit)
        override suspend fun deleteAccount(): Result<Unit> = Result.success(Unit)
    }

    private fun buildStore(repo: AuthRepository): RegisterStore =
        RegisterStoreFactory(DefaultStoreFactory(), repo).create()

    private fun fillValidFields(store: RegisterStore) {
        store.accept(RegisterStore.Intent.UpdateDisplayName("Test User"))
        store.accept(RegisterStore.Intent.UpdateEmail("valid@example.com"))
        store.accept(RegisterStore.Intent.UpdateUsername("testuser"))
        store.accept(RegisterStore.Intent.UpdatePassword("Passw0rd!"))
        store.accept(RegisterStore.Intent.UpdateConfirmPassword("Passw0rd!"))
    }

    // ── Submit with valid fields emits Registered label ───────────────────────

    @Test
    fun submit_validFields_emitsRegisteredLabel() = runTest {
        val store = buildStore(fakeRepo())
        fillValidFields(store)

        var receivedLabel: RegisterStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { receivedLabel = it } }

        store.accept(RegisterStore.Intent.Submit)
        advanceUntilIdle()
        job.cancel()

        assertIs<RegisterStore.Label.Registered>(receivedLabel)
    }

    // ── Submit with empty email sets validation error without calling repo ─────

    @Test
    fun submit_emptyEmail_setsValidationError_withoutCallingRepo() = runTest {
        var repoCalled = false
        val repo = object : AuthRepository {
            override suspend fun register(email: String, password: String, displayName: String): Result<AuthSession> {
                repoCalled = true
                return Result.failure(AppError.Unknown())
            }
            override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> =
                Result.failure(AppError.Unknown())
            override suspend fun verifyEmail(email: String, otp: String): Result<Unit> = Result.success(Unit)
            override suspend fun resendVerificationOtp(email: String): Result<Unit> = Result.success(Unit)
            override suspend fun logout() {}
            override fun isLoggedIn() = false
            override suspend fun changePassword(currentPassword: String, newPassword: String): Result<Unit> = Result.success(Unit)
            override suspend fun changeEmailRequest(newEmail: String): Result<Unit> = Result.success(Unit)
            override suspend fun changeEmailConfirm(newEmail: String, otp: String): Result<Unit> = Result.success(Unit)
            override suspend fun deleteAccount(): Result<Unit> = Result.success(Unit)
        }
        val store = buildStore(repo)
        store.accept(RegisterStore.Intent.UpdatePassword("Passw0rd!"))
        store.accept(RegisterStore.Intent.UpdateConfirmPassword("Passw0rd!"))
        store.accept(RegisterStore.Intent.Submit)
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertFalse(repoCalled, "repository must not be called when validation fails")
        assertNotNull(state.emailError, "emailError must be set")
        assertFalse(state.isLoading)
    }

    // ── Submit with repo returning Conflict sets globalError ──────────────────

    @Test
    fun submit_repositoryReturnsConflict_setsGlobalError() = runTest {
        val repo = fakeRepo(Result.failure(AppError.Conflict("email taken")))
        val store = buildStore(repo)
        fillValidFields(store)

        store.accept(RegisterStore.Intent.Submit)
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertIs<AppError.Conflict>(state.globalError)
        assertFalse(state.isLoading)
        assertNull(state.emailError)
    }
}
