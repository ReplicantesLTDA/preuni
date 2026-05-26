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
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.presentation.RootComponent
import com.preuni.shared.ui.theme.PreuniTheme
import com.preuni.shared.presentation.auth.AuthComponent
import com.preuni.shared.presentation.auth.LoginScreen
import com.preuni.shared.presentation.auth.LoginStore
import com.preuni.shared.presentation.auth.OtpLoginScreen
import com.preuni.shared.presentation.auth.RegisterScreen
import com.preuni.shared.presentation.auth.RegisterStore
import com.preuni.shared.presentation.auth.VerifyEmailScreen
import com.preuni.shared.presentation.friends.FriendsScreen
import com.preuni.shared.presentation.learn.LearnScreen
import com.preuni.shared.presentation.league.LeagueScreen
import com.preuni.shared.presentation.main.MainComponent
import com.preuni.shared.presentation.navigation.BottomNavigation
import com.preuni.shared.presentation.navigation.BottomTab
import com.preuni.shared.presentation.settings.SettingsScreen
import com.preuni.shared.ui.components.StatusMetric
import com.preuni.shared.ui.components.TopStatusBar
import androidx.compose.foundation.layout.Column
import com.preuni.shared.presentation.onboarding.OnboardingScreen
import com.preuni.shared.presentation.welcome.WelcomeScreen
import com.preuni.shared.presentation.profile.ChangeEmailScreen
import com.preuni.shared.presentation.profile.ChangeEmailStore
import com.preuni.shared.presentation.profile.ChangePasswordStore
import com.preuni.shared.presentation.profile.ChangeTrackScreen
import com.preuni.shared.presentation.profile.ConfirmNewEmailScreen
import com.preuni.shared.presentation.profile.DeleteAccountScreen
import com.preuni.shared.presentation.profile.EditPasswordScreen
import com.preuni.shared.presentation.profile.EditUsernameScreen
import com.preuni.shared.presentation.profile.ProfileComponent
import com.preuni.shared.presentation.profile.ProfileScreen
import com.preuni.shared.presentation.profile.ProfileStore
import com.preuni.shared.presentation.simulate.SimulateScreen

/**
 * Placeholder status-bar metrics shown above the core tabs. Real values flow
 * in once the streak/XP feed is wired into MainComponent in a follow-up.
 */
private fun placeholderStatusMetrics(): List<StatusMetric> = listOf(
    StatusMetric(
        icon = "🔥",
        value = "—",
        caption = "Streak",
        accessibilityLabel = "Streak ainda não disponível",
        isPlaceholder = true,
    ),
    StatusMetric(
        icon = "⭐",
        value = "—",
        caption = "XP",
        accessibilityLabel = "XP ainda não disponível",
        isPlaceholder = true,
    ),
    StatusMetric(
        icon = "🦉",
        value = "—",
        caption = "Liga",
        accessibilityLabel = "Liga ainda não disponível",
        isPlaceholder = true,
    ),
)

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
        Column(modifier = Modifier.fillMaxSize().padding(paddingValues)) {
            if (selectedTab != BottomTab.PROFILE) {
                TopStatusBar(metrics = placeholderStatusMetrics())
            }
            Box(modifier = Modifier.fillMaxSize()) {
                when (selectedTab) {
                    BottomTab.LEARN -> LearnScreen(
                        store = component.learnStore,
                        onOpenLesson = { /* PlaceholderLessonScreen not yet in nav */ },
                    )
                    BottomTab.SIMULATE -> SimulateScreen()
                    BottomTab.FRIENDS -> FriendsScreen()
                    BottomTab.LEAGUE -> LeagueScreen()
                    BottomTab.PROFILE -> ProfileContent(component.profileComponent)
                }
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
            onLogout = component::logout,
            onSettings = component::navigateToSettings,
        )
        is ProfileComponent.Child.Settings -> SettingsScreen(
            onBack = component::navigateBack,
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
        is ProfileComponent.Child.EditUsername -> {
            val state by child.store.stateFlow.collectAsState()
            LaunchedEffect(child.store) {
                child.store.labels.collect { label ->
                    if (label is ProfileStore.Label.UsernameSaved) component.navigateBack()
                }
            }
            EditUsernameScreen(
                currentUsername = state.student?.username ?: "",
                onSave = { username -> child.store.accept(ProfileStore.Intent.UpdateUsername(username)) },
                onBack = component::navigateBack,
            )
        }
        is ProfileComponent.Child.EditPassword -> {
            val state by child.store.stateFlow.collectAsState()
            LaunchedEffect(child.store) {
                child.store.labels.collect { label ->
                    if (label is ChangePasswordStore.Label.Saved) component.navigateBack()
                }
            }
            EditPasswordScreen(
                onSave = { current, new -> child.store.accept(ChangePasswordStore.Intent.Submit(current, new)) },
                onBack = component::navigateBack,
                error = state.error,
            )
        }
        is ProfileComponent.Child.ChangeEmail -> {
            val state by child.store.stateFlow.collectAsState()
            LaunchedEffect(child.store) {
                child.store.labels.collect { label ->
                    if (label is ChangeEmailStore.Label.EmailChanged) component.navigateBack()
                }
            }
            if (state.phase == ChangeEmailStore.Phase.REQUEST) {
                ChangeEmailScreen(
                    onSubmit = { email -> child.store.accept(ChangeEmailStore.Intent.RequestCode(email)) },
                    onBack = component::navigateBack,
                    isLoading = state.isLoading,
                    error = state.error,
                )
            } else {
                ConfirmNewEmailScreen(
                    newEmail = state.newEmail,
                    onSubmit = { otp -> child.store.accept(ChangeEmailStore.Intent.ConfirmCode(otp)) },
                    onResend = { child.store.accept(ChangeEmailStore.Intent.ResendCode) },
                    onBack = component::navigateBack,
                    isLoading = state.isLoading,
                    error = state.error,
                )
            }
        }
        is ProfileComponent.Child.DeleteAccount -> {
            val state by child.store.stateFlow.collectAsState()
            LaunchedEffect(child.store) {
                child.store.labels.collect { label ->
                    if (label is ProfileStore.Label.AccountDeleted) component.logout()
                }
            }
            DeleteAccountScreen(
                onConfirm = { child.store.accept(ProfileStore.Intent.ConfirmDeleteAccount) },
                onBack = component::navigateBack,
                isLoading = state.isLoading,
            )
        }
    }
}
