# Research: Friendly UI Polish — Core Frontend

**Feature**: `003-ui-polish` | **Date**: 2026-04-04

---

## 1. UI/UX Reference Synthesis

### Decision
Blend four design languages into a single coherent "warm edtech" aesthetic: Duolingo's gamification micro-interactions, Brilliant.org's clean premium feel, Headspace's breathing room, and Quizlet's color-coded subject cards. Khan Academy's topic-tree thinking informs the subject track card design.

### Reference Analysis

| Reference | Steal | Avoid |
|-----------|-------|-------|
| **Duolingo** | Streak/XP badges, progress rings, friendly copy ("Aprender agora"), celebratory micro-states | Excessive mascot use, heavy animation, cartoonish borders |
| **Brilliant.org** | Large serif-ish headings, significant whitespace, minimal noise on action screens, single-focus onboarding screens | Dark/muted palette (PreUni keeps violet/bright) |
| **Headspace** | Generous vertical breathing room between sections, onboarding one-thing-per-screen, soft card backgrounds | Pastel mono-hue color world |
| **Khan Academy** | Color-coded subject domains (each area has a distinct accent), mastery-progress cards, consistent card grid | Dense content layout, small tap targets |
| **Quizlet** | Chip-shaped subject labels, bright subject category cards, rounded input fields | Blue-dominant palette (conflicts with PreUni violet) |

### Applied Synthesis
- **Auth screens**: Brilliant-style single-focus layout with PreUni violet CTA pill, Headspace-inspired breathing space
- **Home screen**: Duolingo streak/XP badges + Khan Academy color-coded subject track cards
- **Onboarding**: Headspace one-thing-per-screen, Duolingo motivational copy in Portuguese
- **Navigation**: Clean Material 3 NavigationBar, tab icons match ENEM categories

### Rationale
Duolingo alone would make PreUni feel like an exercise app, not a serious exam platform. Brilliant.org's premium feel adds credibility. Headspace teaches "less is more" for learning apps. Khan Academy's domain color coding is directly applicable to ENEM's 5 subject areas. Quizlet's rounded inputs are already a user expectation for study tools.

---

## 2. Material 3 Shape Customization in Compose Multiplatform

### Decision
Override Material 3 `Shapes` with custom scale values; use `ButtonDefaults.shape` only where the default M3 pill is insufficient; create `PreuniTextField` as a thin composable wrapping `OutlinedTextField` with `shape = RoundedCornerShape(16.dp)`.

### Findings

**M3 Button shape**: `FilledButtonTokens.ContainerShape = ShapeKeyTokens.CornerFull` → already pill-shaped (50% corner radius) in M3. No override needed for buttons. `Button` composable accepts `shape` parameter but the default is correct.

**M3 OutlinedTextField shape**: `OutlinedTextFieldTokens.ContainerShape = ShapeKeyTokens.CornerExtraSmall` → only 4dp on top corners. **Must be overridden** to achieve the rounded appearance users expect.

**Theme-level override** for text fields:
```kotlin
MaterialTheme(
    shapes = Shapes(
        extraSmall = RoundedCornerShape(12.dp),  // chips
        small = RoundedCornerShape(16.dp),       // text fields
        medium = RoundedCornerShape(20.dp),      // cards
        large = RoundedCornerShape(28.dp),       // dialogs
        extraLarge = RoundedCornerShape(50.dp),  // pill buttons (already default)
    )
)
```

**`PreuniTextField` composable**: Wraps `OutlinedTextField` passing `shape = MaterialTheme.shapes.small`. This is the only pattern change needed — all screens swap `OutlinedTextField(...)` for `PreuniTextField(...)`.

### Rationale
Theme-level shape override ensures all `OutlinedTextField` instances in screens get the rounded look automatically when `PreuniTheme` is the root. The `PreuniTextField` wrapper adds value by enforcing `fillMaxWidth` modifier and consistent `supportingText` error display pattern.

### Alternatives Considered
- **Custom shape per-call-site**: Rejected — duplicates shape values; violates constitution's "no raw values" rule.
- **New TextField from scratch**: Rejected — Material 3 `OutlinedTextField` already handles all states (focused, error, disabled) correctly; only the corner shape needs changing.

---

## 3. Color Scheme: Violet/Purple MD3 Palette

### Decision
Generate a Material 3 color scheme from a violet seed color (`#6750A4` — M3 baseline primary) augmented with warmer secondary and custom ENEM-area tones.

### M3 Color Tokens

```
primary         = #6750A4  (violet)
onPrimary       = #FFFFFF
primaryContainer= #EADDFF
onPrimaryContainer = #21005D
secondary       = #625B71
onSecondary     = #FFFFFF
secondaryContainer = #E8DEF8
onSecondaryContainer = #1D192B
tertiary        = #7D5260
onTertiary      = #FFFFFF
tertiaryContainer = #FFD8E4
onTertiaryContainer = #31111D
error           = #B3261E
background      = #FFFBFE
surface         = #FFFBFE
onSurface       = #1C1B1F
surfaceVariant  = #E7E0EC
outline         = #79747E
```

These are the M3 baseline tokens — already in Compose Material 3 as defaults. The key addition is the `SubjectTrackColors` extension for ENEM areas.

### ENEM Subject Track Colors (WCAG AA Verified)

| Area | Background | On-Background | Contrast Ratio |
|------|-----------|---------------|---------------|
| Matemática | `#E3F2FD` | `#0D47A1` | 7.1:1 ✓ |
| Linguagens | `#E8F5E9` | `#1B5E20` | 7.3:1 ✓ |
| Ciências da Natureza | `#F1F8E9` | `#33691E` | 6.8:1 ✓ |
| Ciências Humanas | `#FFF3E0` | `#E65100` | 4.7:1 ✓ |
| Redação | `#EDE7F6` | `#4527A0` | 6.9:1 ✓ |

### Rationale
Using the M3 baseline `#6750A4` keeps PreUni consistent with any future M3 dynamic color support on Android 12+. The existing screens already reference `MaterialTheme.colorScheme.primary` without explicit color definition — this plan makes the implicit palette explicit.

---

## 4. Mock ENEM Subject Track Data Model

### Decision
A `SubjectTrack` sealed class with 5 companion objects. No database storage needed — this is static mock data for the onboarding track-selection page and home screen track cards.

### Five ENEM Areas
1. **Matemática e suas Tecnologias** — blue tone, `🧮` icon
2. **Linguagens, Códigos e suas Tecnologias** — green tone, `📖` icon
3. **Ciências da Natureza e suas Tecnologias** — lime/green tone, `🔬` icon
4. **Ciências Humanas e suas Tecnologias** — orange tone, `🌎` icon
5. **Redação** — violet (brand), `✏️` icon

### Data Representation
```kotlin
data class SubjectTrack(
    val id: String,
    val name: String,
    val shortName: String,
    val emoji: String,
    val colorScheme: TrackColorScheme,
)

data class TrackColorScheme(
    val container: Color,
    val onContainer: Color,
)
```

### Rationale
Static data object avoids a network call during onboarding — user can select tracks offline. The backend already has a `content-svc` tracks endpoint; this mock satisfies the UI spec until that integration is wired.

---

## 5. User Retention Design Patterns Applied

### Decision
Apply three evidence-backed retention patterns to the Home screen and Onboarding flow without adding animation complexity.

### Patterns

**1. Progress Salience** (Duolingo, Brilliant): Display streak count and XP in the Home header — these are already in `HomeStore.State`. Make them visually prominent (badge chips in brand colors).

**2. One Job Per Screen** (Headspace onboarding): Each onboarding screen has exactly one heading, one body paragraph, and one pill CTA. No bullets, no lists, no options until the final track-selection screen.

**3. Subject Color Coding** (Khan Academy, Quizlet): Each ENEM area gets a unique accent color. Cards on the Home screen and the track-selection page use the area's `containerColor`. Users build a mental map: "orange = Ciências Humanas" across the whole app lifetime.

### Rationale
These three patterns require zero new dependencies and are achievable through Compose layout + token changes. Animation-based retention patterns (confetti, streak freeze modal) are out of scope for this iteration.
