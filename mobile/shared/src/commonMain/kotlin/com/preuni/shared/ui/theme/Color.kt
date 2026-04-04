package com.preuni.shared.ui.theme

import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.ui.graphics.Color

// ─── Brand primitives ───────────────────────────────────────────────────────

private val Violet40 = Color(0xFF6750A4)
private val Violet90 = Color(0xFFEADDFF)
private val Violet10 = Color(0xFF21005D)

private val VioletVariant40 = Color(0xFF625B71)
private val VioletVariant90 = Color(0xFFE8DEF8)
private val VioletVariant10 = Color(0xFF1D192B)

private val Tertiary40 = Color(0xFF7D5260)
private val Tertiary90 = Color(0xFFFFD8E4)
private val Tertiary10 = Color(0xFF31111D)

private val Error40 = Color(0xFFB3261E)
private val Error90 = Color(0xFFF9DEDC)
private val Error10 = Color(0xFF410E0B)

private val Neutral10 = Color(0xFF1C1B1F)
private val Neutral90 = Color(0xFFE6E1E5)
private val Neutral99 = Color(0xFFFFFBFE)

private val NeutralVariant30 = Color(0xFF49454F)
private val NeutralVariant50 = Color(0xFF79747E)
private val NeutralVariant80 = Color(0xFFCAC4D0)
private val NeutralVariant90 = Color(0xFFE7E0EC)

// ─── Semantic status colors (used for password strength indicator) ──────────

internal val WarningAmber = Color(0xFFF59E0B)
internal val SuccessGreen = Color(0xFF10B981)

// ─── Light color scheme ─────────────────────────────────────────────────────

val preuniLightColorScheme = lightColorScheme(
    primary = Violet40,
    onPrimary = Color.White,
    primaryContainer = Violet90,
    onPrimaryContainer = Violet10,

    secondary = VioletVariant40,
    onSecondary = Color.White,
    secondaryContainer = VioletVariant90,
    onSecondaryContainer = VioletVariant10,

    tertiary = Tertiary40,
    onTertiary = Color.White,
    tertiaryContainer = Tertiary90,
    onTertiaryContainer = Tertiary10,

    error = Error40,
    onError = Color.White,
    errorContainer = Error90,
    onErrorContainer = Error10,

    background = Neutral99,
    onBackground = Neutral10,
    surface = Neutral99,
    onSurface = Neutral10,

    surfaceVariant = NeutralVariant90,
    onSurfaceVariant = NeutralVariant30,
    outline = NeutralVariant50,
    outlineVariant = NeutralVariant80,
)
