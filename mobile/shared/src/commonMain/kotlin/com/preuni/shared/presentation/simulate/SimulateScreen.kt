package com.preuni.shared.presentation.simulate

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
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.SectionHeader
import com.preuni.shared.ui.theme.LocalSpacing

@Composable
fun SimulateScreen(
    steps: List<WritingStep> = DefaultWritingFlow,
    onStartCurrentStep: () -> Unit = {},
) {
    val s = LocalSpacing.current
    val total = steps.size
    val doneCount = steps.count { it.status == WritingStep.Status.Done }
    val currentIndex = steps.indexOfFirst { it.status == WritingStep.Status.Current }
        .let { if (it >= 0) it else total }
    val progress = (doneCount.toFloat() / total).coerceIn(0f, 1f)
    val currentStep = steps.getOrNull(currentIndex)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .statusBarsPadding()
            .verticalScroll(rememberScrollState()),
    ) {
        Column(modifier = Modifier.padding(horizontal = s.xl, vertical = s.lg)) {
            Text(
                text = "Redação",
                style = MaterialTheme.typography.headlineMedium,
            )
            Spacer(Modifier.height(s.xs))
            Text(
                text = "Sua jornada de escrita acontece em cinco passos calmos. Vamos um por vez.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )

            Spacer(Modifier.height(s.lg))

            // Progress + stage indicator
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "Etapa ${(currentIndex + 1).coerceAtMost(total)} de $total",
                    style = MaterialTheme.typography.labelLarge,
                    color = MaterialTheme.colorScheme.primary,
                )
                Text(
                    text = "$doneCount concluídas",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            Spacer(Modifier.height(s.sm))
            LinearProgressIndicator(
                progress = { progress },
                modifier = Modifier.fillMaxWidth().height(8.dp),
                color = MaterialTheme.colorScheme.primary,
                trackColor = MaterialTheme.colorScheme.surfaceVariant,
            )
        }

        if (currentStep != null) {
            CurrentStepCard(
                step = currentStep,
                onStart = onStartCurrentStep,
            )
        }

        SectionHeader(title = "Todas as etapas")

        steps.forEachIndexed { index, step ->
            StepRow(index = index + 1, step = step)
        }

        Spacer(Modifier.height(s.xl))
    }
}

@Composable
private fun CurrentStepCard(
    step: WritingStep,
    onStart: () -> Unit,
) {
    val s = LocalSpacing.current
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = s.xl),
        color = MaterialTheme.colorScheme.primaryContainer,
        shape = MaterialTheme.shapes.medium,
    ) {
        Column(modifier = Modifier.padding(s.lg)) {
            Text(
                text = "Agora",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onPrimaryContainer,
                fontWeight = FontWeight.SemiBold,
            )
            Spacer(Modifier.height(s.xs))
            Text(
                text = step.title,
                style = MaterialTheme.typography.titleLarge,
                color = MaterialTheme.colorScheme.onPrimaryContainer,
            )
            Spacer(Modifier.height(s.sm))
            Text(
                text = step.summary,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onPrimaryContainer,
            )
            Spacer(Modifier.height(s.lg))
            PreuniButton(text = "Continuar", onClick = onStart)
        }
    }
}

@Composable
private fun StepRow(index: Int, step: WritingStep) {
    val s = LocalSpacing.current
    val colors = MaterialTheme.colorScheme
    val (badgeColor, textColor) = when (step.status) {
        WritingStep.Status.Done -> colors.primary to colors.onSurface
        WritingStep.Status.Current -> colors.tertiary to colors.onSurface
        WritingStep.Status.Locked -> colors.surfaceVariant to colors.onSurfaceVariant
    }
    val badgeLabel = when (step.status) {
        WritingStep.Status.Done -> "✓"
        WritingStep.Status.Current -> "$index"
        WritingStep.Status.Locked -> "🔒"
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = s.xl, vertical = s.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .clip(CircleShape)
                .background(badgeColor),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = badgeLabel,
                style = MaterialTheme.typography.labelLarge,
                color = colors.onPrimary,
                textAlign = TextAlign.Center,
            )
        }
        Spacer(Modifier.size(s.md))
        Column(modifier = Modifier.fillMaxWidth()) {
            Text(
                text = step.title,
                style = MaterialTheme.typography.titleSmall,
                color = textColor,
            )
            Text(
                text = step.summary,
                style = MaterialTheme.typography.bodySmall,
                color = colors.onSurfaceVariant,
            )
            if (step.status == WritingStep.Status.Locked) {
                Text(
                    text = "Disponível depois das etapas anteriores.",
                    style = MaterialTheme.typography.labelSmall,
                    color = colors.onSurfaceVariant,
                )
            }
        }
    }
}

