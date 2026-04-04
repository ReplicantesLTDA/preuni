# Implementation Plan: Friendly UI Polish — Core Frontend

**Branch**: `003-ui-polish` | **Date**: 2026-04-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-ui-polish/spec.md`

## Summary

Apply a cohesive, friendly visual identity to PreUni's core screens (auth, home, onboarding, navigation) by introducing a central `PreuniTheme` with custom Material 3 shape and color tokens. Rounded inputs, pill-shaped primary actions, subject-track color coding for the five ENEM areas, and warm Portuguese copy make every screen feel intentional and brand-consistent. No new external libraries — only Material 3 APIs already on the classpath.

## Technical Context

**Language/Version**: Kotlin 2.1.20 (Compose Multiplatform 1.8.0)
**Primary Dependencies**: Material Design 3 (already in classpath), Decompose 3.3.0, MVIKotlin 4.2.0
**Storage**: N/A (UI layer only — no persistence changes)
**Testing**: Manual visual review against acceptance criteria (Kotlin/Wasm dev server); no screenshot testing framework mature enough for KMP triple-target — justified exception below
**Target Platform**: Kotlin/Wasm web (primary visual reference); Android and iOS inherit through shared Compose
**Project Type**: multiplatform-ui (shared composables in `mobile/shared/`)
**Performance Goals**: 60 fps rendering, INP ≤ 200 ms (per constitution)
**Constraints**: No new transitive dependencies; light mode only; WCAG AA contrast ≥ 4.5:1 for all text; LCP ≤ 2.5 s (per constitution)
**Scale/Scope**: ~15 new/modified composables across 4 screen groups; 1 new theme module; 5 mock subject tracks

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Code Quality — no raw values | ✓ PASS | All colors/shapes defined in `PreuniTheme`; screens reference tokens only |
| Code Quality — single responsibility | ✓ PASS | Theme module isolated; component composables ≤ 60 lines each |
| Testing — unit tests | ✓ PASS | Pure data objects (SubjectTrack, DesignToken) have no logic requiring tests |
| Testing — screenshot / visual | ⚠️ EXCEPTION | Compose Multiplatform lacks mature KMP screenshot testing (see Complexity Tracking) |
| UX Consistency — design tokens | ✓ PASS | Central `PreuniTheme` is the design token source; this feature creates it |
| UX Consistency — WCAG AA | ✓ PASS | Color selection in research verified contrast ≥ 4.5:1 against backgrounds |
| Performance — no bundle growth | ✓ PASS | No new dependencies; shape/color tokens are compile-time objects |
| Accessibility — ARIA / keyboard | ✓ PASS | Compose web generates standard HTML inputs; no degradation |

*Post-Phase 1 re-check: all gates pass — no new violations introduced during design.*

## Project Structure

### Documentation (this feature)

```text
specs/003-ui-polish/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   ├── PreuniButton.md
│   ├── PreuniTextField.md
│   └── SubjectTrackCard.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
mobile/shared/src/commonMain/kotlin/com/preuni/shared/
├── ui/
│   ├── theme/
│   │   ├── PreuniTheme.kt          # MaterialTheme wrapper (colors + shapes + typography)
│   │   ├── Color.kt                # Full MD3 light color scheme (violet primary)
│   │   ├── Shape.kt                # Custom shape scale (rounded inputs, pill buttons)
│   │   └── SubjectTrackColors.kt   # 5 ENEM track color tokens
│   └── components/
│       ├── PreuniButton.kt         # Pill-shaped primary button
│       ├── PreuniTextField.kt      # Rounded outlined text field
│       └── SubjectTrackCard.kt     # Colored card for ENEM subject tracks
└── presentation/
    ├── auth/
    │   ├── LoginScreen.kt          # MODIFIED — use PreuniTheme components
    │   ├── RegisterScreen.kt       # MODIFIED
    │   ├── VerifyEmailScreen.kt    # MODIFIED
    │   └── OtpLoginScreen.kt       # MODIFIED
    ├── home/
    │   └── HomeScreen.kt           # MODIFIED — add track cards, greeting warmth
    ├── onboarding/
    │   └── OnboardingScreen.kt     # MODIFIED — pill CTAs, track card on page 4
    └── navigation/
        └── BottomNavigation.kt     # MODIFIED — proper icons per tab
```

**Structure Decision**: New `ui/theme/` and `ui/components/` packages under `shared`. All existing presentation files are modified in-place — no new screen files needed. The theme wrapper (`PreuniTheme`) is placed at the `PreuniApp` call site in `PreuniApp.kt`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| No screenshot tests | Compose Multiplatform 1.8.0 screenshot testing targets JVM-only (Robolectric on Android). Web (Wasm) and iOS targets have no equivalent headless renderer available. | Skipping entirely is acceptable given the feature is purely visual and manually reviewable via `make run-web`. This exception is tracked and should be revisited when KMP screenshot testing matures. |
