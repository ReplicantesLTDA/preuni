# Quickstart: Fix Backend Integration Issues (Home + Profile)

**Feature**: 008-fix-backend-integrations
**Date**: 2026-04-15

---

## What this feature delivers

1. Home and Profile data integration behaves consistently across Android, iOS, and Web.
2. Home/Profile requests use strict timeout + transient retry behavior.
3. Request telemetry captures timing and outcomes with sensitive data redaction.
4. End-to-end/integration tests catch contract regressions before merge.

---

## Local setup

```bash
# Infrastructure
docker compose -f infra/docker-compose.yml up -d postgres redis

# Start backend stack (includes gateway)
make run-backend
```

---

## Run app targets

```bash
# Web
make run-web

# Android
make run-android

# iOS
make run-ios
```

Login with a valid user and verify:
- Home loads profile summary metrics (name, XP, streak, readiness)
- Profile loads full account snapshot (name, username, email, avatar metadata)
- Manual retry refreshes data after simulated failure

---

## Run tests

### KMP shared tests

```bash
cd mobile/shared && ./gradlew desktopTest --tests "*.UserApiClientTest"
cd mobile/shared && ./gradlew desktopTest --tests "*.HomeStoreTest"
cd mobile/shared && ./gradlew desktopTest --tests "*.ProfileStoreTest"
cd mobile/shared && ./gradlew desktopTest --tests "*.NetworkStackFactoryTest"
```

### user-svc integration tests

```bash
cd backend/svc/user
TEST_DB_URL="postgres://preuni:preuni@localhost:5432/preuni?search_path=users&sslmode=disable" \
  go test ./handler/... -run Integration
```

---

## Expected outcomes

- Home/Profile requests carry bearer auth on all client targets.
- Fetch failures produce user-friendly retry behavior (no raw technical errors).
- Retry attempts occur only for transient conditions (NetworkError).
- Contract drift (missing fields, malformed payloads, validation envelope changes) fails tests.

---

## Degraded-network verification

### Simulating network failure (Android emulator)

1. Start the app and reach the Home screen.
2. In Android Studio, open **Extended Controls → Cellular → Network type → No network**.
3. Pull to refresh (or navigate away and back to Home).
4. **Expected**: Loading indicator appears, then after ~3 retries (≈3.5s) an error card
   appears with "Não foi possível carregar os dados." and a "Tentar novamente" button.

### Simulating slow network (Android emulator)

1. In **Extended Controls → Cellular**, set speed to **Edge** (≈1 Kbps).
2. Navigate to Home. The per-request timeout fires at 10s.
3. **Expected**: Same error card surfaces within ~10 seconds.

### Verifying auto-retry recovers

1. Disable network as above.
2. Open Home — error state appears.
3. Re-enable network.
4. Tap "Tentar novamente".
5. **Expected**: Loading indicator → student data renders successfully.

### Verifying non-transient errors are not retried

1. Invalidate the session (clear app data or expire the token).
2. Open Home.
3. **Expected**: Error surfaces immediately (single request, no retry delay) since
   `401 Unauthorized` is classified as non-transient.

### Telemetry verification (debug builds)

Logcat filter: `Telemetry[` — each request attempt emits a structured line:
```
Telemetry[method=GET route=/v1/students/me status=200 attempt=0 duration=142ms outcome=SUCCESS]
```
Confirm no `Authorization:` header values or email addresses appear in logcat output.
