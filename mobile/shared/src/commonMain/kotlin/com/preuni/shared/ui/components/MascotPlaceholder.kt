package com.preuni.shared.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp

/**
 * Friendly intentional placeholder for the upcoming mascot artwork.
 * Renders a colored circle with a simple emoji glyph — stable layout so
 * swapping in the final illustration does not shift surrounding content.
 *
 * No external assets, no animation, no third-party artwork.
 */
@Composable
fun MascotPlaceholder(
    modifier: Modifier = Modifier,
    accessibilityLabel: String = "Mascote (em breve)",
) {
    Box(
        modifier = modifier
            .size(96.dp)
            .background(
                color = MaterialTheme.colorScheme.primaryContainer,
                shape = CircleShape,
            )
            .semantics { contentDescription = accessibilityLabel },
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = "🦉",
            style = MaterialTheme.typography.displaySmall,
        )
    }
}
