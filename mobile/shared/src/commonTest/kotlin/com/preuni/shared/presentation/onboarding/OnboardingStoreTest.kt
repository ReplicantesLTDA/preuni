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
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class OnboardingStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private fun buildStore(
        completeResult: Result<Unit> = Result.success(Unit),
    ): OnboardingStore {
        return OnboardingStoreFactory(
            storeFactory = DefaultStoreFactory(),
            completeOnboarding = { _ -> completeResult },
        ).create()
    }

    // ── Track selection ───────────────────────────────────────────────────────

    @Test
    fun `ToggleTrack adds track to selectedTrackIds`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        assertTrue("track-math" in store.stateFlow.first().selectedTrackIds)
    }

    @Test
    fun `ToggleTrack toggles track off when already selected`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        assertTrue(store.stateFlow.first().selectedTrackIds.isEmpty())
    }

    // ── Complete with 0 tracks → validation label ─────────────────────────────

    @Test
    fun `Complete with no tracks selected emits ValidationError label`() = runTest {
        val store = buildStore()
        var receivedLabel: OnboardingStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { receivedLabel = it } }

        store.accept(OnboardingStore.Intent.Complete)
        advanceUntilIdle()
        job.cancel()

        assertIs<OnboardingStore.Label.ValidationError>(receivedLabel)
    }

    // ── Complete with tracks → Completed label ────────────────────────────────

    @Test
    fun `Complete with tracks selected calls completeOnboarding and emits Completed label`() = runTest {
        var calledWith: List<String>? = null
        val store = OnboardingStoreFactory(
            storeFactory = DefaultStoreFactory(),
            completeOnboarding = { ids ->
                calledWith = ids
                Result.success(Unit)
            },
        ).create()

        var receivedLabel: OnboardingStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { receivedLabel = it } }

        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        store.accept(OnboardingStore.Intent.ToggleTrack("track-port"))
        store.accept(OnboardingStore.Intent.Complete)
        advanceUntilIdle()
        job.cancel()

        assertIs<OnboardingStore.Label.Completed>(receivedLabel)
        assertFalse(calledWith.isNullOrEmpty())
        assertTrue("track-math" in calledWith!!)
    }
}
