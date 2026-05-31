package com.preuni.shared.ui.theme

import androidx.compose.runtime.Immutable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

/**
 * Spacing scale used across the preuni UI.
 *
 * Use these tokens instead of raw `dp` literals so spacing stays consistent
 * across screens and easy to adjust globally. The names match the 4-pt scale
 * the wireframe uses.
 */
@Immutable
data class PreuniSpacing(
    val xxs: Dp = 2.dp,
    val xs: Dp = 4.dp,
    val sm: Dp = 8.dp,
    val md: Dp = 12.dp,
    val lg: Dp = 16.dp,
    val xl: Dp = 24.dp,
    val xxl: Dp = 32.dp,
    val xxxl: Dp = 48.dp,
)

val preuniSpacing = PreuniSpacing()

/**
 * CompositionLocal so screens can read spacing via `LocalSpacing.current.lg`
 * inside a `PreuniTheme { ... }` scope.
 */
val LocalSpacing = staticCompositionLocalOf { preuniSpacing }
