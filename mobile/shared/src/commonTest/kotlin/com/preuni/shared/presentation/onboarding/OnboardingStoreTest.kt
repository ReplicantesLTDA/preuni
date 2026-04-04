package com.preuni.shared.presentation.onboarding

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertIs
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class OnboardingStoreTest {

    private fun buildStore(
        completeResult: Result<Unit> = Result.success(Unit),
    ): OnboardingStore {
        return OnboardingStoreFactory(
            storeFactory = DefaultStoreFactory(),
            completeOnboarding = { _ -> completeResult },
        ).create()
    }

    // ── Page navigation ───────────────────────────────────────────────────────

    @Test
    fun `NextPage intent advances page index`() = runTest {
        val store = buildStore()
        assertEquals(0, store.stateFlow.first().pageIndex)
        store.accept(OnboardingStore.Intent.NextPage)
        assertEquals(1, store.stateFlow.first { it.pageIndex > 0 }.pageIndex)
    }

    @Test
    fun `PreviousPage intent decrements page index`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.NextPage)
        store.stateFlow.first { it.pageIndex == 1 }
        store.accept(OnboardingStore.Intent.PreviousPage)
        assertEquals(0, store.stateFlow.first { it.pageIndex == 0 }.pageIndex)
    }

    @Test
    fun `PreviousPage does not go below 0`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.PreviousPage)
        assertEquals(0, store.stateFlow.first().pageIndex)
    }

    @Test
    fun `NextPage does not exceed 3`() = runTest {
        val store = buildStore()
        repeat(10) { store.accept(OnboardingStore.Intent.NextPage) }
        assertEquals(3, store.stateFlow.first { it.pageIndex == 3 }.pageIndex)
    }

    // ── Track selection ───────────────────────────────────────────────────────

    @Test
    fun `ToggleTrack adds track to selectedTrackIds`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        assertTrue("track-math" in store.stateFlow.first { it.selectedTrackIds.isNotEmpty() }.selectedTrackIds)
    }

    @Test
    fun `ToggleTrack toggles track off when already selected`() = runTest {
        val store = buildStore()
        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        store.stateFlow.first { "track-math" in it.selectedTrackIds }
        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        assertTrue(store.stateFlow.first { "track-math" !in it.selectedTrackIds }.selectedTrackIds.isEmpty())
    }

    // ── Complete with 0 tracks → validation label ─────────────────────────────

    @Test
    fun `Complete with no tracks selected emits ValidationError label`() = runTest {
        val store = buildStore()
        val label = store.labels.first { true }
        // Navigate to last page
        repeat(3) { store.accept(OnboardingStore.Intent.NextPage) }
        store.accept(OnboardingStore.Intent.Complete)
        assertIs<OnboardingStore.Label.ValidationError>(label.also {})
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

        store.accept(OnboardingStore.Intent.ToggleTrack("track-math"))
        store.accept(OnboardingStore.Intent.ToggleTrack("track-port"))
        val label = store.labels.first { true }
        store.accept(OnboardingStore.Intent.Complete)

        assertIs<OnboardingStore.Label.Completed>(label.also {})
        assertFalse(calledWith.isNullOrEmpty())
        assertTrue("track-math" in calledWith!!)
    }
}
