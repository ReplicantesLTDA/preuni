package com.preuni.shared.presentation.onboarding

import com.arkivanov.decompose.ComponentContext
import com.arkivanov.mvikotlin.core.store.StoreFactory

class OnboardingComponent(
    componentContext: ComponentContext,
    private val storeFactory: StoreFactory,
    private val onCompleted: () -> Unit,
    private val setActiveTrackId: (String) -> Unit = {},
    private val completeOnboarding: suspend (trackIds: List<String>) -> Result<Unit> = { Result.success(Unit) },
) : ComponentContext by componentContext {

    val store: OnboardingStore = OnboardingStoreFactory(
        storeFactory = storeFactory,
        setActiveTrackId = setActiveTrackId,
        completeOnboarding = completeOnboarding,
    ).create()

    fun onComplete() = onCompleted()
}
