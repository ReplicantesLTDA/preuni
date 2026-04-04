package com.preuni.shared.ui.theme

import androidx.compose.ui.graphics.Color

data class TrackColorScheme(
    val container: Color,
    val onContainer: Color,
)

data class SubjectTrack(
    val id: String,
    val name: String,
    val shortName: String,
    val emoji: String,
    val colorScheme: TrackColorScheme,
)

// WCAG AA contrast ratios verified in specs/003-ui-polish/data-model.md
val SubjectTracks = listOf(
    SubjectTrack(
        id = "matematica",
        name = "Matemática e suas Tecnologias",
        shortName = "Matemática",
        emoji = "🧮",
        colorScheme = TrackColorScheme(
            container = Color(0xFFE3F2FD),
            onContainer = Color(0xFF0D47A1),
        ),
    ),
    SubjectTrack(
        id = "linguagens",
        name = "Linguagens, Códigos e suas Tecnologias",
        shortName = "Linguagens",
        emoji = "📖",
        colorScheme = TrackColorScheme(
            container = Color(0xFFE8F5E9),
            onContainer = Color(0xFF1B5E20),
        ),
    ),
    SubjectTrack(
        id = "ciencias-natureza",
        name = "Ciências da Natureza e suas Tecnologias",
        shortName = "Ciências da Natureza",
        emoji = "🔬",
        colorScheme = TrackColorScheme(
            container = Color(0xFFF1F8E9),
            onContainer = Color(0xFF33691E),
        ),
    ),
    SubjectTrack(
        id = "ciencias-humanas",
        name = "Ciências Humanas e suas Tecnologias",
        shortName = "Ciências Humanas",
        emoji = "🌎",
        colorScheme = TrackColorScheme(
            container = Color(0xFFFFF3E0),
            onContainer = Color(0xFFE65100),
        ),
    ),
    SubjectTrack(
        id = "redacao",
        name = "Redação",
        shortName = "Redação",
        emoji = "✏️",
        colorScheme = TrackColorScheme(
            container = Color(0xFFEDE7F6),
            onContainer = Color(0xFF4527A0),
        ),
    ),
)
