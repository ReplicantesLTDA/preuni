# UI Contract: MascotPlaceholder

**Component**: `MascotPlaceholder`
**Package**: `com.preuni.shared.ui.components`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/MascotPlaceholder.kt`

---

## Purpose

Provide a friendly, intentional placeholder for the upcoming mascot artwork. The component preserves layout, tone, and accessibility until real assets are available.

## Signature

```kotlin
@Composable
fun MascotPlaceholder(
    modifier: Modifier = Modifier,
    accessibilityLabel: String = "Mascote (em breve)",
)
```

## Behavior

- Renders a stable shape (e.g., circle/rounded container) with a simple placeholder content (emoji/icon/monogram).
- Must not load external images.
- Must provide an accessibility label.

## Visual States

| State | Appearance |
|-------|-----------|
| Default | Friendly placeholder that matches the design system colors and shapes |
| Reduced motion | No animation required (avoid adding motion as part of this contract) |

## Constraints

- No third-party artwork.
- Uses theme tokens only (colors, shapes, spacing).

## Acceptance Criteria

- [ ] Placeholder looks intentional (not “missing image”).
- [ ] Screen readers announce a meaningful label.
- [ ] Layout does not shift when the mascot is replaced later.
