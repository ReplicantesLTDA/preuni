package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.auth.AuthRepository

class RegisterComponent(
    componentContext: ComponentContext,
    storeFactory: StoreFactory,
    authRepository: AuthRepository,
    val onBack: () -> Unit,
) : ComponentContext by componentContext {

    val store: RegisterStore = RegisterStoreFactory(storeFactory, authRepository).create()
}
