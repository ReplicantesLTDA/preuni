# UI Contract: PreuniButton

**Component**: `PreuniButton`
**Package**: `com.preuni.shared.ui.components`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/PreuniButton.kt`

---

## Purpose

Pill-shaped primary action button. Wraps Material 3 `Button` with enforced `fillMaxWidth` modifier and a minimum touch target height of 52dp for comfortable mobile tap areas.

## Signature

```kotlin
@Composable
fun PreuniButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    isLoading: Boolean = false,
)
```

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `text` | `String` | Yes | — | Button label text |
| `onClick` | `() -> Unit` | Yes | — | Click callback |
| `modifier` | `Modifier` | No | `Modifier` | Compose modifier chain |
| `enabled` | `Boolean` | No | `true` | When false: muted appearance, no click events |
| `isLoading` | `Boolean` | No | `false` | Shows `CircularProgressIndicator` instead of text |

## Visual States

| State | Appearance |
|-------|-----------|
| Default | `colorScheme.primary` fill, white text, pill corners |
| Pressed | Material 3 ripple on `primary` |
| Disabled | `colorScheme.onSurface` at 38% opacity fill |
| Loading | `CircularProgressIndicator` (20dp, strokeWidth 2dp) centered; `enabled = false` implied |

## Constraints

- Always `fillMaxWidth` unless caller explicitly overrides `modifier` with width constraint.
- Minimum height: 52dp (ensures comfortable touch target on mobile).
- Text style: `MaterialTheme.typography.labelLarge` (M3 button default).
- Shape: `MaterialTheme.shapes.extraLarge` (50dp radius = pill) — inherits from theme.

## Usage Example

```kotlin
PreuniButton(
    text = "Entrar",
    onClick = { store.accept(LoginStore.Intent.Submit) },
    isLoading = state.isLoading,
)
```

## Acceptance Criteria

- [ ] Corners are fully rounded (pill shape) at all viewport widths.
- [ ] Loading state shows spinner, button is non-interactive.
- [ ] Disabled state is visually distinct from enabled.
- [ ] Minimum touch area ≥ 44×44dp (WCAG 2.5.5).
