# UI Contract: SubjectTrackCard

**Component**: `SubjectTrackCard`
**Package**: `com.preuni.shared.ui.components`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/SubjectTrackCard.kt`

---

## Purpose

Tappable card representing one ENEM subject area. Uses the area's `TrackColorScheme` for visual identity. Used on the onboarding track-selection screen and the home screen subject grid.

## Signature

```kotlin
@Composable
fun SubjectTrackCard(
    track: SubjectTrack,
    selected: Boolean = false,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
)
```

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `track` | `SubjectTrack` | Yes | — | The subject area to display |
| `selected` | `Boolean` | No | `false` | Selection state (onboarding use case) |
| `onClick` | `() -> Unit` | Yes | — | Tap callback |
| `modifier` | `Modifier` | No | `Modifier` | Compose modifier chain |

## Visual States

| State | Appearance |
|-------|-----------|
| Default | `track.colorScheme.container` fill, `track.colorScheme.onContainer` text, `shapes.medium` (20dp) corners |
| Selected | Same as default + `colorScheme.primary` border (2dp) + checkmark overlay or border highlight |
| Pressed | Material 3 ripple on container color |

## Layout

```
┌─────────────────────────┐
│  [emoji]  [shortName]   │  ← Row, vertically centered
│  (optional progress bar)│  ← shown only on Home screen
└─────────────────────────┘
```

- Card height: 72dp min (comfortable touch target)
- Emoji size: `headlineSmall` text size
- Subject name: `titleMedium`, color = `track.colorScheme.onContainer`
- Corners: `MaterialTheme.shapes.medium` (20dp)

## Variants

| Variant | Where Used | Notes |
|---------|-----------|-------|
| `selected = false/true` | Onboarding track selection | Multi-select; at least 1 required to proceed |
| no selection UI | Home screen track grid | `selected` ignored, no border/checkmark rendered |

## Constraints

- Width: `fillMaxWidth` by default when used in a single-column list; caller controls in grid layouts.
- The five tracks are always displayed in the canonical order: Matemática → Linguagens → Ciências da Natureza → Ciências Humanas → Redação.

## Usage Example

```kotlin
SubjectTrackCard(
    track = SubjectTracks[0], // Matemática
    selected = state.selectedTrackIds.contains("matematica"),
    onClick = { store.accept(OnboardingStore.Intent.ToggleTrack("matematica")) },
)
```

## Acceptance Criteria

- [ ] Each track card uses its distinct container background color.
- [ ] Text on each card passes WCAG AA contrast (≥ 4.5:1) — verified in data-model.md.
- [ ] Selected state shows a clear visual distinction (border or checkmark).
- [ ] Card is tappable with minimum 44dp height touch target.
- [ ] Emoji and short name are visible without overflow at 360px viewport width.
