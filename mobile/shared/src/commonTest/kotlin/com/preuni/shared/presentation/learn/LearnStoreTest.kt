package com.preuni.shared.presentation.learn

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.content.Track
import com.preuni.shared.domain.learn.NodeState
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
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class LearnStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private val fakeTrack = Track(
        id = "matematica",
        name = "Matemática",
        slug = "matematica",
        colorToken = "blue",
        lessonCount = 15,
    )

    private fun buildStore(
        activeTrackId: String? = "matematica",
        tracks: List<Track> = listOf(fakeTrack),
    ): LearnStore =
        LearnStoreFactory(
            storeFactory = DefaultStoreFactory(),
            getActiveTrackId = { activeTrackId },
            getTracks = { Result.success(tracks) },
        ).create()

    // ── Load ─────────────────────────────────────────────────────────────────

    @Test
    fun `Load sets activeTrack and generates stub modules`() = runTest {
        val store = buildStore()
        store.accept(LearnStore.Intent.Load)
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertEquals(fakeTrack, state.activeTrack)
        assertTrue(state.modules.isNotEmpty())
        assertEquals(NodeState.ACTIVE, state.modules.first().state)
        assertTrue(state.modules.drop(1).all { it.state == NodeState.LOCKED })
    }

    @Test
    fun `Load with no activeTrackId leaves activeTrack null`() = runTest {
        val store = buildStore(activeTrackId = null)
        store.accept(LearnStore.Intent.Load)
        advanceUntilIdle()

        val state = store.stateFlow.first()
        assertNull(state.activeTrack)
        assertTrue(state.modules.isEmpty())
    }

    // ── TapNode ───────────────────────────────────────────────────────────────

    @Test
    fun `TapNode on ACTIVE node emits OpenLesson label`() = runTest {
        val store = buildStore()
        store.accept(LearnStore.Intent.Load)
        advanceUntilIdle()

        var label: LearnStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        val activeModule = store.stateFlow.first().modules.first { it.state == NodeState.ACTIVE }
        store.accept(LearnStore.Intent.TapNode(activeModule.id))
        advanceUntilIdle()
        job.cancel()

        assertIs<LearnStore.Label.OpenLesson>(label)
        assertEquals(activeModule.id, (label as LearnStore.Label.OpenLesson).moduleId)
    }

    @Test
    fun `TapNode on LOCKED node emits ShowLockedMessage label`() = runTest {
        val store = buildStore()
        store.accept(LearnStore.Intent.Load)
        advanceUntilIdle()

        var label: LearnStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        val lockedModule = store.stateFlow.first().modules.first { it.state == NodeState.LOCKED }
        store.accept(LearnStore.Intent.TapNode(lockedModule.id))
        advanceUntilIdle()
        job.cancel()

        assertIs<LearnStore.Label.ShowLockedMessage>(label)
    }
}
