package com.preuni.shared.presentation.onboarding

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.SubjectTrackCard
import com.preuni.shared.ui.theme.SubjectTracks
import kotlinx.coroutines.launch

private val PAGES = listOf(
    OnboardingPage(
        emoji = "🎓",
        title = "Bem-vindo ao PreUni!",
        body = "Prepare-se para o ENEM de forma inteligente e sem pressão. Aprenda no seu ritmo.",
    ),
    OnboardingPage(
        emoji = "🔁",
        title = "Aprendizado por repetição",
        body = "Nosso método de repetição espaçada garante que você revise o conteúdo no momento certo para fixar melhor.",
    ),
    OnboardingPage(
        emoji = "📝",
        title = "Simule o ENEM",
        body = "Treine com simulados completos e receba feedback detalhado sobre suas redações com inteligência artificial.",
    ),
)

@Composable
fun OnboardingScreen(
    store: OnboardingStore,
    onCompleted: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(OnboardingStore.State())
    val pagerState = rememberPagerState(pageCount = { 4 })
    val scope = rememberCoroutineScope()

    LaunchedEffect(store) {
        store.labels.collect { label ->
            when (label) {
                is OnboardingStore.Label.Completed -> onCompleted()
                is OnboardingStore.Label.ValidationError -> Unit
            }
        }
    }

    if (pagerState.currentPage != state.pageIndex) {
        scope.launch { pagerState.animateScrollToPage(state.pageIndex) }
    }

    Column(
        modifier = Modifier.fillMaxSize().padding(horizontal = 24.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.weight(1f),
        ) { page ->
            if (page < 3) {
                val p = PAGES[page]
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Text(
                        p.emoji,
                        style = MaterialTheme.typography.displayLarge,
                        modifier = Modifier.align(Alignment.CenterHorizontally),
                    )
                    Spacer(Modifier.height(24.dp))
                    Text(
                        p.title,
                        style = MaterialTheme.typography.headlineMedium,
                        textAlign = TextAlign.Center,
                    )
                    Spacer(Modifier.height(16.dp))
                    Text(
                        p.body,
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        textAlign = TextAlign.Center,
                    )
                }
            } else {
                // Track selection page
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
                        )
                        Spacer(Modifier.height(8.dp))
                    }
                }
            }
        }

        // Page indicator dots (pill-shaped active dot)
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.Center,
        ) {
            repeat(4) { index ->
                val selected = state.pageIndex == index
                if (selected) {
                    Box(
                        Modifier
                            .width(24.dp)
                            .height(8.dp)
                            .clip(MaterialTheme.shapes.extraLarge)
                            .background(MaterialTheme.colorScheme.primary),
                    )
                } else {
                    Box(
                        Modifier
                            .size(8.dp)
                            .clip(CircleShape)
                            .background(MaterialTheme.colorScheme.outlineVariant),
                    )
                }
                if (index < 3) Spacer(Modifier.width(6.dp))
            }
        }

        Spacer(Modifier.height(24.dp))

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            if (state.pageIndex > 0) {
                TextButton(onClick = { store.accept(OnboardingStore.Intent.PreviousPage) }) {
                    Text("Voltar")
                }
            } else {
                Spacer(Modifier.weight(1f))
            }

            if (state.pageIndex < 3) {
                PreuniButton(
                    text = "Próximo",
                    onClick = { store.accept(OnboardingStore.Intent.NextPage) },
                    modifier = Modifier.width(160.dp),
                )
            } else {
                PreuniButton(
                    text = "Começar",
                    onClick = { store.accept(OnboardingStore.Intent.Complete) },
                    enabled = state.selectedTrackIds.isNotEmpty() && !state.isLoading,
                    isLoading = state.isLoading,
                    modifier = Modifier.width(160.dp),
                )
            }
        }
    }
}

private data class OnboardingPage(val emoji: String, val title: String, val body: String)
