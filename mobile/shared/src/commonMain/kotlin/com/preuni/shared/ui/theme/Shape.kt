package com.preuni.shared.ui.theme

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Shapes
import androidx.compose.ui.unit.dp

val preuniShapes = Shapes(
    extraSmall = RoundedCornerShape(12.dp),  // chips, small badges
    small = RoundedCornerShape(16.dp),       // text fields
    medium = RoundedCornerShape(20.dp),      // cards
    large = RoundedCornerShape(28.dp),       // dialogs, bottom sheets
    extraLarge = RoundedCornerShape(50.dp),  // pill buttons
)
