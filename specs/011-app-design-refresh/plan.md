# Implementation Plan: App Design Refresh — Core Flow

**Branch**: `011-app-design-refresh` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-app-design-refresh/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Refresh the app's UI to match the wireframed user journey and “warm, motivating” feel: consistent design tokens, coherent scaffolding (top status + bottom navigation), Portuguese-friendly copy, and intentional empty/locked states. The mascot is represented via a placeholder treatment until final artwork is available.

## Technical Context

**Language/Version**: Kotlin 2.1.20 (Compose Multiplatform 1.8.0)
**Primary Dependencies**: Material Design 3 (already in classpath), Decompose 3.3.0, MVIKotlin 4.2.0
**Storage**: N/A (UI/UX refactor; keep existing device-local flags/state such as welcome + active track)
**Testing**: Manual visual review via Kotlin/Wasm web; unit tests only for any new pure logic introduced during refactor
**Target Platform**: Kotlin/Wasm web (primary visual reference); Android and iOS inherit through shared Compose
**Project Type**: multiplatform-ui (shared composables + Decompose components in `mobile/shared/`)
**Performance Goals**: 60 fps rendering; interaction responsiveness (INP) ≤ 200 ms (constitution)
**Constraints**: No new external UI dependencies; design tokens (colors/shapes/spacing/typography) must be centralized; consistent empty/loading/error states; accessibility semantics for all interactive elements
**Scale/Scope**: Refactor the core wireframed flow (~16 screens) spanning welcome, auth, track selection, main navigation, writing journey, friends, league, profile and settings; keep backend unchanged

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Code Quality — readability & SRP | ✓ PASS | Keep UI scaffolding as small composables; avoid catch-all components |
| Code Quality — no dead code | ✓ PASS | Remove unused strings/composables after nav rewrite |
| UX Consistency — design tokens only | ⚠️ IN SCOPE | Current screens contain ad-hoc spacing (`dp`). This feature introduces spacing tokens and migrates all touched screens to tokens; no new raw values allowed in modified areas |
| UX Consistency — copy consistency | ⚠️ IN SCOPE | Some profile strings are English; this feature normalizes core flow copy to Portuguese tone |
| Accessibility — labels & keyboard | ✓ PASS | Maintain `contentDescription` for icons; use Compose semantics for interactive elements |
| Performance — no bundle growth | ✓ PASS | No new dependencies; avoid new heavy assets |
| Testing — test-first | ✓ PASS | No new business logic planned; any new non-trivial pure logic added must have unit tests |
| Testing — screenshot/visual | ⚠️ EXCEPTION | KMP cross-target screenshot testing is not mature; manual visual checklist via web is acceptable |

*Post-Phase 1 re-check: all gates expected to PASS except the screenshot-testing exception (tracked below).*

## Project Structure

### Documentation (this feature)

```text
specs/011-app-design-refresh/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
```text
mobile/shared/src/commonMain/kotlin/com/preuni/shared/
├── PreuniApp.kt                                   # MODIFIED — new scaffold + updated tab mapping
├── presentation/
│   ├── RootComponent.kt                           # MODIFIED — keep flow; ensure welcome/auth/onboarding/main routing still coherent
│   ├── main/MainComponent.kt                      # MODIFIED — update tab enum mapping and selected tab defaults
│   ├── navigation/
│   │   ├── BottomNavigation.kt                    # MODIFIED — match wireframe destinations + labels
│   │   └── BottomTab.kt                           # MODIFIED (or merged) — new destinations list
│   ├── welcome/WelcomeScreen.kt                   # MODIFIED — mascot placeholder + visual tone
│   ├── auth/{LoginScreen,RegisterScreen,...}.kt   # MODIFIED — wireframe-aligned layout + copy
│   ├── onboarding/OnboardingScreen.kt             # MODIFIED — “Escolher matéria” matches wireframe tone
│   ├── learn/LearnScreen.kt                       # MODIFIED — “Trilha” framing and header treatment
│   ├── simulate/SimulateScreen.kt                 # MODIFIED/REPLACED — becomes the writing experience entry (or writing placeholder)
│   ├── friends/                                   # NEW — placeholder screens aligned to wireframe
│   ├── league/                                    # NEW — placeholder screens aligned to wireframe
│   └── settings/                                  # NEW — settings screen aligned to wireframe (UI only)
└── ui/
    ├── theme/
    │   ├── PreuniTheme.kt                         # existing
    │   ├── Color.kt                               # existing
    │   ├── Shape.kt                               # existing
    │   ├── SubjectTrackColors.kt                  # existing
    │   └── Spacing.kt                             # NEW — spacing tokens to reduce ad-hoc `dp`
    └── components/
        ├── PreuniButton.kt                        # existing
        ├── PreuniTextField.kt                     # existing
        ├── SubjectTrackCard.kt                    # existing
        ├── MascotPlaceholder.kt                   # NEW — placeholder for future mascot
        ├── TopStatusBar.kt                        # NEW — compact status row (streak/XP/etc)
        └── SectionHeader.kt                       # NEW — consistent headings for lists/blocks
```

**Structure Decision**: Mobile/KMP shared UI is the source of truth. All UX changes land in `mobile/shared/` so Android, iOS, and web share the same flow and visual language. New screens are added only where wireframe destinations do not exist yet (friends/league/settings), and they start as UI placeholders with clearly labeled empty states.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| No screenshot tests | Compose Multiplatform does not provide stable cross-target screenshot testing for Wasm + iOS targets in this repo. | Manual review via Wasm dev server is sufficient for a visual refactor and matches prior UI features. |

## Phase 0 — Research (output: `research.md`)

- Confirm what already exists in the codebase (current `PreuniTheme`, core components, existing screen structure) and decide what is reused vs refactored.
- Translate the wireframe into a concrete navigation mapping (tab names, default tab, and what each destination renders).
- Define the mascot placeholder treatment (shape, placement, tone, accessibility label).
- Define consistent empty/loading/error patterns for friends/league/settings placeholders.

## Phase 1 — Design & Contracts (outputs: `data-model.md`, `contracts/*`, `quickstart.md`)

- Define the design token surface needed for the refresh (spacing + any additional semantic colors).
- Document UI contracts for reusable building blocks introduced by the refresh (top status bar, bottom navigation, mascot placeholder, list items).
- Provide quickstart instructions to run the web preview and manually review screens against acceptance criteria.

## Phase 1 — Agent Context Update

- Run `.specify/scripts/bash/update-agent-context.sh claude` after contracts are written to keep the agent context aligned with new UI primitives.
