package com.preuni.shared.presentation.welcome

import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.main.store.DefaultStoreFactory
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
import kotlin.test.assertIs

@OptIn(ExperimentalCoroutinesApi::class)
class WelcomeStoreTest {

    private val storeFactory = DefaultStoreFactory()

    @BeforeTest
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @AfterTest
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun nextPage_advancesPageIndex() = runTest {
        val store = WelcomeStoreFactory(storeFactory).create()
        store.accept(WelcomeStore.Intent.NextPage)
        advanceUntilIdle()
        assertEquals(1, store.state.pageIndex)
        store.accept(WelcomeStore.Intent.NextPage)
        advanceUntilIdle()
        assertEquals(2, store.state.pageIndex)
    }

    @Test
    fun nextPage_onLastPage_emitsCompleted() = runTest {
        val store = WelcomeStoreFactory(storeFactory).create()

        var receivedLabel: WelcomeStore.Label? = null
        val job = launch(Dispatchers.Main) {
            store.labels.collect { receivedLabel = it }
        }

        // Advance from page 0 → 1 → 2 → complete
        store.accept(WelcomeStore.Intent.NextPage)
        store.accept(WelcomeStore.Intent.NextPage)
        store.accept(WelcomeStore.Intent.NextPage)
        advanceUntilIdle()
        job.cancel()

        assertIs<WelcomeStore.Label.Completed>(receivedLabel)
    }

    @Test
    fun skip_emitsCompletedImmediately() = runTest {
        val store = WelcomeStoreFactory(storeFactory).create()

        var receivedLabel: WelcomeStore.Label? = null
        val job = launch(Dispatchers.Main) {
            store.labels.collect { receivedLabel = it }
        }

        store.accept(WelcomeStore.Intent.Skip)
        advanceUntilIdle()
        job.cancel()

        assertIs<WelcomeStore.Label.Completed>(receivedLabel)
    }
}
