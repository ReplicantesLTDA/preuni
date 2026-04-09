package com.preuni.shared.presentation.onboarding

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
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
import kotlin.test.assertIs
import kotlin.test.assertNull

@OptIn(ExperimentalCoroutinesApi::class)
class OnboardingStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private fun buildStore(
        setActiveTrackId: (String) -> Unit = {},
        completeOnboarding: suspend (List<String>) -> Result<Unit> = { Result.success(Unit) },
    ): OnboardingStore =
        OnboardingStoreFactory(
            storeFactory = DefaultStoreFactory(),
            setActiveTrackId = setActiveTrackId,
            completeOnboarding = completeOnboarding,
        ).create()

    // ── Track selection ───────────────────────────────────────────────────────

    @Test
    fun `SelectTrack sets selectedTrackId`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.SelectTrack("track-math"))
        assertEquals("track-math", store.stateFlow.first().selectedTrackId)
    }

    @Test
    fun `SelectTrack replaces previous selection`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.SelectTrack("track-math"))
        store.accept(OnboardingStore.Intent.SelectTrack("track-port"))
        assertEquals("track-port", store.stateFlow.first().selectedTrackId)
    }

    // ── Complete with no track → validation label ─────────────────────────────

    @Test
    fun `Complete with no track selected emits ValidationError`() = runTest {
        val store = buildStore()
        var label: OnboardingStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        store.accept(OnboardingStore.Intent.Complete)
        advanceUntilIdle()
        job.cancel()

        assertIs<OnboardingStore.Label.ValidationError>(label)
    }

    // ── Complete success → saves locally and emits Completed immediately ───────

    @Test
    fun `Complete saves activeTrackId locally and emits Completed`() = runTest {
        var savedTrackId: String? = null
        val store = buildStore(setActiveTrackId = { savedTrackId = it })

        var label: OnboardingStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        store.accept(OnboardingStore.Intent.SelectTrack("track-math"))
        store.accept(OnboardingStore.Intent.Complete)
        advanceUntilIdle()
        job.cancel()

        assertEquals("track-math", savedTrackId)
        assertIs<OnboardingStore.Label.Completed>(label)
    }

    // ── Complete with backend failure → still emits Completed (fire-and-forget) ─

    @Test
    fun `Complete emits Completed even when backend call fails`() = runTest {
        val store = buildStore(
            completeOnboarding = { Result.failure(RuntimeException("network error")) },
        )

        var label: OnboardingStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        store.accept(OnboardingStore.Intent.SelectTrack("track-math"))
        store.accept(OnboardingStore.Intent.Complete)
        advanceUntilIdle()
        job.cancel()

        assertIs<OnboardingStore.Label.Completed>(label)
        assertNull(store.stateFlow.first().error, "error should NOT be shown on backend failure")
    }
}
