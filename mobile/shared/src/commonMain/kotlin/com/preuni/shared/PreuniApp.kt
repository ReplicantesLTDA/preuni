package com.preuni.shared

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import com.arkivanov.decompose.extensions.compose.subscribeAsState
import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.preuni.shared.presentation.RootComponent
import com.preuni.shared.ui.theme.PreuniTheme
import com.preuni.shared.presentation.auth.AuthComponent
import com.preuni.shared.presentation.auth.LoginScreen
import com.preuni.shared.presentation.auth.LoginStore
import com.preuni.shared.presentation.auth.OtpLoginScreen
import com.preuni.shared.presentation.auth.RegisterScreen
import com.preuni.shared.presentation.auth.RegisterStore
import com.preuni.shared.presentation.auth.VerifyEmailScreen
import com.preuni.shared.presentation.home.HomeScreen
import com.preuni.shared.presentation.learn.LearnScreen
import com.preuni.shared.presentation.main.MainComponent
import com.preuni.shared.presentation.navigation.BottomNavigation
import com.preuni.shared.presentation.navigation.BottomTab
import com.preuni.shared.presentation.onboarding.OnboardingScreen
import com.preuni.shared.presentation.welcome.WelcomeScreen
import com.preuni.shared.presentation.profile.ChangeEmailScreen
import com.preuni.shared.presentation.profile.ChangeTrackScreen
import com.preuni.shared.presentation.profile.ConfirmNewEmailScreen
import com.preuni.shared.presentation.profile.DeleteAccountScreen
import com.preuni.shared.presentation.profile.EditPasswordScreen
import com.preuni.shared.presentation.profile.EditUsernameScreen
import com.preuni.shared.presentation.profile.ProfileComponent
import com.preuni.shared.presentation.profile.ProfileScreen
import com.preuni.shared.presentation.simulate.SimulateScreen

/**
 * Root composable entry point for all platforms.
 * Wires the Decompose component tree into Compose UI.
 */
@Composable
fun PreuniApp(component: RootComponent) {
    val childStack by component.childStack.subscribeAsState()

    PreuniTheme {
        when (val child = childStack.active.instance) {
            is RootComponent.Child.Welcome -> WelcomeScreen(
                store = child.component.store,
                onCompleted = child.component.onCompleted,
            )
            is RootComponent.Child.Auth -> AuthContent(child.component)
            is RootComponent.Child.Onboarding -> OnboardingScreen(
                store = child.component.store,
                onCompleted = child.component::onComplete,
            )
            is RootComponent.Child.Main -> MainContent(child.component)
        }
    }
}

@Composable
private fun AuthContent(component: AuthComponent) {
    val childStack by component.childStack.subscribeAsState()

    when (val child = childStack.active.instance) {
        is AuthComponent.Child.Login -> {
            LaunchedEffect(child.store) {
                child.store.labels.collect { label ->
                    when (label) {
                        is LoginStore.Label.LoggedIn -> component.onLoginSuccess()
                    }
                }
            }
            LoginScreen(
                store = child.store,
                onSignInWithCode = component::navigateToOtpLogin,
                onCreateAccount = component::navigateToRegister,
            )
        }
        is AuthComponent.Child.Register -> {
            LaunchedEffect(child.component.store) {
                child.component.store.labels.collect { label ->
                    when (label) {
                        RegisterStore.Label.Registered ->
                            component.navigateToVerifyEmail(child.component.store.state.email)
                    }
                }
            }
            RegisterScreen(
                store = child.component.store,
                onBack = child.component.onBack,
            )
        }
        is AuthComponent.Child.VerifyEmail -> VerifyEmailScreen(
            store = child.component.store,
            email = child.component.email,
            onVerified = child.component.onVerified,
        )
        is AuthComponent.Child.OtpLogin -> OtpLoginScreen(
            store = child.component.store,
            onBack = component::navigateBack,
        )
    }
}

@Composable
private fun MainContent(component: MainComponent) {
    val selectedTab by component.selectedTab.collectAsState()

    Scaffold(
        bottomBar = {
            BottomNavigation(
                selectedTab = selectedTab,
                onTabSelected = component::selectTab,
            )
        }
    ) { paddingValues ->
        Box(modifier = Modifier.fillMaxSize().padding(paddingValues)) {
            when (selectedTab) {
                BottomTab.HOME -> HomeScreen(
                    store = component.homeStore,
                    onStartLearning = { component.selectTab(BottomTab.LEARN) },
                )
                BottomTab.LEARN -> LearnScreen(
                    store = component.learnStore,
                    onOpenLesson = { /* PlaceholderLessonScreen not yet in nav */ },
                )
                BottomTab.SIMULATE -> SimulateScreen()
                BottomTab.PROFILE -> ProfileContent(component.profileComponent)
            }
        }
    }
}

@Composable
private fun ProfileContent(component: ProfileComponent) {
    val childStack by component.childStack.subscribeAsState()

    when (val child = childStack.active.instance) {
        is ProfileComponent.Child.Profile -> ProfileScreen(
            store = child.store,
            onEditUsername = component::navigateToEditUsername,
            onEditPassword = component::navigateToEditPassword,
            onChangeEmail = component::navigateToChangeEmail,
            onChangeTrack = { component.navigateToChangeTrack(null) },
            onDeleteAccount = component::navigateToDeleteAccount,
        )
        is ProfileComponent.Child.ChangeTrack -> ChangeTrackScreen(
            store = child.store,
            contentRepository = child.contentRepository,
            onBack = component::navigateBack,
            onSaved = {
                component.navigateBack()
                component.onSwitchToLearn()
            },
        )
        ProfileComponent.Child.EditUsername -> EditUsernameScreen(
            currentUsername = "",
            onSave = { component.navigateBack() },
            onBack = component::navigateBack,
        )
        ProfileComponent.Child.EditPassword -> EditPasswordScreen(
            onSave = { _, _ -> component.navigateBack() },
            onBack = component::navigateBack,
        )
        ProfileComponent.Child.ChangeEmail -> ChangeEmailScreen(
            onSubmit = { component.navigateToConfirmNewEmail() },
        )
        ProfileComponent.Child.ConfirmNewEmail -> ConfirmNewEmailScreen(
            newEmail = "",
            onSubmit = { component.navigateBack() },
            onResend = {},
        )
        ProfileComponent.Child.DeleteAccount -> DeleteAccountScreen(
            onConfirm = {},
            onBack = component::navigateBack,
        )
    }
}
