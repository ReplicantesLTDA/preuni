package com.preuni.shared.presentation.welcome

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory

class WelcomeComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    val onCompleted: () -> Unit,
) : ComponentContext by componentContext {

    val store: WelcomeStore = WelcomeStoreFactory(storeFactory).create()
}
