package com.preuni.shared.presentation.profile

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
import com.preuni.shared.domain.error.AppError
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
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
import kotlin.test.assertTrue

@OptIn(ExperimentalCoroutinesApi::class)
class ChangeTrackStoreTest {

    private val storeFactory = DefaultStoreFactory()
    private val initialTrackIds = setOf("math", "sciences")

    @BeforeTest
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterTest
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun load_populatesSelectedTrackIds() = runTest {
        val store = buildStore(initialTrackIds = initialTrackIds) { Result.success(Unit) }
        store.accept(ChangeTrackStore.Intent.Load(initialTrackIds))
        advanceUntilIdle()
        assertEquals(initialTrackIds, store.state.selectedTrackIds)
    }

    @Test
    fun save_validSelection_emitsSavedLabel() = runTest {
        val store = buildStore(initialTrackIds = setOf("math")) { Result.success(Unit) }
        store.accept(ChangeTrackStore.Intent.Load(setOf("math")))
        store.accept(ChangeTrackStore.Intent.ToggleTrack("sciences"))

        var receivedLabel: ChangeTrackStore.Label? = null
        val job = launch(Dispatchers.Main) {
            store.labels.collect { receivedLabel = it }
        }
        store.accept(ChangeTrackStore.Intent.Save)
        advanceUntilIdle()
        job.cancel()

        assertIs<ChangeTrackStore.Label.Saved>(receivedLabel)
        assertFalse(store.state.isLoading)
    }

    @Test
    fun save_emptySelection_emitsValidationError_withoutCallingRepo() = runTest {
        var updateCalled = false
        val store = buildStore(initialTrackIds = setOf("math")) {
            updateCalled = true
            Result.success(Unit)
        }
        // Load then deselect all
        store.accept(ChangeTrackStore.Intent.Load(setOf("math")))
        store.accept(ChangeTrackStore.Intent.ToggleTrack("math"))

        var receivedLabel: ChangeTrackStore.Label? = null
        val job = launch(Dispatchers.Main) {
            store.labels.collect { receivedLabel = it }
        }
        store.accept(ChangeTrackStore.Intent.Save)
        advanceUntilIdle()
        job.cancel()

        assertFalse(updateCalled)
        assertIs<ChangeTrackStore.Label.ValidationError>(receivedLabel)
        assertNull(store.state.error)
        assertTrue(store.state.selectedTrackIds.isEmpty())
    }

    @Test
    fun save_repositoryError_setsError() = runTest {
        val store = buildStore(initialTrackIds = setOf("math")) {
            Result.failure(AppError.Unknown())
        }
        store.accept(ChangeTrackStore.Intent.Load(setOf("math")))
        store.accept(ChangeTrackStore.Intent.Save)
        advanceUntilIdle()

        assertNotNull(store.state.error)
        assertIs<AppError.Unknown>(store.state.error)
        assertFalse(store.state.isLoading)
    }

    private fun buildStore(
        initialTrackIds: Set<String> = emptySet(),
        updateTracks: suspend (List<String>) -> Result<Unit>,
    ): ChangeTrackStore =
        ChangeTrackStoreFactory(storeFactory, updateTracks).create()
}
