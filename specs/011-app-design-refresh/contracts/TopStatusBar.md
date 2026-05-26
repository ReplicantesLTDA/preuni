# UI Contract: TopStatusBar

**Component**: `TopStatusBar`
**Package**: `com.preuni.shared.ui.components`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/TopStatusBar.kt`

---

## Purpose

Show compact, scannable status metrics at the top of core app destinations (wireframe-aligned). Reinforces continuity across sections.

## Signature

```kotlin
@Composable
fun TopStatusBar(
    modifier: Modifier = Modifier,
    metrics: List<StatusMetric>,
)
```

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `metrics` | `List<StatusMetric>` | Yes | — | Ordered list of metrics (streak, XP, placeholders) |
| `modifier` | `Modifier` | No | `Modifier` | Compose modifier chain |

## Behavior

- Displays metrics in a single row with consistent spacing.
- Metrics may be placeholders; placeholders must be visually clear without looking “broken”.

## Constraints

- Uses theme tokens (including spacing tokens).
- Text must remain legible at small widths.

## Acceptance Criteria

- [ ] Metrics do not clip at ~360px width.
- [ ] Streak + XP are always visible when data exists.
- [ ] Placeholder metrics are clearly placeholders.
