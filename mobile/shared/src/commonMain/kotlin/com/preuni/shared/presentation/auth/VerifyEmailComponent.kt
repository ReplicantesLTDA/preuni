package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory

class VerifyEmailComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val onVerified: () -> Unit,
) : ComponentContext by componentContext
