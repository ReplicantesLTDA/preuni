# Feature Specification: Onboarding Flow and UX Fixes

**Feature Branch**: `005-onboarding-ux-fixes`
**Created**: 2026-04-07
**Status**: Draft
**Input**: User description: "welcome screen on first install before login; fix Home page error; allow track switching after onboarding"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Track Re-selection After Onboarding (Priority: P1)

A logged-in user wants to change the subject tracks they selected during onboarding. Currently, once onboarding is completed there is no way to revisit track selection — the user is permanently locked into their initial choice with no path back.

**Why this priority**: This is the most disruptive bug reported. Users who made a wrong choice during track selection are stuck. It blocks the core learning flow and creates frustration immediately after onboarding.

**Independent Test**: A user who has already completed onboarding can navigate to their Profile and find an option to change subject tracks, select different ones, save, and return to the Home screen with the updated selection reflected.

**Acceptance Scenarios**:

1. **Given** a logged-in user who has completed onboarding, **When** they open the Profile tab, **Then** they see an option to manage or change their subject tracks.
2. **Given** a user on the track management screen, **When** they toggle different tracks and confirm, **Then** their selection is saved and the app reflects the new tracks.
3. **Given** a user changing tracks, **When** they leave the screen without confirming, **Then** their previous selection is preserved unchanged.
4. **Given** a user on the track management screen, **When** they attempt to deselect all tracks and confirm, **Then** they receive a validation message and cannot save an empty selection.

---

### User Story 2 — Welcome Screen on First Install (Priority: P2)

A new user who has never opened the app sees a brief introduction explaining what PreUni is and what it does — before being asked to create an account or log in. This introduction should only appear once: on the very first session. Returning visitors (logged in or not) skip it entirely.

**Why this priority**: First impressions matter for user retention. A cold start that immediately shows a login form without context creates drop-off. However, it does not block the core learning loop for existing users, so it ranks below the track bug.

**Independent Test**: Clearing all app data (simulating a fresh install) and launching the app shows welcome/intro slides. Closing and reopening the app (without clearing data) goes directly to the login screen instead of repeating the intro.

**Acceptance Scenarios**:

1. **Given** a brand-new install with no prior session data, **When** the user opens the app, **Then** they see the welcome/intro flow before any login prompt.
2. **Given** a user who has already seen the welcome flow, **When** they reopen the app while logged out, **Then** they land directly on the login screen.
3. **Given** a user who has already seen the welcome flow, **When** they reopen the app while logged in, **Then** they land directly on the Home screen.
4. **Given** a user viewing the welcome slides, **When** they reach the last slide, **Then** they are taken to the login/register screen.
5. **Given** a user viewing the welcome slides, **When** they tap a "Skip" or "Entrar" shortcut, **Then** they go directly to the login/register screen without finishing all slides.

---

### User Story 3 — Home Page Loads Student Dashboard (Priority: P3)

The Home tab (Início) always shows "Algo deu errado, tente novamente" for every logged-in user. It should instead display the student's personalized dashboard: greeting with their name, study streak, total XP, a call-to-action to start studying, ENEM readiness score (when available), and their enrolled subject tracks.

**Why this priority**: The dashboard is important but users can still navigate to Learn, Simulate, and Profile tabs. The two issues above have more immediate impact on getting users into the app at all.

**Independent Test**: A registered, verified, and logged-in user opens the Home tab and sees their name, streak count, XP total, and the "Aprender agora" button — with no error message shown.

**Acceptance Scenarios**:

1. **Given** a logged-in user with a valid profile, **When** they open the Home tab, **Then** they see a personalized greeting with their display name.
2. **Given** a logged-in user, **When** the Home tab loads, **Then** their streak count and XP total are displayed.
3. **Given** a logged-in user with no study history yet, **When** the Home tab loads, **Then** the CTA says "Comece a estudar" and readiness score section is hidden.
4. **Given** a logged-in user whose profile fails to load, **When** the Home tab shows the error, **Then** a "Tentar novamente" button is visible and retrying eventually loads the dashboard.
5. **Given** a newly registered user whose profile is being created, **When** they reach the Home tab immediately after registration, **Then** the app shows a loading state and then the dashboard — not a persistent error.

---

### Edge Cases

- What happens if a user dismisses the welcome slides mid-flow? The "seen" flag should only be written after completion or explicit skip, so a partial view replays the intro.
- What if the user loses connectivity while loading the Home dashboard? The error state with retry button must appear; no blank screen.
- What if a user tries to save zero tracks during re-selection? Validation must block the save with a clear message.
- What if the student profile is created asynchronously (user service delay after registration)? The Home screen should retry automatically a reasonable number of times before showing the error state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: On the very first app launch (no prior session), the system MUST display welcome/intro slides before the login screen.
- **FR-002**: The system MUST persist a "welcome seen" flag locally so the intro is never shown again after the first completion or skip.
- **FR-003**: The "welcome seen" flag MUST be independent of login state — a logged-out returning user skips the intro.
- **FR-004**: Welcome slides MUST include a skip/shortcut action that takes the user directly to login/register.
- **FR-005**: The Profile screen MUST include an option to view and change enrolled subject tracks.
- **FR-006**: The track re-selection screen MUST prevent saving an empty track selection with a clear validation message.
- **FR-007**: Confirmed track changes MUST be persisted and reflected immediately in the app without requiring a full logout/login cycle.
- **FR-008**: The Home tab MUST load and display the student's profile (display name, streak, XP) on every successful session.
- **FR-009**: The Home tab MUST show a retry action when the profile fails to load.
- **FR-010**: The Home tab MUST show "Comece a estudar" for users with zero XP and "Continue estudando" for users with XP greater than zero.
- **FR-011**: The ENEM readiness score section on the Home tab MUST only appear when the student's readiness score is greater than zero.

### Key Entities

- **WelcomeSeenFlag**: A locally stored boolean indicating whether the current device has completed or skipped the welcome flow. Not tied to any user account.
- **StudentProfile**: The user's display name, streak count, total XP, readiness score, and enrolled track IDs — fetched from the backend on each Home tab load.
- **TrackSelection**: The set of subject track IDs the student has enrolled in, editable from the Profile screen after initial onboarding.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of first-time installs show the welcome flow before the login screen; 0% of returning sessions show it again.
- **SC-002**: A user who completed onboarding can reach the track re-selection screen in 2 taps or fewer from the Profile tab.
- **SC-003**: The Home tab loads student data within 3 seconds on a standard connection; the error state appears within 5 seconds when the backend is unreachable.
- **SC-004**: Zero users see the "Algo deu errado" error on the Home tab when the backend is healthy and the user profile exists.
- **SC-005**: Track changes made on the re-selection screen are reflected on the Home tab within 1 second of confirmation.

## Assumptions

- The "welcome seen" flag is stored on the device (not in the user's account), so clearing app data resets it — this is the intended behavior for simulating a fresh install in testing.
- The welcome slides are the existing intro/onboarding pages currently shown after login; they will be moved to pre-login without redesign (visual content stays the same).
- The Home page error is caused by the user service backend failing to return the student profile — either because the user service is down, the profile wasn't created during registration, or the profile endpoint requires the user service to be running alongside auth. Fixing the backend connectivity is a prerequisite handled separately; this spec covers the frontend resilience and retry behavior.
- Track re-selection will update the same backend endpoint used during initial onboarding.
- The bottom navigation tabs (Início, Aprender, Simular, Perfil) and their routing are not changing — only content within Home and Profile tabs is affected.
