package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory

class VerifyEmailComponent(
    componentContext: ComponentContext,
    storeFactory: StoreFactory,
    val email: String = "",
    val onVerified: () -> Unit,
) : ComponentContext by componentContext {

    val store: VerifyEmailStore = VerifyEmailStoreFactory(
        storeFactory = storeFactory,
        email = email,
        // TODO: wire to AuthApiClient.verifyEmail once method is added to AuthRepository
        onSubmit = { _, _ -> Result.success(Unit) },
        onResend = { _ -> Result.success(Unit) },
    ).create()
}
