# Implementation Plan: Onboarding Flow and UX Fixes

**Branch**: `005-onboarding-ux-fixes` | **Date**: 2026-04-07 | **Spec**: [spec.md](./spec.md)

## Summary

Three fixes targeting the post-auth UX: (1) split the 4-page onboarding into a pre-login welcome flow (pages 0–2) + a post-login track-selection flow (page 3), gated by a locally-stored "welcome seen" flag in `SecureStorage`; (2) add a "Alterar matérias" path from `ProfileComponent` backed by a new `ChangeTrackStore` that calls the existing `PATCH /v1/students/me/onboarding` endpoint; (3) add auto-retry to `HomeStore` so transient profile-load failures recover silently instead of landing users in the error state.

## Technical Context

**Language/Version**: Kotlin 2.1.20 / Compose Multiplatform 1.8.0  
**Primary Dependencies**: Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Ktor Client (HTTP)  
**Storage**: `SecureStorage` (expect/actual, already exists) — key-value local store  
**Testing**: KMP `desktopTest` (JVM), `runTest` + `UnconfinedTestDispatcher`, `MockEngine`  
**Target Platform**: Android, iOS, Kotlin/Wasm (web) via Compose Multiplatform  
**Project Type**: Cross-platform mobile + web app  
**Performance Goals**: Navigation transitions ≤ 200 ms INP; Home tab data load ≤ 3 s  
**Constraints**: No new dependencies; no schema changes; no new backend endpoints  
**Scale/Scope**: Affects 4 existing KMP files + 1 new component (WelcomeComponent); 0 backend changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Test-first for new features | PASS | `WelcomeStore`, `ChangeTrackStore`, `HomeStore` retry logic all require unit tests before implementation |
| Unit tests for business logic | PASS | New stores get `runTest`-based unit tests; `HomeStore` retry test added |
| Design system adherence | PASS | All new composables use Material 3 tokens only; no raw hex/px values |
| Consistent feedback patterns | PASS | Loading, error, and empty states follow existing `HomeStore`/`ProfileStore` patterns |
| Mobile-first | PASS | No layout changes; existing responsive structure preserved |
| No premature abstraction | PASS | `ChangeTrackStore` is standalone (same pattern as existing stores); no shared base class invented |
| Single responsibility | PASS | `WelcomeComponent` handles only welcome slides + seen-flag write; `OnboardingComponent` handles only track selection |
| No dead code | PASS | Pages 0–2 move from `OnboardingScreen` to `WelcomeScreen`; the 4-page pager and `TOTAL_PAGES = 4` constant are removed |

**No violations to justify.**

## Project Structure

### Documentation (this feature)

```text
specs/005-onboarding-ux-fixes/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (affected paths)

```text
mobile/shared/src/commonMain/kotlin/com/preuni/shared/
├── data/auth/
│   └── SecureStorage.kt              # add KEY_WELCOME_SEEN constant + welcomeSeen()/markWelcomeSeen()
├── presentation/
│   ├── welcome/                      # NEW package
│   │   ├── WelcomeComponent.kt       # NEW
│   │   ├── WelcomeScreen.kt          # NEW (moves pages 0–2 from OnboardingScreen)
│   │   └── WelcomeStore.kt           # NEW (pageIndex, skip/next/complete intents)
│   ├── onboarding/
│   │   ├── OnboardingScreen.kt       # MODIFIED: remove pages 0–2; track-selection only
│   │   └── OnboardingStore.kt        # MODIFIED: TOTAL_PAGES → 1; remove NextPage/PreviousPage routing for welcome pages
│   ├── profile/
│   │   ├── ProfileComponent.kt       # MODIFIED: add Config.ChangeTrack + navigateToChangeTrack()
│   │   └── ProfileScreen.kt          # MODIFIED: add "Alterar matérias" OutlinedButton
│   ├── home/
│   │   └── HomeStore.kt              # MODIFIED: auto-retry up to 3× with 1 s delay before emitting error
│   └── RootComponent.kt              # MODIFIED: add Config.Welcome; fix initialConfig routing
│
├── commonTest/kotlin/com/preuni/shared/
│   ├── presentation/welcome/
│   │   └── WelcomeStoreTest.kt       # NEW
│   ├── presentation/profile/
│   │   └── ChangeTrackStoreTest.kt   # NEW
│   └── presentation/home/
│       └── HomeStoreTest.kt          # MODIFIED: add retry test cases
```

**Structure Decision**: Pure KMP shared-module changes. No backend, no new Gradle modules, no new dependencies. Backend track-enrollment is handled by the existing `PATCH /v1/students/me/onboarding` endpoint.
