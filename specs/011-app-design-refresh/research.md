# Research: App Design Refresh — Core Flow

**Date**: 2026-05-25
**Inputs**:
- Wireframe reference: `mobile/wireframe/wireframe.html`
- Feature spec: `specs/011-app-design-refresh/spec.md`

This document resolves planning unknowns and records the decisions that guide implementation.

## Decision 1 — Map wireframe destinations onto existing navigation

**Decision**: Update the main bottom navigation to match the wireframe destinations while reusing existing screens where possible.

**Chosen mapping**:
- **Trilha** → reuse existing `LearnScreen` (subject path)
- **Redação** → repurpose existing `SimulateScreen` area into the writing experience entry (initially a UI placeholder aligned to the wireframe)
- **Amigos** → new screen (UI placeholder)
- **Liga** → new screen (UI placeholder)
- **Perfil** → reuse existing `ProfileComponent`/`ProfileScreen` and add a settings destination reachable from a gear action

**Rationale**: The wireframe defines the product’s mental model (study path + writing + social + status + self-management). Reusing the existing Learn/Profile stack minimizes risk while allowing the navigation labels and layout to match the new design intent.

**Alternatives considered**:
- Keep current tabs (Início/Aprender/Simular/Perfil): rejected because it conflicts with the wireframed journey and keeps the UI feeling like a generic template.
- Add wireframe screens without changing navigation: rejected because the experience would remain confusing (users cannot discover destinations naturally).

## Decision 2 — Introduce a shared “top status bar” UI primitive

**Decision**: Add a reusable top status bar component consistent with the wireframe (compact, scannable, present on main destinations), showing what is available today and placeholders for future metrics.

**What it shows**:
- Always: **streak** and **XP** (existing student fields)
- Placeholders (until real data exists): any additional counters shown in the wireframe (displayed with safe default values and clear semantics)

**Rationale**: A persistent, calm status bar reinforces continuity across sections and makes the product feel intentional.

**Alternatives considered**:
- Keep stats only on Home/Profile: rejected because it breaks cross-screen cohesion.
- Hide unknown metrics entirely: rejected because the wireframe layout expects the space; placeholders avoid layout churn later.

## Decision 3 — Mascot placeholder treatment

**Decision**: Implement a `MascotPlaceholder` composable that occupies the mascot slot in the welcome flow (and any other spot where the wireframe shows mascot presence), using a neutral-friendly placeholder.

**Constraints**:
- No custom artwork is introduced in this feature.
- Placeholder must still feel “designed”, not like a missing image.
- Placeholder must include an accessibility label (screen readers).

**Rationale**: The mascot is part of the emotional tone, but the asset is not ready. A placeholder keeps the composition and “feeling” intact.

**Alternatives considered**:
- Remove the mascot region: rejected because it changes the intended emotional anchor of the welcome screen.
- Use a random third-party illustration: rejected (inconsistent style + licensing risk).

## Decision 4 — Centralize spacing tokens (reduce ad-hoc dp)

**Decision**: Add spacing tokens (e.g., xs/sm/md/lg/xl) and migrate all screens touched by this feature to use them.

**Rationale**: The constitution requires design token adherence. Central spacing scales also make the “breathable” layout consistent across screens.

**Alternatives considered**:
- Keep per-screen `dp` literals: rejected due to consistency and governance requirements.

## Decision 5 — Copy consistency (Portuguese, friendly, encouraging)

**Decision**: Normalize user-facing strings in the refreshed flow to Portuguese and to the tone implied by the wireframe.

**Rationale**: Mixed-language UI undermines trust and breaks the warm, motivating feel.

**Alternatives considered**:
- Leave English strings for later: rejected because this feature is explicitly about “feeling” and end-to-end flow coherence.

## Decision 6 — Placeholder screens must still be “real” experiences

**Decision**: New destinations (Amigos/Liga/Ajustes) may start as UI placeholders, but they must include:
- A clear title
- A friendly empty state explaining what exists now vs what comes next
- A consistent scaffold (top status + bottom nav)

**Rationale**: “Em breve” alone feels unfinished; a structured empty state preserves product confidence.

**Alternatives considered**:
- Plain centered text placeholders: rejected because it contradicts the goal of a major design refresh.

---

## Summary of resolved unknowns

- Navigation structure follows the wireframe and reuses existing Learn/Profile.
- Mascot is a deliberate placeholder (no artwork dependency).
- Spacing tokens are introduced to meet design-token governance.
- New tabs can ship as UI placeholders while maintaining a polished feel.