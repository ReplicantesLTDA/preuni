package com.preuni.shared.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import com.preuni.shared.ui.theme.LocalSpacing

/**
 * Friendly placeholder shown on screens with no data yet ("coming soon",
 * "no friends added", etc.). Always shows title + body; primary action is
 * optional and rendered via [PreuniButton] when present.
 */
@Composable
fun EmptyState(
    title: String,
    body: String,
    modifier: Modifier = Modifier,
    primaryActionText: String? = null,
    onPrimaryAction: (() -> Unit)? = null,
) {
    val s = LocalSpacing.current
    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = s.xl, vertical = s.xxl),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        MascotPlaceholder()
        Spacer(Modifier.height(s.lg))
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            color = MaterialTheme.colorScheme.onSurface,
            textAlign = TextAlign.Center,
        )
        Spacer(Modifier.height(s.sm))
        Text(
            text = body,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center,
        )
        if (primaryActionText != null && onPrimaryAction != null) {
            Spacer(Modifier.height(s.lg))
            PreuniButton(
                text = primaryActionText,
                onClick = onPrimaryAction,
            )
        }
    }
}
