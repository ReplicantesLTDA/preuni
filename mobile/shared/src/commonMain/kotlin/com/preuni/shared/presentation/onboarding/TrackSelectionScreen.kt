package com.preuni.shared.presentation.onboarding

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.domain.content.ContentRepository
import com.preuni.shared.domain.content.Track

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun TrackSelectionScreen(
    store: OnboardingStore,
    contentRepository: ContentRepository,
) {
    val state by store.stateFlow.collectAsState(OnboardingStore.State())
    var tracks by remember { mutableStateOf<List<Track>>(emptyList()) }
    var loadError by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        contentRepository.getTracks().fold(
            onSuccess = { tracks = it },
            onFailure = { loadError = "Failed to load tracks. Please try again." },
        )
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 24.dp, vertical = 16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Top,
    ) {
        Text("Quais matérias você quer estudar?", style = MaterialTheme.typography.titleLarge)
        Spacer(Modifier.height(8.dp))
        Text(
            "Escolha ao menos uma. Você pode mudar depois.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(24.dp))

        if (loadError != null) {
            Text(loadError!!, color = MaterialTheme.colorScheme.error)
        } else {
            FlowRow(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.fillMaxWidth(),
            ) {
                tracks.forEach { track ->
                    FilterChip(
                        selected = track.id in state.selectedTrackIds,
                        onClick = { store.accept(OnboardingStore.Intent.ToggleTrack(track.id)) },
                        label = { Text(track.name) },
                    )
                }
            }
        }

        Spacer(Modifier.height(32.dp))

        Button(
            onClick = { store.accept(OnboardingStore.Intent.Complete) },
            enabled = state.selectedTrackIds.isNotEmpty() && !state.isLoading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Começar")
        }
    }
}
