package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
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
import kotlin.test.assertNotNull
import kotlin.test.assertNull

@OptIn(ExperimentalCoroutinesApi::class)
class ChangeTrackStoreTest {

    @BeforeTest
    fun setUp() { Dispatchers.setMain(UnconfinedTestDispatcher()) }

    @AfterTest
    fun tearDown() { Dispatchers.resetMain() }

    private fun buildStore(
        setActiveTrackId: (String) -> Unit = {},
        updateTracks: suspend (List<String>) -> Result<Unit> = { Result.success(Unit) },
    ): ChangeTrackStore =
        ChangeTrackStoreFactory(
            storeFactory = DefaultStoreFactory(),
            setActiveTrackId = setActiveTrackId,
            updateTracks = updateTracks,
        ).create()

    // ── Load ─────────────────────────────────────────────────────────────────

    @Test
    fun `Load with initialTrackId sets selectedTrackId`() = runTest {
        val store = buildStore()
        store.accept(ChangeTrackStore.Intent.Load(initialTrackId = "matematica"))
        assertEquals("matematica", store.stateFlow.first().selectedTrackId)
    }

    // ── Select ────────────────────────────────────────────────────────────────

    @Test
    fun `SelectTrack replaces previous selection`() = runTest {
        val store = buildStore()
        store.accept(ChangeTrackStore.Intent.Load(initialTrackId = "matematica"))
        store.accept(ChangeTrackStore.Intent.SelectTrack("linguagens"))
        assertEquals("linguagens", store.stateFlow.first().selectedTrackId)
    }

    // ── Save success ──────────────────────────────────────────────────────────

    @Test
    fun `Save with valid selection saves locally and emits Saved`() = runTest {
        var savedId: String? = null
        val store = buildStore(setActiveTrackId = { savedId = it })
        store.accept(ChangeTrackStore.Intent.Load(initialTrackId = "matematica"))

        var label: ChangeTrackStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        store.accept(ChangeTrackStore.Intent.Save)
        advanceUntilIdle()
        job.cancel()

        assertEquals("matematica", savedId)
        assertIs<ChangeTrackStore.Label.Saved>(label)
        assertFalse(store.stateFlow.first().isLoading)
    }

    // ── Save with no selection → validation ───────────────────────────────────

    @Test
    fun `Save with no selection emits ValidationError without calling repo`() = runTest {
        var updateCalled = false
        val store = buildStore(updateTracks = { updateCalled = true; Result.success(Unit) })

        var label: ChangeTrackStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        store.accept(ChangeTrackStore.Intent.Save)
        advanceUntilIdle()
        job.cancel()

        assertFalse(updateCalled)
        assertIs<ChangeTrackStore.Label.ValidationError>(label)
        assertNull(store.stateFlow.first().error)
    }

    // ── Save with backend error → still emits Saved (fire-and-forget) ─────────

    @Test
    fun `Save emits Saved even when backend call fails`() = runTest {
        val store = buildStore(updateTracks = { Result.failure(AppError.Unknown()) })
        store.accept(ChangeTrackStore.Intent.Load(initialTrackId = "matematica"))

        var label: ChangeTrackStore.Label? = null
        val job = launch(Dispatchers.Main) { store.labels.collect { label = it } }

        store.accept(ChangeTrackStore.Intent.Save)
        advanceUntilIdle()
        job.cancel()

        assertIs<ChangeTrackStore.Label.Saved>(label)
        assertNull(store.stateFlow.first().error, "backend errors should not surface in UI")
    }
}
