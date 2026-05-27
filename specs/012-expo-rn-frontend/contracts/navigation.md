# Contract: Navigation (Expo Router)

**Library**: `expo-router` (file-based, on top of React Navigation). `typedRoutes: true` in `app.json`.

## Route tree

```text
mobile/app/
├── _layout.tsx                  # providers (Query, Theme, SafeArea, Fonts), auth gate
├── index.tsx                    # decides /(tabs)/trilha vs /(auth)/login vs /(onboarding)/welcome
├── +not-found.tsx
│
├── (auth)/_layout.tsx           # public; redirects to /(tabs) if authed
├── (auth)/welcome.tsx           # first-launch mascot screen
├── (auth)/login.tsx
├── (auth)/register.tsx
├── (auth)/verify-email.tsx      # ?email=…
├── (auth)/otp-login.tsx
├── (auth)/password-reset.tsx
│
├── (onboarding)/_layout.tsx     # authed + onboardingCompleted=false; redirects to /(tabs) if completed
├── (onboarding)/welcome.tsx
├── (onboarding)/profile.tsx
├── (onboarding)/interests.tsx
│
└── (tabs)/_layout.tsx           # authed + onboardingCompleted=true; hosts TopStatusBar + BottomNavigation
    ├── trilha/index.tsx
    ├── trilha/[trackId].tsx
    ├── redacao/index.tsx
    ├── redacao/[promptId].tsx
    ├── simulado/index.tsx
    └── perfil/
        ├── index.tsx
        ├── edit.tsx
        ├── change-email.tsx
        ├── change-password.tsx
        └── data-export.tsx
```

## Gating rules (enforced in layouts, not in screens)

| Layout | Allowed when |
|--------|--------------|
| `(auth)/_layout.tsx` | `session.status === 'anon'` — else `redirect('/(tabs)/trilha')` |
| `(onboarding)/_layout.tsx` | `session.status === 'authed' && !student.onboardingCompleted` — else redirect to `(tabs)` or `(auth)` |
| `(tabs)/_layout.tsx` | `session.status === 'authed' && student.onboardingCompleted` — else redirect |

While `session.status === 'loading'`, the root renders a splash and waits — no flicker.

## Tabs

`(tabs)/_layout.tsx` renders the **BottomNavigation** contract for Trilha / Redação / Simulado / Perfil. **TopStatusBar** is rendered at this layout level too, hidden on routes that opt out via `unstable_settings`.

## Deep-link / web routing

- iOS/Android scheme: `preuni://` (configured in `app.json`).
- Web: served at `/` after `expo export --platform web`.
- Verify-email URLs delivered by the backend resolve to `/(auth)/verify-email?email=…` (no token in URL; OTP entered manually).

## Acceptance

- [ ] Hard-refreshing any URL on web lands on the correct gated route.
- [ ] Back button on Android pops within a group; from a group root it returns to the previous tab, not out of the app.
- [ ] `session.status === 'loading'` never shows a tab bar.
