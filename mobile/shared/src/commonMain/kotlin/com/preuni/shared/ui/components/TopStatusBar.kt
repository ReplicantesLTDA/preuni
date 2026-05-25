package com.preuni.shared.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.Immutable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import com.preuni.shared.ui.theme.LocalSpacing

/**
 * A single compact metric displayed in the top status bar (e.g., streak, XP).
 *
 * @param icon emoji or short glyph rendered before [value]
 * @param value primary number / label (e.g. "12", "1 450")
 * @param caption secondary label (e.g. "dias", "XP")
 * @param accessibilityLabel screen-reader description ("12 dias de streak")
 * @param isPlaceholder true → render dimmed to signal "data not yet available"
 */
@Immutable
data class StatusMetric(
    val icon: String,
    val value: String,
    val caption: String,
    val accessibilityLabel: String,
    val isPlaceholder: Boolean = false,
)

/**
 * Compact status row shown at the top of the core destinations. Reinforces
 * continuity across sections by always being present in the same place.
 */
@Composable
fun TopStatusBar(
    metrics: List<StatusMetric>,
    modifier: Modifier = Modifier,
) {
    val s = LocalSpacing.current
    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surfaceVariant)
            .padding(horizontal = s.lg, vertical = s.sm),
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        metrics.forEach { metric ->
            MetricCell(metric = metric)
        }
    }
}

@Composable
private fun MetricCell(metric: StatusMetric) {
    val colors = MaterialTheme.colorScheme
    val foreground = if (metric.isPlaceholder) colors.onSurfaceVariant else colors.onSurface
    Column(
        modifier = Modifier
            .semantics { contentDescription = metric.accessibilityLabel },
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = "${metric.icon}  ${metric.value}",
            style = MaterialTheme.typography.titleMedium,
            color = foreground,
            textAlign = TextAlign.Center,
        )
        Text(
            text = metric.caption,
            style = MaterialTheme.typography.labelSmall,
            color = colors.onSurfaceVariant,
            textAlign = TextAlign.Center,
        )
    }
}
