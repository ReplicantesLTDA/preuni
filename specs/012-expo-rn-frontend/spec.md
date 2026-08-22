# Feature Specification: Migrate frontend to React Native + Expo (wireframe-driven redesign)

**Feature Branch**: `012-expo-rn-frontend`
**Created**: 2026-05-26
**Status**: Draft
**Input**: User description: "Migrate frontend from Kotlin Multiplatform to React Native + TypeScript + Expo (Expo Go for fast visual/tactile testing). Re-implement the existing product following `wireframe.html` for layout, colors, fonts and overall design feel. Where the wireframe omits a flow that currently exists in the KMP app, adapt that flow into the wireframe's visual language. Reason for the switch: stronger AI/agent tooling support on the RN/TS stack."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Student uses the redesigned core learning loop on a phone (Priority: P1)

A returning student opens the app on their phone, sees the Trilha home with their streak, XP and recommended next step, taps into the next activity, completes it, and returns to a clearly updated Trilha. The look and feel matches `wireframe.html` (colors, typography, spacing, mascot, status bar).

**Why this priority**: This is the daily-use loop. Without it the rewrite delivers no value, regardless of how many secondary screens ship.

**Independent Test**: Install the new app on a physical phone via Expo Go, sign in as an existing test student, complete one recommended activity from Trilha, observe the Trilha home reflect the change (XP/streak/next step). No KMP build involved.

**Acceptance Scenarios**:

1. **Given** a signed-in student with existing progress, **When** they open the app, **Then** the Trilha home renders with their current streak, XP and a recommended next activity, styled per the wireframe.
2. **Given** the student finishes an activity, **When** they return to Trilha, **Then** XP, streak and the recommended next step update without requiring a manual refresh.
3. **Given** the device is offline, **When** the student opens the app, **Then** they see a clear offline state instead of a blank or broken screen.

---

### User Story 2 - New student signs up, verifies email, and finishes onboarding (Priority: P1)

A first-time visitor registers with email + password, receives an OTP, verifies their email (with the option to resend), completes onboarding, and lands on Trilha as a logged-in user.

**Why this priority**: Without auth + onboarding the app cannot acquire new users, even if the core loop works for existing accounts.

**Independent Test**: From a clean install, register a new email, receive the OTP via the backend's existing mail path, verify, finish onboarding, land on Trilha.

**Acceptance Scenarios**:

1. **Given** a new visitor, **When** they submit a valid registration, **Then** an OTP is sent and the verify-email screen appears.
2. **Given** the visitor did not receive the OTP, **When** they tap "resend", **Then** a new OTP is sent and they are not forced to re-register.
3. **Given** a verified new user, **When** they finish onboarding, **Then** they land on Trilha already authenticated (no extra login step).
4. **Given** invalid credentials at login, **When** the user submits, **Then** the form surfaces a clear, localized error styled per the wireframe.

---

### User Story 3 - Student manages their profile (Priority: P2)

A signed-in student opens Perfil, edits their display name, changes their avatar, changes email (re-verifies), changes password, and signs out.

**Why this priority**: Required for retention and account hygiene; not strictly needed for first daily use but required to consider the rewrite at parity with the current app.

**Independent Test**: Sign in, change each field independently, sign out, sign in with new credentials.

**Acceptance Scenarios**:

1. **Given** a signed-in student, **When** they update display name, **Then** the new name is visible across the app on next view.
2. **Given** an avatar change, **When** the upload completes, **Then** the new avatar appears in the top status bar and Perfil.
3. **Given** an email change, **When** the user submits, **Then** a verification flow is triggered for the new address and the old email keeps working until verification succeeds.
4. **Given** a successful sign-out, **When** the app is reopened, **Then** the user lands on the unauthenticated entry, not on Trilha.

---

### User Story 4 - Student uses Redação and Simulado tabs (Priority: P2)

A student taps the Redação and Simulado tabs from the bottom navigation, sees their respective home screens (history, prompts, simulation entry points) rendered with the wireframe's visual language.

**Why this priority**: These are first-class tabs in the wireframe. They must exist visually at launch; deep functionality can iterate after the rewrite.

**Independent Test**: From Trilha, tap each tab, confirm a non-placeholder screen with the wireframe styling, navigate back.

**Acceptance Scenarios**:

1. **Given** a signed-in student on Trilha, **When** they tap Redação, **Then** the Redação home renders with the wireframe styling and any data the backend already exposes.
2. **Given** Simulado has no data yet for this user, **When** they open the tab, **Then** an empty state styled per the wireframe is shown, not a blank screen.

---

### User Story 5 - Same build runs on iOS, Android and Web (Priority: P3)

The same codebase produces working builds for iOS, Android, and Web. A reviewer can open the app on each platform and complete the P1 stories.

**Why this priority**: Multiplatform is a stated goal but not blocking for the first internal demo. iOS + Android are higher priority than Web.

**Independent Test**: Run the project on each of iOS (on a device), Android (on a device), and Web (browser). Complete US1 on each.

**Acceptance Scenarios**:

1. **Given** the project's run command for each platform, **When** executed, **Then** the app boots and the Trilha home renders without platform-specific crashes.
2. **Given** a layout reviewed at ~360px width on Web, **When** rendered, **Then** the wireframe layout holds without clipping.

---

### Edge Cases

- App opened with an expired refresh token: user is routed to login without a crash; no infinite refresh loop.
- OTP resend tapped repeatedly: client and/or backend rate-limit feedback is shown clearly, styled per the wireframe.
- Avatar upload fails mid-way: previous avatar is preserved; user sees an actionable error.
- Backend returns the `{ error: { code, field, message } }` envelope: every screen surfaces a human-readable message, never raw JSON.
- Web build at narrow widths (≤360px): no horizontal scroll, no clipped top-status-bar metrics.
- Device locale is not pt-BR: app still renders (copy may fall back to pt-BR for v1).
- Slow / flaky network: long-running calls show a loading state and time out gracefully.

## Requirements *(mandatory)*

### Functional Requirements

**Visual + design system**

- **FR-001**: The app MUST adopt the colors, typography, spacing and overall design language defined by `wireframe.html` as the source of truth.
- **FR-002**: A reusable top status bar MUST display streak, XP and avatar consistent with the wireframe across the authenticated screens that show it.
- **FR-003**: A reusable bottom navigation MUST expose the primary destinations shown in the wireframe (Trilha, Redação, Simulado, Perfil).
- **FR-004**: Empty states, mascot placeholders, primary CTAs and form controls MUST follow the wireframe's component shapes (rounded corners, elevation, color usage).

**Auth + onboarding**

- **FR-005**: Users MUST be able to register with email + password, receive an OTP, verify their email, and request a resend of the OTP.
- **FR-006**: Users MUST be able to log in with email + password, log out, and have their session restored on app relaunch while the refresh token is valid.
- **FR-007**: Verified users MUST be routed to onboarding (if not completed) or to Trilha (if completed) without an extra manual login step after verification.
- **FR-008**: Users MUST be able to reset their password using the existing backend reset flow.

**Core learning loop**

- **FR-009**: Trilha home MUST render the signed-in student's streak, XP, readiness score, and a recommended next activity.
- **FR-010**: Completing an activity MUST update Trilha state visibly (XP, streak, next step) without requiring a manual refresh.
- **FR-011**: Where the wireframe does not depict a flow that the current KMP app supports (e.g. subject track path, FSRS review queues), the rewrite MUST adapt that flow into the wireframe's visual language rather than dropping it.

**Profile**

- **FR-012**: Users MUST be able to view and update their display name, avatar, email (with re-verification), and password from Perfil.
- **FR-013**: Sign-out MUST clear local session state and route the user to the unauthenticated entry.

**Secondary tabs**

- **FR-014**: Redação and Simulado MUST each render a tab home consistent with the wireframe, surfacing whatever backend data already exists and using styled empty states otherwise.

**Platform + delivery**

- **FR-015**: A single frontend codebase MUST build for iOS, Android, and Web.
- **FR-016**: A developer or reviewer MUST be able to launch the app on a real iOS or Android device using Expo Go (or its equivalent dev client) without producing native binaries.
- **FR-017**: The previous Kotlin Multiplatform frontend (`mobile/shared`, `mobile/androidApp`, `mobile/iosApp`, `mobile/webApp`) MUST be removed once the new app reaches P1 + P2 story parity. Until then, both may coexist, but only the new app is shipped to users.
- **FR-018**: The new frontend MUST consume the existing backend monolith without requiring backend API changes for parity behavior.

**Error handling + resilience**

- **FR-019**: The client MUST parse the backend's `{ error: { code, field, message } }` envelope and surface user-readable messages across every screen.
- **FR-020**: Expired or invalid refresh tokens MUST route the user to login without crashing or looping.

**Localization**

- **FR-021**: Copy MUST default to pt-BR, matching the current KMP app. Additional locales are out of scope for v1.

### Key Entities

This feature does not introduce new persistent entities. It re-presents the existing student, session, track, redação, simulado and profile data already owned by the backend monolith. Local-only state remains limited to session tokens and minor UI flags (e.g., welcome-seen, active track id) carried over from the current app.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new student can go from app open → registered → email-verified → onboarded → Trilha in under 3 minutes on a typical 4G connection.
- **SC-002**: 100% of P1 + P2 user stories are demonstrable on iOS, Android and Web from a single codebase before the old KMP frontend is removed.
- **SC-003**: An agent (or developer) can ship a new screen end-to-end (route → UI → data → tests) in under one focused day, validated by at least one screen built this way during the rewrite.
- **SC-004**: 95% of Trilha home loads complete within 2 seconds on a typical 4G connection with a warm session.
- **SC-005**: Crash-free session rate on iOS + Android is at least 99% during the first internal-testing week.
- **SC-006**: A reviewer running the project on Expo Go sees a UI change within 10 seconds of saving a file (hot reload).
- **SC-007**: Zero regressions in available account/profile operations vs. the current KMP app once US3 ships (display name, avatar, email change with verification, password change, sign-out).

## Assumptions

- The Go monolith backend (auth + users + mail in-process, plus existing endpoints) remains the source of truth. No backend changes are required for parity behavior; any additions are out of scope for this spec.
- Existing API contracts (login, register, verify, resend, refresh, profile, avatar PUT, error envelope) are stable and already exercised by the current KMP client.
- The new frontend lives in a fresh directory (working name: `mobile/` or similar) at the repo root. The existing `mobile/` KMP tree is removed once P1 + P2 are demonstrably at parity.
- Tech-stack decisions captured for the planning phase (not part of the user-facing spec): React Native + TypeScript on Expo; Expo Router for navigation; Zustand for client state; TanStack Query for server cache; Zod for schema validation. These are decisions, not open questions.
- `wireframe.html` (at the repo root) is the visual contract; where it is silent, the implementer adapts current KMP behavior into the wireframe's design language rather than inventing new visual patterns.
- Copy stays pt-BR for v1, matching the current app.
- Push notifications, deep-linking beyond auth callbacks, and offline write queues are out of scope for v1.
- Distribution channel for early testing is Expo Go on real devices + a web build for desktop review; store submission (App Store / Play Store / production web hosting) is a follow-up effort.
