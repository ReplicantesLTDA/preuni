package com.preuni.shared.ui.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.preuni.shared.ui.theme.SubjectTrack

@Composable
fun SubjectTrackCard(
    track: SubjectTrack,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    selected: Boolean = false,
) {
    Card(
        onClick = onClick,
        modifier = modifier.fillMaxWidth(),
        shape = MaterialTheme.shapes.medium,
        colors = CardDefaults.cardColors(
            containerColor = track.colorScheme.container,
            contentColor = track.colorScheme.onContainer,
        ),
        border = if (selected) BorderStroke(2.dp, MaterialTheme.colorScheme.primary) else null,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 14.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = track.emoji,
                style = MaterialTheme.typography.headlineSmall,
            )
            Column {
                Text(
                    text = track.shortName,
                    style = MaterialTheme.typography.titleMedium,
                    color = track.colorScheme.onContainer,
                )
                Text(
                    text = track.name,
                    style = MaterialTheme.typography.bodySmall,
                    color = track.colorScheme.onContainer.copy(alpha = 0.75f),
                )
            }
        }
    }
}
