package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.auth.AuthRepository
import com.preuni.shared.domain.auth.AuthSession

class OtpLoginComponent(
    componentContext: ComponentContext,
    storeFactory: StoreFactory,
    authRepository: AuthRepository,
    val onLoggedIn: () -> Unit,
) : ComponentContext by componentContext {

    val store: OtpLoginStore = OtpLoginStoreFactory(
        storeFactory = storeFactory,
        authRepository = authRepository,
        // TODO: wire to AuthApiClient.requestOtp once method is added to AuthRepository
        requestOtp = { _ -> Result.success(Unit) },
        verifyOtp = { _, _ -> Result.failure(UnsupportedOperationException("OTP login not wired")) },
    ).create()
}
