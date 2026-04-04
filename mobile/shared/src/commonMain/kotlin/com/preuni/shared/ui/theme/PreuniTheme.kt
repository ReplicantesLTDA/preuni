package com.preuni.shared.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable

@Composable
fun PreuniTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = preuniLightColorScheme,
        shapes = preuniShapes,
        content = content,
    )
}
