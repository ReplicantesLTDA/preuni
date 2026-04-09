# Implementation Plan: Subject Track Path

**Branch**: `006-subject-track-path` | **Date**: 2026-04-08 | **Spec**: [spec.md](./spec.md)

## Summary

Three tightly coupled fixes: (1) fix the "Começar" button by switching `OnboardingStore` to single-select and making navigation fire-and-forget (save activeTrackId locally, navigate immediately, sync to backend in background); (2) build the gamified S-curve learning path screen (`LearnStore` + `LearnScreen`) with star-shaped module nodes and bezier connecting lines, rendered from stub data; (3) update `ChangeTrackStore` to single-select and wire subject-switch to navigate to the Learn tab. Android-first; KMP shared module only.

## Technical Context

**Language/Version**: Kotlin 2.1.20 / Compose Multiplatform 1.8.0
**Primary Dependencies**: Decompose 3.3.0 (navigation), MVIKotlin 4.2.0 (state), Compose Canvas (path drawing)
**Storage**: `SecureStorage` (expect/actual, already exists) — add `active_track_id` key to `TokenStore`
**Testing**: KMP `desktopTest` (JVM), `runTest` + `UnconfinedTestDispatcher`
**Target Platform**: Android-first; KMP shared module (iOS and web not tested in this feature)
**Performance Goals**: S-curve path renders in ≤ 1 s; tab switch ≤ 300 ms INP
**Constraints**: No new dependencies; no new backend endpoints; stub module data where real content doesn't exist
**Scale/Scope**: Modifies 5 existing KMP files + 3 new files (LearnStore, LearnScreen, PlaceholderLessonScreen)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate                          | Status | Notes                                                                                     |
|-------------------------------|--------|-------------------------------------------------------------------------------------------|
| Test-first for new features   | PASS   | `LearnStoreTest` and updated `OnboardingStoreTest`/`ChangeTrackStoreTest` written before impl |
| Unit tests for business logic | PASS   | LearnStore load, tap-active, tap-locked, tap-completed all have test cases                |
| Design system adherence       | PASS   | Star node uses `MaterialTheme.colorScheme` tokens; no raw hex; subject colours from existing `SubjectTrackColors` |
| Consistent feedback patterns  | PASS   | Loading/error/empty states follow HomeStore/ProfileStore pattern                          |
| Mobile-first                  | PASS   | Android emulator is primary test target; path layout designed for phone-width screens     |
| No premature abstraction      | PASS   | `LearnStore` standalone; no shared base with OnboardingStore; stub data inline            |
| Single responsibility         | PASS   | `LearnStore` owns only active-track + module state; `StarNode` composable handles only node rendering |
| No dead code                  | PASS   | `selectedTrackIds` field removed from both stores; old `ToggleTrack` intent removed       |

**No violations to justify.**

## Project Structure

### Documentation (this feature)

```text
specs/006-subject-track-path/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/
│   └── ui-contracts.md  # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (affected paths)

```text
mobile/shared/src/commonMain/kotlin/com/preuni/shared/
├── data/auth/
│   └── TokenStore.kt                         # ADD: getActiveTrackId / setActiveTrackId / clearActiveTrackId
├── presentation/
│   ├── learn/                                # NEW package
│   │   ├── LearnStore.kt                     # NEW — State(activeTrackId, modules, isLoading, error)
│   │   ├── LearnScreen.kt                    # NEW — S-curve path + StarNode composable
│   │   └── PlaceholderLessonScreen.kt        # NEW — minimal "coming soon" screen
│   ├── onboarding/
│   │   ├── OnboardingStore.kt                # MODIFIED: Set<String> → String?; ToggleTrack → SelectTrack; fire-and-forget completion
│   │   └── TrackSelectionScreen.kt           # MODIFIED: chip single-select UX
│   ├── profile/
│   │   └── ChangeTrackStore.kt               # MODIFIED: Set<String> → String?; ToggleTrack → SelectTrack
│   └── main/
│       └── MainComponent.kt                  # MODIFIED: expose selectTab(BottomTab) publicly; default tab = LEARN
│
mobile/shared/src/commonTest/kotlin/com/preuni/shared/
├── presentation/learn/
│   └── LearnStoreTest.kt                     # NEW — 4 test cases
├── presentation/onboarding/
│   └── OnboardingStoreTest.kt                # MODIFIED — single-select assertions
└── presentation/profile/
    └── ChangeTrackStoreTest.kt               # MODIFIED — single-select assertions
```
