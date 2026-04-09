# Feature Specification: Subject Track Path

**Feature Branch**: `006-subject-track-path`
**Created**: 2026-04-08
**Status**: Draft
**Input**: User description: "Single subject selection → gamified S-curve learning path; fix Começar button; Android-first"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Fix Subject Selection and Navigate to Track (Priority: P1)

After completing registration and email verification, the user lands on the onboarding screen where they pick one subject (e.g. "Matemática"). Tapping "Começar" should take them to that subject's learning track. Currently the button does nothing.

**Why this priority**: This is a blocker — no user can proceed past onboarding. Every new account hits this dead end immediately after signup.

**Independent Test**: A newly registered user can select exactly one subject, tap "Começar", and arrive at the subject's learning track screen. The onboarding screen is not shown again on the next launch.

**Acceptance Scenarios**:

1. **Given** a user on the subject-selection screen with no subject chosen, **When** they tap "Começar", **Then** the button is disabled and nothing happens.
2. **Given** a user who has tapped one subject chip, **When** they tap "Começar", **Then** they are taken to that subject's track screen.
3. **Given** a user who already completed onboarding, **When** they relaunch the app, **Then** they land directly on the track screen (not onboarding).
4. **Given** a user who tapped Subject A then Subject B, **When** they tap "Começar", **Then** only Subject B is selected (single-selection enforced).

---

### User Story 2 — View the Gamified Learning Path (Priority: P1)

Once a subject is selected the user sees a scrollable screen showing their learning track: a vertical sequence of star-shaped module buttons arranged in an S-curve zigzag pattern (first button top-right → snake down → last button bottom-left). The path is rendered even when no real content exists yet — placeholder module nodes are shown.

**Why this priority**: This is the core visual identity of the app's gamification. Without it the track screen has nothing to show and the feature has no value.

**Independent Test**: Navigate to any subject's track screen and scroll through a visible S-curve path of at least 10 module nodes. The path renders without backend data.

**Acceptance Scenarios**:

1. **Given** a user on the track screen, **When** the screen loads, **Then** they see a vertically scrollable path of star-shaped buttons arranged in a zigzag S pattern.
2. **Given** the first module node, **When** the user looks at it, **Then** it appears at the top-right of the path and is visually highlighted as the starting point.
3. **Given** subsequent nodes, **When** the user scrolls down, **Then** nodes alternate left–right forming a continuous S-shape with visible connecting lines or curves between them.
4. **Given** a module node the user has not yet reached, **When** they view it, **Then** it appears locked/greyed out compared to unlocked ones.
5. **Given** the current active module, **When** the user taps its star button, **Then** a placeholder lesson screen or "coming soon" message is shown.
6. **Given** a completed module, **When** the user views it, **Then** it is visually distinct (e.g. filled star, different colour) from locked and active nodes.

---

### User Story 3 — Switch Subject from Profile (Priority: P2)

A user who is tired of their current subject can go to their Profile, tap "Alterar matérias", pick a different single subject, and be taken to that subject's track. Their previous subject's progress is preserved — switching does not erase it.

**Why this priority**: Retention depends on users not feeling locked in. This builds on US1/US2 and the existing "Alterar matérias" flow from feature 005.

**Independent Test**: A user currently on the Matemática track goes to Profile → Alterar matérias → selects Português → is taken to the Português track screen.

**Acceptance Scenarios**:

1. **Given** a logged-in user with a subject already chosen, **When** they open Profile and tap "Alterar matérias", **Then** the current subject is pre-selected.
2. **Given** a user on the subject-change screen, **When** they select a new subject and save, **Then** they are taken to the new subject's track screen.
3. **Given** a user who switches subjects, **When** they switch back later, **Then** their progress on the original subject is still intact.

---

### User Story 4 — Bottom Navigation Between Sections (Priority: P2)

The app has a bottom navigation bar with tabs for Track (Aprender), Home (Início), and Profile (Perfil). The user can tap any tab to switch sections without losing state.

**Why this priority**: Navigation scaffolding is needed so users can reach Profile to switch subjects and return to their track without losing context.

**Independent Test**: From the track screen, tap Profile tab, then tap the Track tab — user returns to the track screen at the same scroll position.

**Acceptance Scenarios**:

1. **Given** a user on the track screen, **When** they tap the Profile tab, **Then** the Profile screen appears without restarting the track.
2. **Given** a user on Profile, **When** they tap the Track tab, **Then** they return to the track at their previous scroll position.
3. **Given** a user with no subject yet chosen, **When** the app launches, **Then** the Track tab triggers the subject-selection flow before showing the path.

---

### Edge Cases

- What if the content API returns zero modules for a subject? Show the path with placeholder nodes — do not show an empty screen.
- What if a user somehow has multiple subjects stored (legacy from the old multi-select flow)? Treat the first stored subject as the active one.
- What if the user taps a locked module node? Show a friendly "Complete the previous modules first" message — no navigation occurs.
- What if the device is offline when the track screen loads? Show cached data if available; otherwise show an offline notice with a retry button.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The subject-selection screen MUST enforce single-selection — tapping a second chip deselects the first.
- **FR-002**: The "Começar" button MUST be disabled until exactly one subject is selected.
- **FR-003**: Tapping "Começar" with one subject selected MUST navigate to that subject's track screen and mark onboarding as complete.
- **FR-004**: The track screen MUST display module nodes in a vertically scrollable S-curve zigzag layout.
- **FR-005**: Module nodes MUST be star-shaped and visually distinguish between: locked, active (current), and completed states.
- **FR-006**: The path MUST render with placeholder/stub nodes when no real module content exists from the backend.
- **FR-007**: Connecting lines or curves MUST visually join adjacent module nodes along the S-curve path.
- **FR-008**: Tapping an active module node MUST open a placeholder lesson or "coming soon" screen.
- **FR-009**: Tapping a locked module node MUST show a non-navigating feedback message.
- **FR-010**: The bottom navigation MUST include at minimum: Track, Home, and Profile tabs.
- **FR-011**: Switching subjects from Profile MUST update the active track without erasing progress on the previous subject.
- **FR-012**: The track screen MUST preserve scroll position when the user navigates away and returns via the bottom tab.

### Key Entities

- **SubjectTrack**: The learning path for one subject. Contains an ordered list of module nodes. One track is active at a time per user.
- **ModuleNode**: A single step on the track. Has a state: `locked`, `active`, or `completed`. Contains a title and optional lesson content reference.
- **ActiveSubject**: The single subject currently selected by the user. Persisted locally and synced to the backend.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A newly registered user can select a subject and reach the track screen in 2 taps or fewer.
- **SC-002**: The track screen renders a visible S-curve path within 1 second of navigation, even with no backend content available.
- **SC-003**: 100% of module nodes on the path are reachable by scrolling — no nodes are cut off or hidden off-screen.
- **SC-004**: Switching between bottom-navigation tabs takes under 300 ms with no visible blank screen.
- **SC-005**: Switching subjects in Profile takes the user to the new track within 1 second of confirming the selection.

## Assumptions

- Single subject selection replaces the previous multi-select model; only one subject is active at a time per user.
- The backend content endpoints for modules may not yet return real data — the UI must degrade gracefully with stub/placeholder nodes.
- The S-curve layout is a fixed visual pattern: nodes alternate left and right as they progress downward, one node per row.
- Bottom navigation already exists in the app (Início, Aprender, Simular, Perfil) — this feature reuses and may reorder tabs to prioritise the track.
- Android is the primary target platform for this feature; iOS and web will follow but are not in scope.
- Per-module progress state is locally tracked for now; backend sync of progress is out of scope for this feature.
- The "Começar" button fix in onboarding and the track screen are tightly coupled and delivered together.
