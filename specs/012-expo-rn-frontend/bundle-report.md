# Bundle report — 012-expo-rn-frontend

**Generated**: 2026-05-27 via `pnpm build:web`.

## Summary

| Metric | Value | Constitution IV budget |
|--------|-------|------------------------|
| Total dist/ | 6.0 MB | — |
| Routes exported | 28 | — |
| Per-route HTML | 21.8 kB | — |
| Shared JS entry (raw) | 2.3 MB | — |
| Shared JS entry (gzip) | **546 kB** | ≤ 150 kB per route chunk |

## Verdict

**Over budget** on the shared entry chunk. The current `expo export --platform web` produces a single shared JS bundle that every route hydrates from. Per-route code splitting is not enabled by default in Expo Router static export.

## Root cause

`expo-router@6` static export inlines:
- React + React DOM + RN-Web runtime
- All route components (28 screens)
- TanStack Query, Zustand, Zod, expo-image, expo-router, react-native-svg, @expo-google-fonts/*

No `lazy()` boundaries are wired around heavy screens. The fonts package alone adds ~150 kB of base64-embedded TTFs.

## Mitigation options (deferred to polish phase)

1. **Switch to server output**: `app.json` → `"web": { "output": "server" }`. Enables file-system-based code splitting per route.
2. **Lazy-load font packages**: Replace `@expo-google-fonts/*` direct imports with a route-level loader that only pulls the families used on the current screen.
3. **Tree-shake `react-native-svg`**: Import only `Svg`, `Circle`, `Path` instead of the full module.
4. **Defer `expo-image-picker` + `expo-image` to feature route**: They are only used in `perfil/edit.tsx` and `(onboarding)/profile.tsx`.

## Action

Track these as Phase 8 polish work (T112). Not blocking US1/US2 MVP gate.
