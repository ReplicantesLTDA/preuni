# Feature Specification: Friendly UI Polish — Core Frontend

**Feature Branch**: `003-ui-polish`
**Created**: 2026-04-04
**Status**: Draft
**Input**: User description: "Friendly interface, pill-shaped buttons, rounded inputs, core frontend screens (login and basic pages) matching the target end-state visual experience."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Polished Auth Screens (Priority: P1)

A student opening the app for the first time sees a welcoming, visually cohesive login and registration experience. The screens feel modern and friendly — not generic. Buttons are pill-shaped, input fields are fully rounded, and the spacing feels comfortable and breathable. The brand identity is clear from the very first screen.

**Why this priority**: Auth screens are the first thing every user sees. A poor first impression increases drop-off before the user ever reaches the core value. This is the highest-leverage surface for perceived quality.

**Independent Test**: Open the web app to the login screen and verify the visual treatment without touching any other part of the app. Delivers standalone value — the auth flow looks intentional and brand-consistent.

**Acceptance Scenarios**:

1. **Given** the user opens the app, **When** the login screen loads, **Then** all action buttons are pill-shaped with generous horizontal padding, the primary CTA is visually dominant, and the screen has a distinctive welcoming character — not a plain default form.
2. **Given** the login screen is displayed, **When** the user focuses on any text input, **Then** the field has fully rounded corners, a visible but soft border or filled background, and clear label/placeholder text.
3. **Given** the registration screen is open, **When** the user views the form, **Then** all fields follow the same rounded input style, the password strength indicator is visually friendly, and the "Create account" button is pill-shaped and prominent.
4. **Given** the email verification screen is shown after registration, **When** the user views it, **Then** the OTP input is large and centered, the screen explains clearly what to do next, and the "Verify" button matches the pill style.

---

### User Story 2 — Welcoming Home / Dashboard (Priority: P2)

After logging in, a student lands on a home page that feels warm and motivating. It greets them by name, gives a clear sense of what to do next, and uses the same rounded visual language established in auth. The screen should feel like the app knows the student and wants to help them succeed at the ENEM.

**Why this priority**: The home page is where retention is built. A cold or cluttered dashboard makes users feel overwhelmed; a warm, clear one makes them want to return. This is the next most visible surface after auth.

**Independent Test**: Log in with a test account and verify the home screen in isolation. Delivers standalone value — logged-in users have a better first-session experience.

**Acceptance Scenarios**:

1. **Given** the user is logged in, **When** the home screen loads, **Then** the user sees a personalized greeting with their display name and a motivational message fitting an exam prep context.
2. **Given** the home screen is displayed, **When** the user views the page, **Then** all interactive elements (buttons, cards, chips) use consistent rounded/pill shapes matching the auth style.
3. **Given** the home screen has no prior study activity, **When** the user views it, **Then** a clear primary call-to-action (e.g., "Start studying") is visually prominent and easy to find.

---

### User Story 3 — Consistent Bottom Navigation (Priority: P3)

The bottom navigation bar is clean, icon-focused, and styled to match the rounded visual language. The active tab is clearly distinguished. Moving between tabs feels instant.

**Why this priority**: Navigation is the skeleton of the app. Inconsistent nav undermines every screen's experience. Lower priority than auth and home, but essential for overall cohesion.

**Independent Test**: Navigate across all bottom tabs and verify the visual treatment independently. Delivers standalone value — navigation feels finished and trustworthy.

**Acceptance Scenarios**:

1. **Given** the user is on the main app, **When** they view the bottom navigation, **Then** icons are clear, labels are legible, and the active tab is highlighted with the brand accent color.
2. **Given** the user taps a bottom tab, **When** the screen changes, **Then** the navigation bar remains stable — no flicker or layout shift.

---

### User Story 4 — Onboarding Flow (Priority: P4)

First-time users go through a brief onboarding after verifying their email. The screens use clear iconography, motivating copy, and the same pill-button style. The flow feels like a warm welcome, not a setup wizard.

**Why this priority**: Onboarding is seen only once per user but is critical for activation. The visual pattern is established by higher-priority screens; this story applies it consistently to the onboarding sequence.

**Independent Test**: Walk through onboarding with a new account. Delivers standalone value — new users are welcomed and oriented before reaching the dashboard.

**Acceptance Scenarios**:

1. **Given** a new user completes email verification, **When** they reach onboarding, **Then** each screen has a single focused message, a clear icon or illustration area, and a pill-shaped "Continue" button.
2. **Given** the user completes onboarding, **When** they reach the main app, **Then** the visual style is continuous — no jarring shift in button shapes, colors, or typography.

---

### Edge Cases

- What happens when the user's display name is very long? Text must truncate or wrap gracefully without breaking the layout.
- What happens on a narrow viewport (360px width)? Inputs and buttons must remain fully usable without clipping or horizontal scroll.
- What happens when a validation error appears? Error states must be styled consistently — accent color, inline helper text — without pushing other elements off-screen.
- What happens when the keyboard appears on mobile? Inputs must remain visible and accessible.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All interactive buttons across auth, home, onboarding, and navigation screens MUST use a pill or high-radius rounded shape — no square corners.
- **FR-002**: All text input fields MUST use fully rounded corners with a visible but subtle border or filled background, plus distinct visual states for default, focused, error, and disabled.
- **FR-003**: The login screen MUST present a clear hierarchy: brand/logo area at top, form centered, secondary actions (sign in with code, create account) below the primary CTA.
- **FR-004**: The registration screen MUST present all fields in a comfortable, scrollable layout with adequate vertical spacing.
- **FR-005**: The email verification screen MUST center the OTP input prominently and explain the required action clearly.
- **FR-006**: The home screen MUST display the logged-in user's display name in a personalized greeting.
- **FR-007**: All screens MUST use a consistent color palette anchored to the app's primary violet/purple brand color.
- **FR-008**: All screens MUST maintain a consistent typography scale: one prominent heading, supporting body text, and small helper/error text.
- **FR-009**: Validation error messages MUST appear inline beside the relevant field without shifting surrounding layout.
- **FR-010**: The bottom navigation bar MUST clearly highlight the active tab using the brand accent color.
- **FR-011**: All screens MUST be fully usable at 360px–430px viewport widths with no horizontal scrolling or clipped elements.
- **FR-012**: Each onboarding screen MUST present a single focused message and a pill-shaped primary action.

### Key Entities

- **Design Token**: A named value for color, border-radius, spacing, or typography applied consistently across all screens.
- **Primary CTA**: The main action button on any screen — always visually dominant and pill-shaped.
- **Input Field**: Any text entry control — fully rounded, with distinct visual states for default, focused, error, and disabled.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A first-time user can identify the primary action on any auth screen within 3 seconds of viewing it — no ambiguity about where to tap.
- **SC-002**: 100% of interactive buttons across auth, home, and onboarding screens are pill or high-radius rounded — zero square-cornered buttons remain.
- **SC-003**: 100% of text input fields use rounded corners — no default rectangular inputs remain.
- **SC-004**: Any two screens from the app share the same visual language — a reviewer can confirm they belong to the same product without prior context.
- **SC-005**: All screens remain fully functional and correctly laid out at 360px width without horizontal scroll.
- **SC-006**: Error messages appear inline next to the relevant field without causing layout shifts for other elements.

---

## Assumptions

- The existing violet/purple Material Design 3 color palette is kept as the brand foundation — only component shapes, spacing, and copy are refined.
- Dark mode is out of scope for this iteration; light mode only.
- Custom illustrations or original artwork are out of scope; iconography, color, and typography create the visual character.
- The Kotlin/Wasm web target is the primary reference for visual polish in this iteration; Android and iOS inherit the improvements through shared Compose components.
- Custom font loading is out of scope; the Material 3 default type scale is used.
- Motion design and animated transitions beyond default Compose behavior are out of scope.
- The existing screen structure (which screens exist and what data they show) is not changed — only the visual presentation is upgraded.
