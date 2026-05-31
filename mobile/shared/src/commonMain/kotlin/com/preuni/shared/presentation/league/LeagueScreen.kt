package com.preuni.shared.presentation.league

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.preuni.shared.ui.components.EmptyState

@Composable
fun LeagueScreen() {
    EmptyState(
        modifier = Modifier.fillMaxSize(),
        title = "Sua liga começa em breve",
        body = "Quando você ganhar XP suficiente, vai entrar em uma liga semanal e competir com outros estudantes no seu nível.",
    )
}
