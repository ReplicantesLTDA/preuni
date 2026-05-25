# UI Contract: BottomNavigation (Wireframe Tabs)

**Component**: `BottomNavigation`
**Package**: `com.preuni.shared.presentation.navigation`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/navigation/BottomNavigation.kt`

---

## Purpose

Provide the primary navigation across the wireframed destinations: Trilha, Redação, Amigos, Liga, Perfil.

## Behavior

- Exactly 5 tabs.
- Selected state is always obvious.
- Icons have accessibility labels.

## Visual States

| State | Appearance |
|-------|-----------|
| Selected | Highlighted icon + label using theme tokens |
| Unselected | Muted icon + label using theme tokens |

## Constraints

- Tab labels must be Portuguese and match the wireframe language.
- No new icon packs beyond what’s already available in the project.

## Acceptance Criteria

- [ ] Switching tabs updates the visible destination.
- [ ] Selected tab is visually distinct.
- [ ] All icons have meaningful `contentDescription`.