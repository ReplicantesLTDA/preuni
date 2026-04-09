# UI Contracts: Subject Track Path

## OnboardingScreen (modified)

**Composable signature** (no change):
```
OnboardingScreen(store: OnboardingStore, onCompleted: () -> Unit, contentRepository: ContentRepository)
```

**Chip behaviour change**:
- Before: `FilterChip` with toggle semantics; multiple chips can be selected
- After: Radio-chip behaviour; selecting one chip automatically deselects the previous

**"Começar" button**:
- Enabled: `state.selectedTrackId != null && !state.isLoading`
- On click: `store.accept(Intent.Complete)`
- Navigation: triggered by `Label.Completed` → calls `onCompleted()`

---

## LearnScreen (new)

```
LearnScreen(store: LearnStore, onOpenLesson: (moduleId: String) -> Unit)
```

**Top section**: Subject header bar showing `activeTrack.name` and track colour accent.

**Main content**: Vertically scrollable S-curve node path.
- Each `ModuleNode` is rendered as a `StarNode` composable
- `StarNode(node: ModuleNode, onTap: () -> Unit)` — size 56.dp, star shape
- Connecting bezier path drawn via `Canvas` behind the nodes
- Node positions: even index → 70% x, odd index → 30% x; 100.dp vertical spacing

**StarNode visual states**:
| State       | Fill                 | Border           | Overlay          |
|-------------|----------------------|------------------|------------------|
| COMPLETED   | subject colour       | gold (2.dp)      | none             |
| ACTIVE      | subject colour       | white (2.dp)     | pulsing ring anim|
| LOCKED      | surfaceVariant       | outline (1.dp)   | lock icon (16.dp)|

**Interactions**:
- Tap ACTIVE node → `Label.OpenLesson(moduleId)` → `onOpenLesson(moduleId)`
- Tap LOCKED node → `Label.ShowLockedMessage` → Snackbar "Complete os módulos anteriores primeiro"
- Tap COMPLETED node → no-op (or replay lesson — out of scope)

---

## PlaceholderLessonScreen (new, minimal)

```
PlaceholderLessonScreen(moduleTitle: String, onBack: () -> Unit)
```

Shows: back button, module title, "Em breve — conteúdo chegando!" message. No logic.

---

## TrackSelectionScreen / ChangeTrackScreen (modified)

Both screens change chip semantics from toggle to single-select:
- `FilterChip` `selected` state: `track.id == state.selectedTrackId` (was `track.id in state.selectedTrackIds`)
- `onClick`: `store.accept(Intent.SelectTrack(track.id))` (was `Intent.ToggleTrack`)

---

## TokenStore additions

```kotlin
fun getActiveTrackId(): String?       // reads KEY_ACTIVE_TRACK_ID from SecureStorage
fun setActiveTrackId(trackId: String) // writes KEY_ACTIVE_TRACK_ID
fun clearActiveTrackId()              // removes KEY_ACTIVE_TRACK_ID
```

Added alongside existing `welcomeSeen()` / `markWelcomeSeen()` pattern.
