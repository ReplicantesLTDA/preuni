package com.preuni.shared.presentation.learn

import androidx.compose.foundation.clickable
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.labels
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.domain.learn.ModuleNode
import com.preuni.shared.domain.learn.NodeState
import com.preuni.shared.ui.theme.SubjectTracks
import kotlinx.coroutines.launch

private val NODE_SIZE = 56.dp
private val NODE_SPACING = 100.dp

@Composable
fun LearnScreen(
    store: LearnStore,
    onOpenLesson: (moduleId: String) -> Unit,
) {
    val state by store.stateFlow.collectAsState(LearnStore.State())
    val snackbarHostState = remember { SnackbarHostState() }
    val scope = rememberCoroutineScope()

    LaunchedEffect(store) {
        store.labels.collect { label ->
            when (label) {
                is LearnStore.Label.OpenLesson -> onOpenLesson(label.moduleId)
                LearnStore.Label.ShowLockedMessage -> {
                    scope.launch {
                        snackbarHostState.showSnackbar("Complete os módulos anteriores primeiro")
                    }
                }
            }
        }
    }

    LaunchedEffect(Unit) {
        store.accept(LearnStore.Intent.Load)
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
    ) { innerPadding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding),
        ) {
            when {
                state.isLoading -> LoadingContent()
                state.modules.isEmpty() -> NoTrackContent()
                else -> TrackPathContent(
                    state = state,
                    onTapNode = { store.accept(LearnStore.Intent.TapNode(it)) },
                )
            }
        }
    }
}

@Composable
private fun TrackPathContent(
    state: LearnStore.State,
    onTapNode: (String) -> Unit,
) {
    val localTrack = SubjectTracks.find { it.id == (state.activeTrackId ?: state.activeTrack?.id) }
    val trackColor = localTrack?.colorScheme?.onContainer ?: MaterialTheme.colorScheme.primary

    val listState = rememberLazyListState()

    // Find the active module index and scroll to it on first load
    LaunchedEffect(state.modules) {
        val activeIndex = state.modules.indexOfFirst { it.state == NodeState.ACTIVE }
        if (activeIndex > 0) listState.animateScrollToItem(activeIndex)
    }

    BoxWithConstraints(modifier = Modifier.fillMaxSize()) {
        val screenWidth = maxWidth

        LazyColumn(
            state = listState,
            modifier = Modifier.fillMaxSize(),
        ) {
            // Subject header
            item {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 24.dp, vertical = 16.dp),
                ) {
                    Text(
                        text = localTrack?.shortName ?: state.activeTrack?.name ?: "Aprender",
                        style = MaterialTheme.typography.headlineSmall,
                        color = trackColor,
                    )
                }
            }

            // Node rows — each item occupies NODE_SPACING height and contains one node
            itemsIndexed(state.modules) { index, node ->
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(NODE_SPACING),
                ) {
                    // Bezier connector drawn behind nodes (except first node has no incoming connector)
                    if (index < state.modules.lastIndex) {
                        NodeConnector(
                            fromIndex = index,
                            screenWidth = screenWidth,
                            trackColor = trackColor,
                        )
                    }

                    // Position the node at alternating sides
                    val xFraction = if (index % 2 == 0) 0.7f else 0.3f
                    val nodeXDp = screenWidth * xFraction - NODE_SIZE / 2

                    Box(
                        modifier = Modifier.offset(x = nodeXDp, y = (NODE_SPACING - NODE_SIZE) / 2),
                    ) {
                        StarNode(
                            node = node,
                            trackColor = trackColor,
                            onTap = { onTapNode(node.id) },
                        )
                    }
                }
            }

            // Bottom spacing
            item { Box(Modifier.height(64.dp)) }
        }
    }
}

@Composable
private fun NodeConnector(
    fromIndex: Int,
    screenWidth: Dp,
    trackColor: Color,
) {
    val startX = if (fromIndex % 2 == 0) 0.7f else 0.3f
    val endX = if ((fromIndex + 1) % 2 == 0) 0.7f else 0.3f

    Canvas(
        modifier = Modifier
            .fillMaxWidth()
            .height(NODE_SPACING),
    ) {
        val w = size.width
        val h = size.height
        val nodeSizePx = NODE_SIZE.toPx()

        val startPoint = Offset(w * startX, nodeSizePx / 2)
        val endPoint = Offset(w * endX, h - nodeSizePx / 2)

        val ctrl1 = Offset(startPoint.x, startPoint.y + h * 0.5f)
        val ctrl2 = Offset(endPoint.x, endPoint.y - h * 0.5f)

        drawBezierPath(startPoint, ctrl1, ctrl2, endPoint, trackColor.copy(alpha = 0.4f))
    }
}

private fun DrawScope.drawBezierPath(
    start: Offset,
    ctrl1: Offset,
    ctrl2: Offset,
    end: Offset,
    color: Color,
) {
    val path = Path().apply {
        moveTo(start.x, start.y)
        cubicTo(ctrl1.x, ctrl1.y, ctrl2.x, ctrl2.y, end.x, end.y)
    }
    drawPath(
        path = path,
        color = color,
        style = Stroke(width = 4.dp.toPx(), cap = StrokeCap.Round),
    )
}

@Composable
private fun LoadingContent() {
    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text("Carregando...", style = MaterialTheme.typography.bodyLarge)
    }
}

@Composable
private fun ErrorContent() {
    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(
            "Erro ao carregar. Tente novamente.",
            style = MaterialTheme.typography.bodyLarge,
            color = MaterialTheme.colorScheme.error,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
private fun NoTrackContent() {
    Box(
        Modifier.fillMaxSize().padding(32.dp),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            "Nenhuma matéria selecionada.\nEscolha uma matéria no seu perfil.",
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}

@Composable
fun StarNode(
    node: ModuleNode,
    trackColor: Color,
    onTap: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val infiniteTransition = rememberInfiniteTransition()
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.25f,
        animationSpec = infiniteRepeatable(
            animation = tween(800),
            repeatMode = RepeatMode.Reverse,
        ),
    )

    val outlineColor = MaterialTheme.colorScheme.outline
    val surfaceVariant = MaterialTheme.colorScheme.surfaceVariant

    Box(
        modifier = modifier.size(NODE_SIZE),
        contentAlignment = Alignment.Center,
    ) {
        Canvas(
            modifier = Modifier
                .size(NODE_SIZE)
                .noRippleClickable(onTap),
        ) {
            val center = Offset(size.width / 2, size.height / 2)
            val radius = size.minDimension / 2

            // Pulsing ring for ACTIVE node
            if (node.state == NodeState.ACTIVE) {
                drawCircle(
                    color = trackColor.copy(alpha = 0.3f),
                    radius = radius * pulseScale,
                    center = center,
                )
            }

            // Star fill
            val fillColor = when (node.state) {
                NodeState.COMPLETED, NodeState.ACTIVE -> trackColor
                NodeState.LOCKED -> surfaceVariant
            }
            drawStarPath(center, radius * 0.85f, fillColor)

            // Star border
            val borderColor = when (node.state) {
                NodeState.COMPLETED -> Color(0xFFFFD700) // gold
                NodeState.ACTIVE -> Color.White
                NodeState.LOCKED -> outlineColor
            }
            val borderWidth = when (node.state) {
                NodeState.LOCKED -> 1.dp.toPx()
                else -> 2.dp.toPx()
            }
            drawStarPath(center, radius * 0.85f, borderColor, strokeWidth = borderWidth)
        }

        // Lock icon for LOCKED nodes
        if (node.state == NodeState.LOCKED) {
            Text("🔒", style = MaterialTheme.typography.labelSmall)
        }
    }
}

private fun DrawScope.drawStarPath(
    center: Offset,
    radius: Float,
    color: Color,
    strokeWidth: Float? = null,
) {
    val path = Path()
    val points = 5
    val innerRadius = radius * 0.4f
    val startAngle = -kotlin.math.PI / 2

    for (i in 0 until points * 2) {
        val angle = startAngle + i * kotlin.math.PI / points
        val r = if (i % 2 == 0) radius else innerRadius
        val x = center.x + (r * kotlin.math.cos(angle)).toFloat()
        val y = center.y + (r * kotlin.math.sin(angle)).toFloat()
        if (i == 0) path.moveTo(x, y) else path.lineTo(x, y)
    }
    path.close()

    if (strokeWidth != null) {
        drawPath(path, color, style = Stroke(width = strokeWidth))
    } else {
        drawPath(path, color)
    }
}

private fun Modifier.noRippleClickable(onClick: () -> Unit): Modifier =
    this.then(Modifier.clickable(onClick = onClick))
