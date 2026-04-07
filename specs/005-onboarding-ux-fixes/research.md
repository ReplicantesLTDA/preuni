# Research: Onboarding Flow and UX Fixes

## Decision 1: Where to store the "welcome seen" flag

**Decision**: Use the existing `SecureStorage` expect/actual class with a new `KEY_WELCOME_SEEN = "welcome_seen"` key.

**Rationale**: `SecureStorage` is already wired into all three platforms (Android Keystore, iOS Keychain, web `sessionStorage`, desktop in-memory). Adding one key avoids any new storage mechanism, new dependency, or new platform actual. The flag is device-local (not account-linked), which matches the spec requirement.

**Alternatives considered**:
- `DataStore Preferences` (KMP): heavier dependency, overkill for a single boolean.
- Separate `SharedPreferences`/`UserDefaults` wrapper: would require new expect/actual implementations on all platforms.
- Storing the flag in the user's backend profile: contradicts the spec (flag must be independent of login state).

---

## Decision 2: How to split the onboarding screen

**Decision**: Extract pages 0–2 ("Bem-vindo", "Aprendizado por repetição", "Simule o ENEM") into a new `WelcomeComponent` / `WelcomeScreen` / `WelcomeStore`. Keep `OnboardingComponent` for page 3 (track selection) only, reducing `TOTAL_PAGES` to 1.

**Rationale**: The current `OnboardingScreen` has a hard-coded `if (page < 3)` branch that renders either welcome copy or the track-selection page. Separating them gives each a single responsibility, removes the branching composable, and allows the welcome flow to be gated pre-login while track selection stays post-login. No visual redesign is required — the same three `OnboardingPage` data objects and pager UI move verbatim to `WelcomeScreen`.

**Alternatives considered**:
- Parameterise `OnboardingScreen` with a `startPage` argument: Would leave dead navigation intents (PreviousPage/NextPage) in `OnboardingStore` when used for track selection only. Fragile.
- Keep a single 4-page flow with a login gate mid-flow: Complex, breaks back-button expectations, and violates single-responsibility.

---

## Decision 3: Routing — where to add Config.Welcome

**Decision**: Add `Config.Welcome` to `RootComponent` sealed interface. New `initialConfig` logic:
```
if (!welcomeSeen)  → Config.Welcome
else if (!loggedIn) → Config.Auth
else if (!trackSelected) → Config.Onboarding   (stub check: tokenStore.load() != null, real check deferred)
else → Config.Main
```
After `WelcomeComponent` completes or is skipped, it navigates to `Config.Auth` (replaceAll).

**Rationale**: `RootComponent` is already the single source of truth for top-level navigation. `Config.Welcome` fits naturally alongside `Config.Auth`, `Config.Onboarding`, `Config.Main`. The existing `onboardingCompleted()` stub (always returns true when session exists) is sufficient for now — track selection is still shown post-login on first login via the `AuthComponent → Config.Onboarding` transition.

**Alternatives considered**:
- A separate `WelcomeRouter` outside `RootComponent`: introduces an extra navigation layer with no benefit.

---

## Decision 4: Track re-selection — backend endpoint

**Decision**: Reuse the existing `PATCH /v1/students/me/onboarding` endpoint (already registered in the user service). The request body `{"enrolled_track_ids": [...]}` already has the `EnrolledTrackIDs` field. A new `updateTracks(trackIds)` method is added to `UserApiClient`.

**Rationale**: The endpoint is idempotent, already authenticated, and semantically correct. The backend handler has a `TODO` comment for actually writing track_enrollment rows, but that is out of scope here — what matters is the UX flow compiles and the API call succeeds.

**Alternatives considered**:
- A dedicated `PATCH /v1/students/me/tracks` endpoint: requires a new backend handler, router registration, and migration — disproportionate to the frontend-only scope of this feature.

---

## Decision 5: HomeStore auto-retry

**Decision**: Add a `retryCount` field to `HomeStore.State` (internal only, not exposed in State). On failure, if `retryCount < 3`, dispatch a `Msg.RetryScheduled`, wait 1 second, and re-dispatch `load()`. Only emit `Msg.ErrorReceived` after the third failure.

**Rationale**: The most common failure mode is a brief delay in profile creation after registration (auth service creates the user profile asynchronously via the user service). A 3-retry × 1 s backoff covers the expected lag without a noticeable delay for users on a healthy backend.

**Alternatives considered**:
- Infinite retry with exponential backoff: risks hanging the UI forever; user should see the error state eventually.
- No retry (status quo): causes every newly registered user to see the error state immediately.
- Client-side retry via the "Tentar novamente" button only: requires the user to act; bad UX for a transient failure.

---

## Decision 6: ChangeTrackStore — new or reuse OnboardingStore?

**Decision**: Create a new `ChangeTrackStore` in `presentation/profile/` with its own `State`, `Intent`, and `Label`. It does not share code with `OnboardingStore`.

**Rationale**: `OnboardingStore` manages `pageIndex` and multi-page navigation — irrelevant for track re-selection which is a single-screen interaction. Creating a purpose-built store with only `selectedTrackIds`, `isLoading`, and `error` state is simpler and avoids coupling two different flows.

**Alternatives considered**:
- Subclassing or delegating to `OnboardingStore`: The stores have incompatible state shapes; inheritance/delegation would complicate both.
- Sharing a `TrackSelectionViewModel` supertype: premature abstraction (only two use sites, and they differ in navigation outcome).
