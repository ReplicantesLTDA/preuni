# Data Model: App Design Refresh — Core Flow

**Feature**: `011-app-design-refresh` | **Date**: 2026-05-25

This feature is primarily UI/UX and navigation refactor in the KMP shared module. No new database tables or backend API contracts are introduced. The “data model” here describes the UI-facing entities and token surfaces that support consistent design.

---

## Entity 1: AppDestination (Bottom Tabs)

Represents a top-level destination in the wireframed bottom navigation.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `String` | Stable identifier (e.g., `"trilha"`, `"redacao"`) |
| `label` | `String` | Portuguese tab label shown to users |
| `icon` | `IconSpec` | Icon representation (must have accessibility label) |

**Validation rules**:
- `label` must be short enough to fit common mobile widths without truncating the primary meaning.
- Each destination must be uniquely identifiable by icon + label.

**State transitions**: selected/unselected only.

**Expected location**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/presentation/navigation/`

---

## Entity 2: StatusMetric (Top Status Bar)

Represents one metric displayed in the persistent top status bar.

| Field | Type | Description |
|-------|------|-------------|
| `key` | `String` | Stable key (e.g., `"streak"`, `"xp"`) |
| `label` | `String` | Short Portuguese label (optional if icon-only) |
| `valueText` | `String` | Renderable value (e.g., `"7"`, `"1.480"`) |
| `isPlaceholder` | `Boolean` | True when the metric is not backed by real data yet |

**Notes**:
- The wireframe contains multiple counters. Today the app reliably has streak + XP; additional counters can be placeholders.

---

## Entity 3: MascotPlaceholder

Represents the placeholder content for the future mascot.

| Field | Type | Description |
|-------|------|-------------|
| `contentType` | `enum` | `EMOJI` / `ICON` / `MONOGRAM` (one is chosen) |
| `accessibilityLabel` | `String` | What screen readers should announce |
| `tone` | `enum` | Friendly/neutral; must match the welcome flow tone |

**Constraint**: Must not introduce custom third-party artwork.

---

## Entity 4: Design Tokens (Theme Extensions)

This feature expands the design-token surface beyond colors/shapes by adding spacing tokens.

| Token Set | Purpose | Examples |
|----------|---------|----------|
| `ColorScheme` | Brand colors and semantic colors | existing `preuniLightColorScheme` |
| `Shapes` | Corner radii scale | existing `preuniShapes` |
| `Spacing` | Consistent layout rhythm | xs/sm/md/lg/xl |

**Rule**: UI code touched in this feature should consume these tokens instead of introducing new one-off values.

---

## Relationships

```
PreuniTheme
  ├── ColorScheme
  ├── Shapes
  └── Spacing (new)

BottomNavigation (composable)
  └── AppDestination (×5)

TopStatusBar (composable)
  └── StatusMetric (×N)

Welcome/Auth screens
  └── MascotPlaceholder
```
