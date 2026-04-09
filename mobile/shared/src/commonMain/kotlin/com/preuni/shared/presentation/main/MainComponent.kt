package com.preuni.shared.presentation.main

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.data.auth.TokenStore
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.content.ContentRepository
import com.preuni.shared.domain.user.UserRepository
import com.preuni.shared.presentation.home.HomeStoreFactory
import com.preuni.shared.presentation.learn.LearnStoreFactory
import com.preuni.shared.presentation.navigation.BottomTab
import com.preuni.shared.presentation.profile.ProfileComponent
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

class MainComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
    private val tokenStore: TokenStore,
    private val userRepository: UserRepository,
    private val contentRepository: ContentRepository,
    private val onLogout: () -> Unit,
) : ComponentContext by componentContext {

    private val _selectedTab = MutableStateFlow(BottomTab.LEARN)
    val selectedTab: StateFlow<BottomTab> = _selectedTab

    val homeStore = HomeStoreFactory(storeFactory, userRepository).create()

    val learnStore = LearnStoreFactory(
        storeFactory = storeFactory,
        getActiveTrackId = { tokenStore.getActiveTrackId() },
        getTracks = { contentRepository.getTracks() },
    ).create()

    val profileComponent = ProfileComponent(
        componentContext = componentContext,
        storeFactory = storeFactory,
        userRepository = userRepository,
        contentRepository = contentRepository,
        onLogout = onLogout,
        onSwitchToLearn = { _selectedTab.value = BottomTab.LEARN },
        setActiveTrackId = { tokenStore.setActiveTrackId(it) },
    )

    fun selectTab(tab: BottomTab) {
        _selectedTab.value = tab
    }

    fun logout() {
        onLogout()
    }
}
