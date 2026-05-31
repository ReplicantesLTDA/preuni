package com.preuni.shared.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.preuni.shared.ui.theme.LocalSpacing

/**
 * Consistent heading for grouped lists/sections (e.g., profile rows, settings
 * groups). Optional trailing slot lets callers attach a counter, action, or
 * status indicator.
 */
@Composable
fun SectionHeader(
    title: String,
    modifier: Modifier = Modifier,
    trailing: (@Composable () -> Unit)? = null,
) {
    val s = LocalSpacing.current
    Row(
        modifier = modifier.padding(horizontal = s.lg, vertical = s.sm),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        if (trailing != null) trailing()
    }
}
