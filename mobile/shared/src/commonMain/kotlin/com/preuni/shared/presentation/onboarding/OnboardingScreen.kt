package com.preuni.shared.presentation.onboarding

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.SubjectTrackCard
import com.preuni.shared.ui.theme.SubjectTracks

@Composable
fun OnboardingScreen(
    store: OnboardingStore,
    onCompleted: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(OnboardingStore.State())
    val pagerState = rememberPagerState(pageCount = { 1 })

    LaunchedEffect(store) {
        store.labels.collect { label ->
            when (label) {
                is OnboardingStore.Label.Completed -> onCompleted()
                is OnboardingStore.Label.ValidationError -> Unit
            }
        }
    }

    Column(
        modifier = Modifier.fillMaxSize().padding(horizontal = 24.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.weight(1f),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState()),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Text(
                    "Escolha suas matérias",
                    style = MaterialTheme.typography.headlineMedium,
                    textAlign = TextAlign.Center,
                )
                Spacer(Modifier.height(8.dp))
                Text(
                    "Você pode escolher mais de uma.",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                Spacer(Modifier.height(16.dp))
                SubjectTracks.forEach { track ->
                    SubjectTrackCard(
                        track = track,
                        selected = state.selectedTrackIds.contains(track.id),
                        onClick = { store.accept(OnboardingStore.Intent.ToggleTrack(track.id)) },
                        modifier = Modifier.fillMaxWidth(),
                    )
                    Spacer(Modifier.height(8.dp))
                }
            }
        }

        Spacer(Modifier.height(24.dp))

        PreuniButton(
            text = "Começar",
            onClick = { store.accept(OnboardingStore.Intent.Complete) },
            enabled = state.selectedTrackIds.isNotEmpty() && !state.isLoading,
            isLoading = state.isLoading,
            modifier = Modifier.width(160.dp),
        )
    }
}
