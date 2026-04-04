# Quickstart: Friendly UI Polish — Core Frontend

**Feature**: `003-ui-polish` | **Date**: 2026-04-04

How to preview, develop, and verify UI changes for this feature.

---

## 1. Start the Dev Server

```bash
# From repo root — start the Kotlin/Wasm web app
make run-web
```

Open `http://localhost:8088` in a browser. This is the primary visual review target.

> Stuck at 98%? That's normal — webpack-dev-server watch mode. The app is running.

---

## 2. Verify the Theme is Applied

After adding `PreuniTheme { ... }` in `PreuniApp.kt`:

1. Open the login screen (`http://localhost:8088`)
2. Check: input fields have rounded corners (≥ 16dp visible rounding)
3. Check: the "Sign in" button is pill-shaped (fully rounded ends)
4. Check: primary color is violet (`#6750A4`)

If the default M3 rectangular text fields still show: confirm `PreuniTheme` wraps `PreuniApp` in `PreuniApp.kt` before the `when (child)` block.

---

## 3. Preview All Auth Screens

```
Login      → http://localhost:8088  (initial screen)
Register   → click "Create account"
VerifyEmail→ complete registration (check mailbox at http://localhost:4000/dev/mailbox)
OtpLogin   → click "Sign in with email code" on login screen
```

For each screen verify:
- [ ] Inputs are rounded
- [ ] Primary CTA is pill-shaped and full-width
- [ ] Error states are inline (enter wrong password, click Sign in)
- [ ] Spacing feels breathable (≥ 24dp between sections)

---

## 4. Preview the Home Screen

Log in with a registered account. Check:
- [ ] Greeting includes display name
- [ ] Streak badge and XP pill are visible
- [ ] Subject track cards appear in the 5-area grid
- [ ] Each track card has its distinct background color
- [ ] "Aprender agora" CTA is pill-shaped

---

## 5. Preview Onboarding (New Account)

Create a new account, verify email, then:
- [ ] Onboarding slides are one-per-screen
- [ ] Page 4 shows all 5 subject track cards
- [ ] At least one track must be selected for "Começar" to activate
- [ ] Selected tracks show visual highlight (border)

---

## 6. Preview the Bottom Navigation

- [ ] All 4 tabs show correct icons (not the placeholder Home icon for LEARN and SIMULATE)
- [ ] Active tab uses brand violet color
- [ ] Tab labels are legible

---

## 7. Test at 360px Viewport (Minimum Width)

In browser DevTools: set viewport to 360px width.

- [ ] No horizontal scrollbar on any screen
- [ ] Input fields and buttons fit within the viewport
- [ ] Cards do not clip or overflow

---

## 8. Quick Component Isolation

To test a component in isolation without running the full app, add a temporary `PreviewScreen` composable in the relevant screen file:

```kotlin
// Temporary debug only — remove before merge
@Composable
fun ComponentDebug() {
    PreuniTheme {
        Column(Modifier.padding(16.dp)) {
            PreuniTextField(value = "", onValueChange = {}, label = "Email")
            Spacer(Modifier.height(8.dp))
            PreuniTextField(
                value = "",
                onValueChange = {},
                label = "Email",
                isError = true,
                errorMessage = "E-mail inválido",
            )
            Spacer(Modifier.height(16.dp))
            PreuniButton(text = "Entrar", onClick = {})
            Spacer(Modifier.height(8.dp))
            PreuniButton(text = "Carregando...", onClick = {}, isLoading = true)
        }
    }
}
```

Wire it temporarily in `PreuniApp.kt` as the only child, then remove before merge.

---

## 9. Backend Not Required for UI Work

All UI changes in this feature are purely presentational. The backend doesn't need to be running for visual development. Use `make run-web` only (no `make run-backend`).

Exception: if testing the full auth flow end-to-end, start the backend:

```bash
make run-backend   # PostgreSQL + Redis + all Go services + mail service
```

---

## 10. Stopping

```bash
make stop-web      # stop webpack-dev-server
make stop-backend  # stop all Docker containers
```
