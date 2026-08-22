# Quickstart — 012-expo-rn-frontend

Run the new React Native + Expo app against the local Go monolith.

## Prerequisites

- Node 20 LTS, **pnpm** 9+
- Xcode 15+ (iOS Simulator) **or** Android Studio with API 34 emulator
- **Expo Go** installed on a physical iPhone or Android phone (for device testing)
- Local backend running (see `backend/app/`): `make dev` from repo root brings up Postgres + Redis + monolith + NGINX gateway on `http://localhost:8080`

## First-time setup

```bash
cd mobile
pnpm install
cp .env.example .env.local
# .env.local contains: EXPO_PUBLIC_API_BASE_URL=http://<your-LAN-ip>:8080
```

> The LAN IP (not `localhost`) is required when running on a physical device via Expo Go — the phone must reach your machine. `ipconfig getifaddr en0` on macOS.

## Run

```bash
pnpm start                       # opens the Expo dev menu — scan QR with Expo Go
pnpm ios                         # iOS Simulator
pnpm android                     # Android emulator
pnpm web                         # browser at http://localhost:8081
```

Hot reload should propagate a saved change to all attached clients within ~10s (SC-006).

## Smoke test — US1 (Trilha core loop)

1. Sign in as the seeded admin: `admin@preuni.com` / `<password from migration 008>`.
2. Land on Trilha; verify streak, XP, readiness score render.
3. Tap the recommended next activity; complete it.
4. Return to Trilha; XP / streak / next-step update without a manual refresh.

## Smoke test — US2 (Register → verify → onboarding)

1. From a clean install (wipe SecureStore via dev menu), tap "Criar conta".
2. Register a new email; receive OTP (logged by the local mail dev sender).
3. Enter OTP. Verify "reenviar" works.
4. Land on onboarding; complete; land on Trilha already authed (FR-007).

## Tests

```bash
pnpm tsc --noEmit
pnpm lint
pnpm test --coverage              # Jest + RTL + msw
pnpm test:e2e:android             # Detox smoke (post-MVP)
```

## Web build (Lighthouse review)

```bash
pnpm expo export --platform web
pnpm dlx serve dist
# open http://localhost:3000 → run Lighthouse (mobile, throttled 4G)
# LCP ≤ 2.5s, INP ≤ 200ms, CLS ≤ 0.1 (Constitution IV)
```

## Coexistence note

`mobile/` (KMP) is **frozen** during this feature. Do not add features there. Once US1–US4 are demonstrably at parity on the new app, a follow-up commit deletes `mobile/` (FR-017) and updates `CLAUDE.md` + `Makefile`.

## Stop

```bash
# foreground processes — Ctrl+C
```
