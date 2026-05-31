package com.preuni.shared.presentation.friends

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.preuni.shared.ui.components.EmptyState

@Composable
fun FriendsScreen() {
    EmptyState(
        modifier = Modifier.fillMaxSize(),
        title = "Amigos chegando em breve",
        body = "Aqui você vai ver seus colegas de jornada, streaks e poderá convidar mais gente para estudar com você.",
    )
}
