# Data Model: Subject Track Path

## Entities

### ModuleNode (new — client-side only for now)

Represents a single step in a subject's learning path.

| Field      | Type        | Description                                      |
|------------|-------------|--------------------------------------------------|
| `id`       | `String`    | Unique identifier (stub: "module-{index}")       |
| `index`    | `Int`       | Zero-based position in the track                 |
| `title`    | `String`    | Display name (stub: "Módulo {index+1}")          |
| `state`    | `NodeState` | `LOCKED`, `ACTIVE`, or `COMPLETED`               |
| `xpReward` | `Int`       | XP awarded on completion (stub: 10)              |

**NodeState transitions**:
```
LOCKED → ACTIVE    (when all previous nodes are COMPLETED)
ACTIVE → COMPLETED (when the user completes the lesson)
COMPLETED → —      (terminal — no regression)
```

The first node of a track is always `ACTIVE` for a new user. All others start `LOCKED`.

---

### OnboardingStore.State (modified from feature 005)

| Field             | Type        | Before (005)        | After (006)      |
|-------------------|-------------|---------------------|------------------|
| `selectedTrackId` | `String?`   | *(not present)*     | NEW — single ID  |
| `selectedTrackIds`| `Set<String>` | present           | REMOVED          |
| `isLoading`       | `Boolean`   | unchanged           | unchanged        |
| `error`           | `AppError?` | unchanged           | unchanged        |

**Intent changes**:
- REMOVED: `ToggleTrack(trackId: String)` (toggle semantics)
- ADDED: `SelectTrack(trackId: String)` (replace semantics)
- UNCHANGED: `Complete`

**Validation**: `selectedTrackId != null` (was `selectedTrackIds.isNotEmpty()`)

---

### ChangeTrackStore.State (modified from feature 005)

Same migration as OnboardingStore:

| Field             | Type        | Before (005)        | After (006)      |
|-------------------|-------------|---------------------|------------------|
| `selectedTrackId` | `String?`   | *(not present)*     | NEW              |
| `selectedTrackIds`| `Set<String>` | present           | REMOVED          |
| `isLoading`       | `Boolean`   | unchanged           | unchanged        |
| `error`           | `AppError?` | unchanged           | unchanged        |

---

### LearnStore.State (new)

| Field           | Type                  | Description                                        |
|-----------------|-----------------------|----------------------------------------------------|
| `activeTrackId` | `String?`             | ID of the currently selected subject track          |
| `activeTrack`   | `Track?`              | Full Track object (name, color, etc.) for header    |
| `modules`       | `List<ModuleNode>`    | Ordered list of module nodes for the path           |
| `isLoading`     | `Boolean`             | True while fetching track/module data               |
| `error`         | `AppError?`           | Set on fatal load failure                           |

**Intents**:
- `Load` — triggered on mount; reads `activeTrackId` from `TokenStore`, loads tracks, generates module stubs
- `TapNode(nodeId: String)` — user taps a module node
- `Retry` — retry after error

**Labels**:
- `OpenLesson(moduleId: String)` — navigate to lesson (placeholder screen for now)
- `ShowLockedMessage` — show "complete previous modules first" snackbar

---

### TokenStore additions (new keys)

Two new methods added to existing `TokenStore`:

```
KEY_ACTIVE_TRACK_ID = "active_track_id"

fun getActiveTrackId(): String?
fun setActiveTrackId(trackId: String)
fun clearActiveTrackId()
```

---

## State Transitions

### Onboarding completion flow

```
[OnboardingScreen]
  User taps SelectTrack(id) → state.selectedTrackId = id
  User taps Complete:
    1. tokenStore.setActiveTrackId(selectedTrackId)   ← local save
    2. navigation.replaceAll(Config.Main)              ← immediate navigation
    3. Background: userRepository.updateTracks([id])  ← fire-and-forget sync
```

### Subject switch flow (Profile → Alterar matérias)

```
[ChangeTrackScreen]
  User taps SelectTrack(id) → state.selectedTrackId = id
  User taps Save:
    1. tokenStore.setActiveTrackId(selectedTrackId)   ← local save
    2. Background: userRepository.updateTracks([id])  ← fire-and-forget sync
    3. Label.Saved → ProfileComponent.onSaved()
       → mainComponent.selectTab(BottomTab.LEARN)     ← switch to Learn tab
```

### LearnStore load flow

```
[LearnScreen mounted]
  Load intent dispatched:
    1. activeTrackId = tokenStore.getActiveTrackId()
    2. If null → state with no track (show subject picker prompt)
    3. tracks = contentRepository.getTracks()
    4. activeTrack = tracks.find { it.id == activeTrackId }
    5. modules = generateStubModules(15)               ← stub for now
    6. state = State(activeTrackId, activeTrack, modules)
```

## S-Curve Layout Model

```
Node index → horizontal position:
  even  → 70% from left (right side)
  odd   → 30% from left (left side)

Vertical spacing between nodes: 100.dp

Connector bezier control points (node i → node i+1):
  start  = center of node i
  end    = center of node i+1
  ctrl1  = (start.x, start.y + 50.dp)
  ctrl2  = (end.x, end.y - 50.dp)
```

Visual example (5 nodes):
```
         ★  ← node 0 (COMPLETED, right)
       ╱
  ★       ← node 1 (COMPLETED, left)
       ╲
         ★  ← node 2 (ACTIVE, right) ← pulsing ring
       ╱
  ★       ← node 3 (LOCKED, left)
       ╲
         ★  ← node 4 (LOCKED, right)
```
