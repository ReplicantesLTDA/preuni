# Research: Subject Track Path

## Decision 1: Why "Começar" doesn't work — root cause

**Decision**: The button IS wired correctly. The real failure is that `OnboardingStore.complete()` calls `userRepository.updateTracks(trackIds)` (a `PATCH /v1/students/me/onboarding` network request). If the backend is unreachable or returns an error, the store dispatches `Msg.ErrorReceived` and never emits `Label.Completed`. Without that label, the `onCompleted()` callback never fires, so navigation to Main never happens. The user sees the button do "nothing" (no spinner, no navigation, just a silent error state that may not be visually obvious).

**Fix strategy**: Save the selected track ID locally (in `TokenStore` via `SecureStorage`) *before* the network call. Navigate to Main immediately after local save. Sync to backend in background (fire-and-forget). This makes onboarding completion instant and resilient to network failures.

**Rationale**: Onboarding state is device-local by design (like `welcomeSeen`). Blocking the user on a network round-trip for local state is a UX anti-pattern. The track enrollment backend endpoint already has a `TODO` for writing `track_enrollments` rows — it's not yet authoritative anyway.

**Alternatives considered**:
- Show error to user and let them retry: Poor UX; the user picked a subject, there's no reason to gate them.
- Fix the backend and keep current flow: Addresses network errors but still blocks on latency.

---

## Decision 2: Single-select for subject (OnboardingStore refactor)

**Decision**: Replace `selectedTrackIds: Set<String>` with `selectedTrackId: String?`. `Intent.ToggleTrack` becomes `Intent.SelectTrack(trackId: String)` — selecting always replaces the previous selection. Validation: `selectedTrackId != null`.

**Rationale**: The user confirmed single-subject model (like Duolingo). The existing multi-select store is wrong by spec. The `TrackSelectionScreen` chip UI must switch from `FilterChip` (toggle) to radio-style chips where selecting one deselects the other.

**Impact on existing tests**: `OnboardingStoreTest` must be updated — the 4 page-navigation tests were already removed in feature 005. Remaining tests need updated assertions for single selection.

**Alternatives considered**:
- Keep multi-select, just limit to 1 in validation: Confusing UX (chips look deselectable but aren't).
- Separate screen: Overkill for what is a simple radio-button interaction.

---

## Decision 3: Active subject storage

**Decision**: Add `activeTrackId` key to `TokenStore` via `SecureStorage`. Store it on onboarding completion and on subject change. `LearnStore` reads it at init via an injected `TokenStore` dependency.

**Rationale**: `TokenStore` already owns device-local user state (tokens, welcome flag). Adding one more key is consistent and avoids introducing a new storage class. The active track is device-local (not account-linked) for now.

**Key**: `KEY_ACTIVE_TRACK_ID = "active_track_id"` → stores the `Track.id` string (e.g. `"matematica"`).

**Alternatives considered**:
- Fetch active track from student profile API (`enrolled_track_ids` field): Not yet returned by the API; fragile.
- Pass through Decompose navigation state: Would require serializable navigation arguments and complicate the back-stack.

---

## Decision 4: S-curve path rendering approach

**Decision**: Use a `LazyColumn` of module nodes where each node is positioned left or right using a `BoxWithConstraints`. A `Canvas` drawn *behind* the nodes paints the connecting bezier curve path. Each node's horizontal offset alternates: even index → 70% from left, odd index → 30% from left.

**Rationale**: `LazyColumn` gives free scrolling, virtualization, and scroll state persistence. Canvas for the connecting path is the idiomatic Compose approach for custom drawing. The zigzag is simple math — no custom `Layout` needed.

**S-curve algorithm**:
- Node `i` is placed at x = `if (i % 2 == 0) 0.7f else 0.3f` of screen width
- The connector between node `i` and `i+1` is drawn as a cubic bezier: control points pulled toward the opposite side to create a smooth S-curve
- Node spacing (vertical): ~100.dp fixed

**Star shape**: Drawn as a `Canvas` using a 5-point star `Path`. Three visual states:
- `COMPLETED`: filled with subject color, gold star outline
- `ACTIVE`: filled with subject color, animated pulsing ring
- `LOCKED`: grey fill, dark outline, lock icon overlay

**Alternatives considered**:
- Custom `Layout` with absolute coordinates: More powerful but complex to implement and maintain.
- Image assets for nodes: Not flexible (can't change state dynamically, harder to animate).
- `HorizontalPager` rows: Wrong scroll axis; loses the S-curve visual.

---

## Decision 5: Stub module data strategy

**Decision**: `LearnStore` generates a fixed list of 15 stub `ModuleNode` objects when no real backend data is available. The first node is `ACTIVE`, the rest are `LOCKED`. Stub nodes have generic titles ("Módulo 1", "Módulo 2", …). A real content endpoint will replace this later.

**Rationale**: The backend content service has no module endpoints yet (`GET /health` only). Shipping a screen that requires backend data would make the feature unusable. Stub data lets us validate the UI and gamification feel immediately.

**Real data path (future)**: A `GET /v1/tracks/{trackId}/modules` endpoint will return the ordered module list. `LearnStore` will call it and replace stubs on success, keeping stubs as fallback.

**Alternatives considered**:
- Block screen on backend data: Not viable — no endpoint exists.
- Hardcode all 5 tracks' modules in the app: Too much maintenance; will be wrong the moment real content exists.

---

## Decision 6: ChangeTrackStore single-select migration (feature 005 compatibility)

**Decision**: Modify existing `ChangeTrackStore` (built in 005) to mirror the OnboardingStore change: `selectedTrackIds: Set<String>` → `selectedTrackId: String?`. After save, call `tokenStore.setActiveTrackId(selectedTrackId)`. Navigate to Learn tab.

**Rationale**: ChangeTrackStore was built for multi-select (matching the old OnboardingStore). Now both must be single-select for consistency.

**Test impact**: `ChangeTrackStoreTest` must be updated — 4 tests need assertion changes.

---

## Decision 7: Navigation after subject change

**Decision**: After completing onboarding OR after changing subject from Profile:
1. Save `activeTrackId` to `TokenStore` locally
2. Navigate to `Main` (or switch to Learn tab if already in Main)
3. Background: sync track selection to backend

`OnboardingComponent.onCompleted()` → `RootComponent.replaceAll(Config.Main)` (unchanged). The Learn tab reads the newly stored `activeTrackId` and shows the correct subject's path.

For "Alterar matérias" from Profile: after `ChangeTrackStore` emits `Label.Saved`, `ProfileComponent` should switch the bottom nav to the Learn tab instead of just going back. This requires MainComponent to expose a `selectTab(BottomTab.LEARN)` that ProfileComponent can call via a callback.

**Alternatives considered**:
- Navigate back to Profile after subject change: Doesn't surface the new track; anticlimactic.
- Restart the app: Heavy-handed; loses scroll state.
