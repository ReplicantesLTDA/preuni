package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.auth.AuthRepository

class OtpLoginComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
    private val onLoggedIn: () -> Unit,
) : ComponentContext by componentContext
