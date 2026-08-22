# Quickstart verification — 012-expo-rn-frontend

**Branch**: `012-expo-rn-frontend`
**Verified on**: 2026-05-27
**Stack**: SDK 54, RN 0.81, React 19.1, Node 26.

## Automated

| Check | Result |
|-------|--------|
| `pnpm typecheck` | ✓ green |
| `pnpm lint` | ✓ green (0 warnings) |
| `pnpm test` | ✓ 21 suites, 79 tests in ~2s |
| `pnpm build:web` | ✓ 28 routes exported to `dist/` |
| `expo-doctor` | ✓ 18/18 checks |

## Manual (recorded)

| Flow | Device | Status |
|------|--------|--------|
| Register → email verify (with resend cooldown) → onboarding (welcome → profile → interests) → Trilha | iPhone 15 via Expo Go | ✓ |
| Admin login (`admin@preuni.com`) → Trilha home with streak/XP/readiness | iPhone 15 | ✓ (after case-converter fix) |
| Logout from Perfil → land on welcome → re-login | iPhone 15 | ✓ |
| Edit displayName | iPhone 15 | ✓ |
| Avatar pick → upload → confirm | iPhone 15 | ✓ (dev S3 stub skipped; confirm 200) |
| Change password | iPhone 15 | not yet exercised |
| Change email (request → OTP confirm) | iPhone 15 | not yet exercised |
| Data export request | iPhone 15 | not yet exercised |
| Account deletion (with `EXCLUIR` confirm) | iPhone 15 | not yet exercised |
| Tab switching (Trilha ↔ Redação ↔ Simulado ↔ Perfil) | iPhone 15 | ✓ (uses `router.navigate`, no stack growth) |

## Manual (pending)

- VoiceOver smoke pass on iPhone (a11y.md unchecked items).
- TalkBack smoke pass on Android (no device tested yet).
- Web build smoke on Chrome / Safari / Firefox (audit at 360 / 768 / 1280 px).
- Lighthouse CI on `dist/` — Constitution IV gates (LCP / INP / CLS).

## Smoke commands

```bash
# Native
cd mobile && pnpm start             # QR code → Expo Go
cd mobile && pnpm ios               # iOS Simulator
cd mobile && pnpm android           # Android emulator

# Web
cd mobile && pnpm web               # dev server at :8081
cd mobile && pnpm build:web && pnpm dlx serve dist   # static export preview
```

## Open user-facing risks

- Avatar URL points to a non-existent S3 object in dev. UI doesn't 404 because `expo-image` silently falls back to placeholder. In production an unsigned S3 bucket would 403.
- Phone needs LAN IP in `mobile/.env.local`. `EXPO_PUBLIC_API_BASE_URL=http://<lan-ip>:8080`. Mismatched IP = "Network request failed".
