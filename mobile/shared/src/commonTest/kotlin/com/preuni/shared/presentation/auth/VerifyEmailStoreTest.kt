package com.preuni.shared.presentation.auth

import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.error.AppError
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
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class VerifyEmailStoreTest {

    @BeforeTest
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterTest
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun buildStore(
        onSubmit: suspend (email: String, code: String) -> Result<Unit> = { _, _ -> Result.success(Unit) },
        onResend: suspend (email: String) -> Result<Unit> = { _ -> Result.success(Unit) },
    ): VerifyEmailStore =
        VerifyEmailStoreFactory(
            storeFactory = DefaultStoreFactory(),
            email = "test@example.com",
            onSubmit = onSubmit,
            onResend = onResend,
        ).create()

    // ── Submit valid code sets verified = true ────────────────────────────────

    @Test
    fun submit_validCode_setsVerifiedTrue() = runTest {
        val store = buildStore(onSubmit = { _, _ -> Result.success(Unit) })
        store.accept(VerifyEmailStore.Intent.UpdateCode("123456"))
        store.accept(VerifyEmailStore.Intent.Submit)
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertTrue(state.verified)
        assertFalse(state.isLoading)
        assertNull(state.error)
    }

    // ── Submit with repo returning error sets error message ───────────────────

    @Test
    fun submit_repositoryError_setsErrorMessage() = runTest {
        val store = buildStore(
            onSubmit = { _, _ -> Result.failure(AppError.Validation("otp", "invalid or expired code")) }
        )
        store.accept(VerifyEmailStore.Intent.UpdateCode("000000"))
        store.accept(VerifyEmailStore.Intent.Submit)
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertNotNull(state.error)
        assertFalse(state.isLoading)
        assertFalse(state.verified)
    }
}
