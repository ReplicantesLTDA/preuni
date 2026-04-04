package com.preuni.shared.presentation.onboarding

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
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import kotlinx.coroutines.launch

private val PAGES = listOf(
    OnboardingPage(
        title = "Bem-vindo ao PreUni!",
        body = "Prepare-se para o ENEM de forma inteligente e sem pressão. Aprenda no seu ritmo.",
    ),
    OnboardingPage(
        title = "Aprendizado por repetição",
        body = "Nosso método de repetição espaçada garante que você revise o conteúdo no momento certo para fixar melhor.",
    ),
    OnboardingPage(
        title = "Simule o ENEM",
        body = "Treine com simulados completos e receba feedback detalhado sobre suas redações com inteligência artificial.",
    ),
    // Page 3 = Track selection (rendered separately)
)

@Composable
fun OnboardingScreen(
    store: OnboardingStore,
    onCompleted: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(OnboardingStore.State())
    val pagerState = rememberPagerState(pageCount = { 4 })
    val scope = rememberCoroutineScope()

    // Sync pager to store pageIndex
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
                    Text(p.title, style = MaterialTheme.typography.headlineMedium)
                    Spacer(Modifier.height(16.dp))
                    Text(
                        p.body,
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            } else {
                // Track selection page handled in TrackSelectionScreen
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text(
                        "Escolha suas matérias",
                        style = MaterialTheme.typography.headlineMedium,
                    )
                }
            }
        }

        // Page indicator dots
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.Center,
        ) {
            repeat(4) { index ->
                val selected = state.pageIndex == index
                Surface(
                    modifier = Modifier.size(if (selected) 10.dp else 7.dp).padding(2.dp),
                    shape = CircleShape,
                    color = if (selected) MaterialTheme.colorScheme.primary
                            else MaterialTheme.colorScheme.outlineVariant,
                ) {}
            }
        }

        Spacer(Modifier.height(24.dp))

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            if (state.pageIndex > 0) {
                TextButton(onClick = { store.accept(OnboardingStore.Intent.PreviousPage) }) {
                    Text("Voltar")
                }
            } else {
                Spacer(Modifier.weight(1f))
            }

            if (state.pageIndex < 3) {
                Button(onClick = { store.accept(OnboardingStore.Intent.NextPage) }) {
                    Text("Próximo")
                }
            } else {
                Button(
                    onClick = {
                        store.accept(OnboardingStore.Intent.Complete)
                    },
                    enabled = state.selectedTrackIds.isNotEmpty() && !state.isLoading,
                ) {
                    Text("Começar")
                }
            }
        }
    }
}

private data class OnboardingPage(val title: String, val body: String)
