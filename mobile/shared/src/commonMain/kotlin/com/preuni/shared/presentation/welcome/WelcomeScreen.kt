package com.preuni.shared.presentation.welcome

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
import androidx.compose.foundation.shape.CircleShape
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
import kotlinx.coroutines.launch

private data class WelcomePage(val emoji: String, val title: String, val body: String)

private val PAGES = listOf(
    WelcomePage(
        emoji = "🎓",
        title = "Bem-vindo ao PreUni!",
        body = "Prepare-se para o ENEM de forma inteligente e sem pressão. Aprenda no seu ritmo.",
    ),
    WelcomePage(
        emoji = "🔁",
        title = "Aprendizado por repetição",
        body = "Nosso método de repetição espaçada garante que você revise o conteúdo no momento certo para fixar melhor.",
    ),
    WelcomePage(
        emoji = "📝",
        title = "Simule o ENEM",
        body = "Treine com simulados completos e receba feedback detalhado sobre suas redações com inteligência artificial.",
    ),
)

@Composable
fun WelcomeScreen(
    store: WelcomeStore,
    onCompleted: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(WelcomeStore.State())
    val pagerState = rememberPagerState(pageCount = { WelcomeStoreFactory.TOTAL_PAGES })
    val scope = rememberCoroutineScope()

    LaunchedEffect(store) {
        store.labels.collect { label ->
            when (label) {
                WelcomeStore.Label.Completed -> onCompleted()
            }
        }
    }

    if (pagerState.currentPage != state.pageIndex) {
        scope.launch { pagerState.animateScrollToPage(state.pageIndex) }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 24.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        // Skip button top-right
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
            TextButton(onClick = { store.accept(WelcomeStore.Intent.Skip) }) {
                Text("Pular")
            }
        }

        HorizontalPager(
            state = pagerState,
            modifier = Modifier.weight(1f),
            userScrollEnabled = false,
        ) { page ->
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
        }

        // Page indicator dots
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.Center,
        ) {
            repeat(WelcomeStoreFactory.TOTAL_PAGES) { index ->
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
                if (index < WelcomeStoreFactory.TOTAL_PAGES - 1) Spacer(Modifier.width(6.dp))
            }
        }

        Spacer(Modifier.height(24.dp))

        PreuniButton(
            text = if (state.pageIndex < WelcomeStoreFactory.TOTAL_PAGES - 1) "Próximo" else "Começar",
            onClick = { store.accept(WelcomeStore.Intent.NextPage) },
            modifier = Modifier.fillMaxWidth(),
        )
    }
}
