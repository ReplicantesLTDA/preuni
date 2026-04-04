package com.preuni.shared.presentation.auth

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class LoginStoreTest {

    private fun fakeRepo(
        result: Result<AuthSession> = Result.success(
            AuthSession("user-1", "user@example.com", "access-tok", "refresh-tok")
        ),
    ): AuthRepository = object : AuthRepository {
        override suspend fun login(emailOrUsername: String, password: String) = result
        override suspend fun logout() {}
        override fun isLoggedIn() = false
    }

    private fun buildStore(repo: AuthRepository) =
        LoginStoreFactory(DefaultStoreFactory(), repo).create()

    // ── Initial state ──────────────────────────────────────────────────────────

    @Test
    fun `initial state is idle with empty fields`() = runTest {
        val store = buildStore(fakeRepo())
        val state = store.stateFlow.first()
        assertEquals("", state.emailOrUsername)
        assertEquals("", state.password)
        assertFalse(state.isLoading)
        assertEquals(null, state.error)
    }

    // ── Field updates ──────────────────────────────────────────────────────────

    @Test
    fun `UpdateEmail intent changes emailOrUsername`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(LoginStore.Intent.UpdateEmailOrUsername("student@example.com"))
        assertEquals("student@example.com", store.stateFlow.first().emailOrUsername)
    }

    @Test
    fun `UpdatePassword intent changes password`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(LoginStore.Intent.UpdatePassword("Abc123"))
        assertEquals("Abc123", store.stateFlow.first().password)
    }

    // ── Submit — success ───────────────────────────────────────────────────────

    @Test
    fun `Submit with valid credentials emits LoggedIn label`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(LoginStore.Intent.UpdateEmailOrUsername("student@example.com"))
        store.accept(LoginStore.Intent.UpdatePassword("Abc123"))

        val label = store.labels.first { true }
        store.accept(LoginStore.Intent.Submit)

        assertIs<LoginStore.Label.LoggedIn>(label.also {})
    }

    @Test
    fun `Submit transitions through loading then clears it on success`() = runTest {
        val store = buildStore(fakeRepo())
        store.accept(LoginStore.Intent.UpdateEmailOrUsername("student@example.com"))
        store.accept(LoginStore.Intent.UpdatePassword("Abc123"))
        store.accept(LoginStore.Intent.Submit)

        // After success isLoading must be false
        val state = store.stateFlow.first()
        assertFalse(state.isLoading)
        assertEquals(null, state.error)
    }

    // ── Submit — invalid credentials ───────────────────────────────────────────

    @Test
    fun `Submit with server Unauthorized sets error state`() = runTest {
        val repo = fakeRepo(Result.failure(AppError.Unauthorized("invalid credentials")))
        val store = buildStore(repo)
        store.accept(LoginStore.Intent.UpdateEmailOrUsername("student@example.com"))
        store.accept(LoginStore.Intent.UpdatePassword("Abc123"))
        store.accept(LoginStore.Intent.Submit)

        val state = store.stateFlow.first()
        assertFalse(state.isLoading)
        assertIs<AppError.Unauthorized>(state.error)
    }

    // ── Submit — empty fields (client-side guard) ──────────────────────────────

    @Test
    fun `Submit with empty email sets validation error without calling repo`() = runTest {
        var repoCalled = false
        val repo = object : AuthRepository {
            override suspend fun login(emailOrUsername: String, password: String): Result<AuthSession> {
                repoCalled = true
                return Result.failure(AppError.Unknown())
            }
            override suspend fun logout() {}
            override fun isLoggedIn() = false
        }
        val store = buildStore(repo)
        store.accept(LoginStore.Intent.UpdatePassword("Abc123"))
        store.accept(LoginStore.Intent.Submit)

        assertFalse(repoCalled)
        assertIs<AppError.Validation>(store.stateFlow.first().error)
    }
}
