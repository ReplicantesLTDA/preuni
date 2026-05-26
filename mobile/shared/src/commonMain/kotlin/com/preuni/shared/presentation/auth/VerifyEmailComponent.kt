package com.preuni.shared.presentation.auth

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory
import com.preuni.shared.domain.auth.AuthRepository

class VerifyEmailComponent(
    componentContext: ComponentContext,
    storeFactory: StoreFactory,
    private val authRepository: AuthRepository,
    val email: String = "",
    val onVerified: () -> Unit,
) : ComponentContext by componentContext {

    val store: VerifyEmailStore = VerifyEmailStoreFactory(
        storeFactory = storeFactory,
        email = email,
        onSubmit = { e, otp -> authRepository.verifyEmail(e, otp) },
        onResend = { e -> authRepository.resendVerificationOtp(e) },
    ).create()
}
