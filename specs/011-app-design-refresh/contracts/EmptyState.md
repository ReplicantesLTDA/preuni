# UI Contract: EmptyState (Friendly Placeholder)

**Component**: `EmptyState`
**Package**: `com.preuni.shared.ui.components`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/EmptyState.kt`

---

## Purpose

Standardize empty/coming-soon states so placeholder screens still feel polished and encouraging.

## Signature

```kotlin
@Composable
fun EmptyState(
    title: String,
    body: String,
    primaryActionText: String? = null,
    onPrimaryAction: (() -> Unit)? = null,
    modifier: Modifier = Modifier,
)
```

## Behavior

- Always shows a title and a short body.
- Optional primary action can be shown when there is a meaningful next step.

## Constraints

- Copy must be friendly, direct, and non-technical.
- Uses theme tokens only.

## Acceptance Criteria

- [ ] Empty states look like part of the same product.
- [ ] Copy explains what exists now vs what comes next.
- [ ] If an action is present, it is clearly a CTA and uses the primary button style.
