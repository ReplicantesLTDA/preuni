# Data Model: Onboarding Flow and UX Fixes

## Entities

### WelcomeSeenFlag

A device-local boolean stored as a string in `SecureStorage`.

| Field | Type | Notes |
|-------|------|-------|
| `welcome_seen` | `String` ("true" / absent) | Key in `SecureStorage`; absent = not yet seen |

**Validation**: No validation — any non-null value means "seen".  
**Lifecycle**: Written once (on welcome complete or skip). Never cleared on logout. Cleared only on full app data wipe.

---

### WelcomeStore.State

Internal state for the pre-login welcome slides (3 pages).

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `pageIndex` | `Int` | `0` | Current slide index (0–2) |

**Intents**: `NextPage`, `PreviousPage`, `Skip`, `Complete`  
**Labels**: `Completed` (emitted on last page "next" or `Skip`; triggers welcome-seen write + navigate to Auth)

---

### ChangeTrackStore.State

Internal state for the post-login track re-selection screen accessed from Profile.

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `selectedTrackIds` | `Set<String>` | (current enrolled tracks) | Pre-populated from user profile |
| `isLoading` | `Boolean` | `false` | True while saving |
| `error` | `AppError?` | `null` | Set on save failure |

**Intents**: `ToggleTrack(trackId)`, `Save`, `Load`  
**Labels**: `Saved` (emitted on success; triggers navigation back to Profile)

**Validation**: `Save` is rejected with `Label.ValidationError` if `selectedTrackIds.isEmpty()`.

---

### HomeStore.State (modified)

Adds internal retry tracking. The public `State` interface is unchanged — `retryCount` is private to the executor.

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `student` | `Student?` | `null` | Loaded from user service |
| `isLoading` | `Boolean` | `false` | True during load or retry |
| `error` | `AppError?` | `null` | Set only after 3 failed attempts |

**Retry logic**: On load failure, executor retries up to 3 times with a 1-second delay. Error state is emitted only on the 3rd failure.

---

## State Transitions

### RootComponent routing (new logic)

```
App launch
  └─ welcomeSeen? ──No──→ Config.Welcome ──complete/skip──→ Config.Auth
                │
               Yes
                └─ loggedIn? ──No──→ Config.Auth
                            │
                           Yes
                            └─ Config.Main   (onboarding flag resolved at runtime — stub ok for now)

AuthComponent.onAuthenticated → navigation.replaceAll(Config.Onboarding)
OnboardingComponent.onCompleted → navigation.replaceAll(Config.Main)
```

### ProfileComponent track re-selection (new route)

```
Config.Profile ──"Alterar matérias"──→ Config.ChangeTrack
Config.ChangeTrack ──save/back──→ Config.Profile (pop)
```
