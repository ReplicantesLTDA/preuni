package com.preuni.shared.presentation.settings

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.preuni.shared.ui.components.EmptyState
import com.preuni.shared.ui.components.SectionHeader
import com.preuni.shared.ui.theme.LocalSpacing

/**
 * Settings screen placeholder. Real preference groups (notifications,
 * accessibility, data privacy) land in follow-up features; for now we
 * keep the navigation surface live with friendly empty-state copy.
 */
@Composable
fun SettingsScreen(onBack: () -> Unit) {
    val s = LocalSpacing.current
    Column(
        modifier = Modifier
            .fillMaxSize()
            .statusBarsPadding(),
    ) {
        Column(modifier = Modifier.padding(horizontal = s.sm, vertical = s.sm)) {
            IconButton(onClick = onBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Voltar")
            }
        }
        Text(
            text = "Ajustes",
            style = MaterialTheme.typography.headlineMedium,
            modifier = Modifier.padding(horizontal = s.lg),
        )
        Spacer(Modifier.height(s.lg))
        SectionHeader(title = "Preferências")
        EmptyState(
            title = "Ajustes em construção",
            body = "Notificações, idioma e privacidade vão aparecer aqui em breve. Por enquanto, ajustes da conta ficam no seu perfil.",
        )
    }
}
