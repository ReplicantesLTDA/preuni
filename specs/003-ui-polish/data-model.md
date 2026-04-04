# Data Model: Friendly UI Polish — Core Frontend

**Feature**: `003-ui-polish` | **Date**: 2026-04-04

This feature is UI-layer-only. No new database tables or API contracts. The data model describes the Compose/Kotlin entities that back the design system.

---

## Entity 1: SubjectTrack

Represents one of the five ENEM subject areas. Static data — no persistence required.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `String` | Stable machine identifier (e.g., `"matematica"`) |
| `name` | `String` | Full Portuguese name (e.g., `"Matemática e suas Tecnologias"`) |
| `shortName` | `String` | Abbreviated label for cards (e.g., `"Matemática"`) |
| `emoji` | `String` | Single emoji for visual association (e.g., `"🧮"`) |
| `colorScheme` | `TrackColorScheme` | Container + on-container color pair |

**Validation rules**: `id` is a non-empty lowercase string with no spaces. The five instances are sealed/companion objects — no user-created instances.

**State transitions**: N/A (static data)

**Location**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/SubjectTrackColors.kt`

---

## Entity 2: TrackColorScheme

Color pair for a subject track card. Inlined into `SubjectTrack`.

| Field | Type | Description |
|-------|------|-------------|
| `container` | `Color` | Card background color (light, WCAG AA verified) |
| `onContainer` | `Color` | Text/icon color on that background (WCAG AA verified) |

**Validation rules**: Both colors must maintain ≥ 4.5:1 contrast ratio. See research.md § 3 for verified values.

---

## Entity 3: DesignToken (implicit — MaterialTheme extension)

Not a data class — these are the Material 3 theme token overrides in `PreuniTheme.kt`.

| Token | Value | Applied To |
|-------|-------|-----------|
| `shapes.small` | `RoundedCornerShape(16.dp)` | Text fields (`PreuniTextField`) |
| `shapes.medium` | `RoundedCornerShape(20.dp)` | Cards |
| `shapes.extraLarge` | `RoundedCornerShape(50.dp)` | Buttons (M3 default — kept) |
| `colorScheme.primary` | `#6750A4` | CTAs, active nav tab, focused input border |
| `colorScheme.primaryContainer` | `#EADDFF` | CTA card background on Home |
| `colorScheme.error` | `#B3261E` | Inline validation error text |

**Location**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/theme/`

---

## Five ENEM Subject Tracks (Static Data)

```kotlin
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
```

---

## Relationships

```
PreuniTheme
  ├── ColorScheme (M3 violet baseline)
  ├── Shapes (custom scale)
  └── Typography (M3 default)

SubjectTrack (×5, static)
  └── TrackColorScheme
        ├── container: Color
        └── onContainer: Color

SubjectTrackCard (composable)
  └── SubjectTrack

PreuniTextField (composable)
  └── uses MaterialTheme.shapes.small (16dp)

PreuniButton (composable)
  └── uses MaterialTheme.shapes.extraLarge (pill, M3 default)
```
