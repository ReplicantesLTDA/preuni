# Feature Specification: Profile Management Fixes & Backend Integration

**Feature Branch**: `007-profile-mgmt-fixes`  
**Created**: 2026-04-08  
**Status**: Draft  
**Input**: User description: "Profile section (Perfil) management actions (change username, email, password, delete account) have no backend communication. Sub-pages within Perfil retain state when switching tabs. No back navigation from sub-pages. Back buttons violate device safe zones. Logout button is missing from Perfil."

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Logout from Profile (Priority: P1)

A logged-in user wants to sign out of the app from the Profile (Perfil) tab. Currently there is no logout option available in the Profile section, making it impossible to log out without uninstalling the app.

**Why this priority**: Logout is a fundamental authentication flow and has been explicitly marked as urgent by the product owner.

**Independent Test**: Can be tested end-to-end by navigating to the Perfil tab, tapping the Logout button, and verifying the user is signed out and redirected to the login/welcome screen.

**Acceptance Scenarios**:

1. **Given** a logged-in user is on the Perfil tab, **When** they tap the Logout button, **Then** their session is terminated and they are redirected to the unauthenticated welcome/login screen.
2. **Given** a logged-in user is on any Profile sub-page, **When** they navigate to the main Perfil page and tap Logout, **Then** the same sign-out flow completes successfully.
3. **Given** a logout request is in progress, **When** the network is temporarily unavailable, **Then** the local session tokens are cleared and the user is still redirected to the login screen (offline-safe logout).

---

### User Story 2 — Profile Sub-page State Resets on Tab Switch (Priority: P1)

When a user navigates into a Profile sub-page (e.g., Change Username), switches to another main tab (e.g., Aprender), and then returns to the Perfil tab, they should always land on the main Profile overview — not the sub-page they were previously on.

**Why this priority**: The current behavior traps users in sub-pages with no way to exit, breaking basic navigation expectations and making the Profile section nearly unusable.

**Independent Test**: Can be tested by opening Change Username, switching to the Aprender tab, returning to Perfil, and confirming the main Profile overview is shown.

**Acceptance Scenarios**:

1. **Given** a user has opened a Profile sub-page (e.g., Change Username), **When** they switch to any other main tab, **Then** the Profile tab's navigation state is reset to the main Profile overview.
2. **Given** the Profile tab has been reset, **When** the user returns to the Perfil tab, **Then** they see the main Profile page, not the previously visited sub-page.
3. **Given** a user is mid-form on a Profile sub-page (e.g., partially typed a new username), **When** they switch tabs and return, **Then** the form is cleared and the main Profile overview is shown (no stale input retained).

---

### User Story 3 — Back Navigation Within Profile Sub-pages (Priority: P2)

From any Profile sub-page (Change Username, Change Email, Change Password, Delete Account), the user must be able to navigate back to the main Profile overview using a visible back button or gesture.

**Why this priority**: Without a back button, users are stuck unless they switch tabs (which itself is broken per US2), making sub-pages dead ends.

**Independent Test**: Can be tested by entering any Profile sub-page and confirming a working back navigation control exists that returns the user to the main Profile overview.

**Acceptance Scenarios**:

1. **Given** a user is on a Profile sub-page, **When** they tap the back arrow or use the back gesture, **Then** they are navigated to the main Profile overview.
2. **Given** a user is on the Change Password sub-page with unsaved changes, **When** they tap back, **Then** they return to the Profile overview (changes are discarded without a confirmation dialog).
3. **Given** any Profile sub-page, **When** the back control is rendered, **Then** it is positioned within the device's safe zone (below the status bar and camera area) and is fully tappable.

---

### User Story 4 — Safe Zone Compliance for All Navigation Controls (Priority: P2)

Back buttons, close icons, and other navigation controls across all screens (including Login with Email and all Profile sub-pages) must be rendered within the device's interactive safe zone — they must never overlap the status bar, notch, camera cutout, or Dynamic Island area.

**Why this priority**: Controls in unsafe areas are physically unreachable on modern devices, making screens inaccessible to users with notched or punch-hole displays.

**Independent Test**: Can be tested on a device with a notch/punch-hole by verifying every navigation control is fully visible and tappable below the system status bar on all affected screens.

**Acceptance Scenarios**:

1. **Given** the Login with Email screen, **When** viewed on any device (including those with notches, punch-holes, or Dynamic Island), **Then** the back arrow is fully visible and tappable within the safe area.
2. **Given** any Profile sub-page, **When** the back button is rendered, **Then** it respects top inset padding and is not obscured by device hardware or OS overlays.
3. **Given** any screen in the app that has a top navigation control, **When** rendered on a device with a non-zero top inset, **Then** the control is offset below the inset value.

---

### User Story 5 — Profile Actions Connected to Backend (Priority: P3)

Users can successfully change their username, email address, and password — and delete their account — from the Perfil section. Each action communicates with the backend, persists the change, and provides clear feedback on success or failure.

**Why this priority**: Core account management functionality; deferred slightly behind navigation fixes because the forms are inaccessible until US2/US3 are resolved.

**Independent Test**: Can be tested per action individually — e.g., change username, confirm the new name appears on the Profile page and persists after app restart.

**Acceptance Scenarios**:

1. **Given** a user submits a new username, **When** the request succeeds, **Then** the updated username is reflected on the Profile page immediately and persists across app restarts.
2. **Given** a user submits a new email address, **When** the request succeeds, **Then** the new email is shown on the Profile page; if email verification is required by the backend, the user sees an appropriate notification.
3. **Given** a user submits a password change with an incorrect current password, **When** the request is processed, **Then** a clear error is displayed and no change is made.
4. **Given** a user submits a password change with valid current and new passwords, **When** the request succeeds, **Then** the user sees a success confirmation.
5. **Given** a user initiates account deletion, **When** they confirm the irreversible action via a confirmation prompt, **Then** their account is deleted, the session is terminated, and they are redirected to the welcome/login screen.
6. **Given** any profile management action, **When** the device has no network connectivity, **Then** an appropriate error message is displayed and no partial state change occurs.

---

### Edge Cases

- What happens when the user taps Logout while a profile update request is in flight?
- How does the system handle a duplicate username or email that already exists for another account?
- What if the session token expires mid-form during a profile update?
- How does account deletion behave if the backend request times out — is the account deleted or not?
- What if a user switches tabs very quickly while a Profile sub-page is loading?

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Profile (Perfil) section MUST display a Logout button that terminates the user's session and redirects them to the unauthenticated welcome/login screen.
- **FR-002**: Logout MUST clear all local session credentials, even when the network is unavailable.
- **FR-003**: The Profile tab MUST reset its navigation stack to the main Profile overview whenever the user navigates away from the Perfil tab to another main tab.
- **FR-004**: All Profile sub-pages (Change Username, Change Email, Change Password, Delete Account) MUST include a back navigation control that returns the user to the main Profile overview.
- **FR-005**: All back navigation controls and top-of-screen interactive elements across the entire app MUST be rendered within the device's safe area insets (respecting status bar, notch, camera cutout, and Dynamic Island).
- **FR-006**: The Change Username action MUST send the updated username to the backend and reflect the confirmed value in the Profile UI on success.
- **FR-007**: The Change Email action MUST send the updated email to the backend and reflect the confirmed value in the Profile UI on success.
- **FR-008**: The Change Password action MUST require the user's current password; an incorrect current password MUST produce a descriptive error and leave the account unchanged.
- **FR-009**: The Delete Account action MUST present a confirmation prompt before submitting the deletion request; on success, the session MUST be terminated and the user redirected to the welcome/login screen.
- **FR-010**: All profile management actions MUST display a loading indicator during the backend request and a clear success or error message upon completion.
- **FR-011**: Profile sub-pages MUST NOT retain form input (text, selections) when the user navigates away and returns.

### Key Entities

- **User Profile**: The authenticated user's account information — username, email address, and password (write-only for change operations).
- **Session / Auth Token**: Local credentials authenticating the user; must be fully cleared on logout or account deletion.
- **Profile Sub-page**: A secondary screen accessible only from within the Perfil tab (Change Username, Change Email, Change Password, Delete Account); navigation stack resets when the user leaves the Perfil tab.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A logged-in user can complete the logout flow in under 5 seconds from tapping Logout to reaching the welcome/login screen.
- **SC-002**: 100% of Profile sub-pages return the user to the main Profile overview when the back control is tapped.
- **SC-003**: 100% of navigation controls across all app screens are fully tappable (not obscured by system UI) on devices with notch, punch-hole, or Dynamic Island form factors.
- **SC-004**: Returning to the Perfil tab after visiting another tab always displays the main Profile overview — never a previously visited sub-page.
- **SC-005**: Each profile management action (change username, email, password, delete account) produces a visible success or error state within 10 seconds of submission under normal network conditions.
- **SC-006**: Zero instances of users being unable to access the main Profile overview due to stuck navigation state after this fix is deployed.

---

## Assumptions

- The backend already exposes endpoints for updating username, email, and password, and for deleting accounts; this feature connects the existing mobile UI to those endpoints.
- Email change may require email verification on the backend side; if so, the app will display a notification prompt but does not need to implement the verification email flow itself.
- Account deletion is permanent and handled server-side; no undo or grace period is required at the mobile level.
- Safe zone handling will use the platform's standard inset APIs — no custom per-device dimension overrides are needed.
- Logout does not require a successful server-side token revocation call to complete; clearing local credentials is sufficient for the mobile session to end.
- The Profile tab navigation reset applies only when switching between the main bottom-navigation tabs; navigating within the Perfil tab itself (going deeper into sub-pages) does not reset state prematurely.
- The Change Track (subject track re-selection) action already in the Profile section is not part of this feature's scope and should remain unaffected.
