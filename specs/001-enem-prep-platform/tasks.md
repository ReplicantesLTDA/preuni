# Tasks: preuni.com.br – Auth, Onboarding & Home (Sprint 1)

**Input**: Design documents from `specs/001-enem-prep-platform/`
**Prerequisites**: plan.md ✅ spec.md ✅ research.md ✅ data-model.md ✅ contracts/rest-api.md ✅

**Scope of this task file**: The 10 items requested by the user — project initialization, authentication (login + register + OTP), welcome/verification emails, user profile management, home screen, and onboarding flow. These are the foundational auth/UX layers required before the learning features (US1–US5 from spec.md).

**Constitution note**: Test tasks are included per the Constitution (II. Testing Standards — TDD is NON-NEGOTIABLE). Unit tests are written first for all business logic; integration tests are written first for all API endpoints.

**Assumption on username**: "dash and slash" interpreted as hyphen (`-`) and underscore (`_`); forward slash `/` is not valid in usernames. Update `AuthValidator` if the intent differs.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no competing changes)
- **[US#]**: Maps to the user stories defined in this scope (US1–US9 below)
- Tests MUST fail before their paired implementation tasks

---

## Phase 1: Setup — Project Initialization (User item 1)

**Purpose**: Create the monorepo skeleton and tool configuration so all subsequent phases can begin.

- [X] T001 Create monorepo root structure with `mobile/`, `backend/`, `infra/`, `specs/` directories and root `.gitignore`
- [X] T002 Initialize KMP Gradle project in `mobile/` with `settings.gradle.kts` declaring `shared`, `androidApp`, `iosApp`, `webApp` modules
- [X] T003 Configure `mobile/shared/build.gradle.kts` with Compose Multiplatform 1.8+, Decompose, MVIKotlin, Ktor Client, SQLDelight, Koin, FSRS-Kotlin, kotlin.test targets (android, ios, jvm, wasmJs)
- [X] T004 [P] Configure `mobile/androidApp/build.gradle.kts` with Android SDK 26 min, Compose UI, and shared module dependency
- [ ] T005 [P] Configure `mobile/iosApp/` Xcode project with Kotlin/Native shared framework integration
- [X] T006 [P] Configure `mobile/webApp/build.gradle.kts` for Kotlin/Wasm target with Compose Web
- [X] T007 [P] Initialize Go workspace in `backend/go.work` declaring modules for all services: `svc/auth`, `svc/user`, `svc/content`, `svc/learning`, `svc/simulation`, `svc/dissertation`, `svc/notification`, `pkg`
- [X] T008 [P] Initialize each Go service module with `go.mod` and the standard layout: `cmd/server/main.go`, `domain/`, `handler/`, `repository/`, `config/`
- [ ] T009 [P] Initialize Elixir Phoenix project in `backend/svc/mail/` with `mix phx.new preuni --no-ecto --no-assets` and add Swoosh dependency in `mix.exs`
- [X] T010 [P] Create `infra/docker-compose.yml` with PostgreSQL 16, Redis 7, and stub service entries for all Go services and NGINX
- [X] T011 [P] Create `infra/nginx/nginx.conf` skeleton with upstream blocks and routing placeholders for all services
- [X] T012 [P] Add `.editorconfig`, `ktlint` config in `mobile/.editorconfig`, and `golangci-lint` config in `backend/.golangci.yml`
- [X] T013 [P] Set up `golang-migrate` CLI runner script in `infra/scripts/migrate.sh` and create per-service migration directories in `infra/migrations/{auth,user,content,learning,simulation,dissertation}/`

**Checkpoint**: `docker compose up postgres redis` starts cleanly; `./gradlew build` compiles shared KMP module; `go build ./...` compiles all Go service stubs; `mix compile` succeeds in mail service.

---

## Phase 2: Foundational — Shared Infrastructure (Blocking)

**Purpose**: Core infrastructure that every user story depends on. Nothing else starts until this is complete.

⚠️ **CRITICAL**: No user story work begins until this phase is complete.

- [X] T014 Create `auth` schema migrations in `infra/migrations/auth/`: `001_create_credentials.sql`, `002_create_refresh_tokens.sql`, `003_create_otp_codes.sql` (exact schemas from `data-model.md`)
- [X] T015 Create `user` schema migrations in `infra/migrations/user/`: `001_create_students.sql`, `002_create_track_enrollments.sql`, `003_create_achievements.sql`, `004_create_student_achievements.sql`, `005_create_xp_events.sql`
- [X] T016 [P] Implement `backend/pkg/logger/logger.go`: zap-based structured logger with `Info`, `Error`, `With` helpers; level configurable from env
- [X] T017 [P] Implement `backend/pkg/errors/errors.go`: typed error catalogue (`ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrValidation`, `ErrInternal`) with HTTP status mappings
- [X] T018 [P] Implement `backend/pkg/config/config.go`: env-var loader using `os.Getenv` with required-field validation and typed structs per service
- [X] T019 [P] Implement `backend/pkg/middleware/auth.go`: HTTP middleware that reads `Authorization: Bearer {jwt}` header, validates JWT signature + expiry using `golang-jwt`, injects `X-User-ID` and `X-User-Email` into request context
- [X] T020 [P] Implement `backend/pkg/middleware/requestid.go`: injects `X-Request-ID` UUID into every request for tracing
- [X] T021 Implement JWT service in `backend/svc/auth/domain/jwt.go`: `IssueAccessToken(userID, email string) (string, error)` and `IssueRefreshToken() (rawToken string, tokenHash string, err error)` using `golang-jwt`; signing key from config
- [X] T022 [P] Implement `backend/svc/auth/repository/refreshtoken.go` with `pgx`: `Store(credentialID, tokenHash, expiresAt)`, `FindByHash(hash) (*RefreshToken, error)`, `Revoke(id)`
- [X] T023 Set up Ktor Client base in `mobile/shared/src/commonMain/data/network/ApiClient.kt`: base URL from config, JSON content negotiation, default timeout (30 s), global error mapping from HTTP status to typed `AppError` sealed class
- [X] T024 Define `AppError` sealed class in `mobile/shared/src/commonMain/domain/error/AppError.kt`: `NetworkError`, `Unauthorized`, `Conflict(message)`, `Validation(field, message)`, `Unknown`
- [X] T025 Set up SQLDelight schema file `mobile/shared/src/commonMain/sqldelight/PreuniDatabase.sq` with initial empty content; configure `SqlDriver` expect/actual in `commonMain`/`androidMain`/`iosMain`/`wasmJsMain`
- [X] T026 Configure Swoosh in `backend/svc/mail/config/config.exs` with `Swoosh.Adapters.Local` for dev and env-driven adapter for prod; create base email layout in `lib/preuni_web/templates/layout/email.html.heex`
- [X] T027 Complete NGINX routing in `infra/nginx/nginx.conf`: upstream blocks for each service port, JWT validation via `auth_request`, `proxy_set_header X-User-ID` forwarding, rate-limit zones (`auth_zone`: 100 r/m per IP; `api_zone`: 1000 r/m per IP)

**Checkpoint**: `go test ./backend/pkg/...` passes; Ktor Client compiles across all KMP targets; PostgreSQL migrations apply cleanly via `infra/scripts/migrate.sh`; NGINX config validates with `nginx -t`.

---

## Phase 3: User Login in KMP (User item 2) 🎯 MVP

**User Story US1**: A student opens the app and signs in with their email/username and password. Client-side validation catches malformed input before any network call.

**Validation rules**:
- **Email**: must match RFC 5322 basic format (`[^@]+@[^@]+\.[^@]+`)
- **Username**: `^[a-z][a-z0-9_-]{1,28}[a-z0-9]$` — 3–30 chars, lowercase letters, digits, hyphen, underscore; must start with a letter
- **Password**: 6–255 chars; at least 1 uppercase letter, 1 lowercase letter, 1 digit

**Independent Test**: Register a test account via curl against auth-svc → log in via the app on Android emulator → verify JWT is stored and home screen is reached.

### Tests for US1 (write first, verify FAIL before T031)

- [X] T028 [P] [US1] Unit tests for `AuthValidator` in `mobile/shared/src/commonTest/domain/auth/AuthValidatorTest.kt`: email format (valid, missing @, empty), username (valid, uppercase rejected, slash rejected, too short), password (valid, too short, no uppercase, no digit, max 255)
- [X] T029 [P] [US1] Unit tests for `LoginStore` state transitions in `mobile/shared/src/commonTest/presentation/auth/LoginStoreTest.kt`: idle → loading → success, idle → loading → error(InvalidCredentials)

### Implementation for US1

- [X] T030 [P] [US1] Implement `AuthValidator` in `mobile/shared/src/commonMain/domain/auth/AuthValidator.kt`: functions `validateEmail(email: String): ValidationResult`, `validateUsername(username: String): ValidationResult`, `validatePassword(password: String): ValidationResult`
- [X] T031 [P] [US1] Implement `AuthRepository` interface in `mobile/shared/src/commonMain/domain/auth/AuthRepository.kt`: `suspend fun login(emailOrUsername: String, password: String): Result<AuthSession>`, `suspend fun logout()`, `fun isLoggedIn(): Boolean`
- [X] T032 [US1] Implement `TokenStore` in `mobile/shared/src/commonMain/data/auth/TokenStore.kt`: SQLDelight-backed storage for `access_token`, `refresh_token`, `user_id`, `user_email`; expect/actual `SecureStorage` delegate for platform keychain
- [X] T033 [P] [US1] Implement `SecureStorage` actual for Android in `mobile/androidApp/src/main/.../data/auth/SecureStorageAndroid.kt` using Android `EncryptedSharedPreferences`
- [X] T034 [P] [US1] Implement `SecureStorage` actual for iOS in `mobile/iosApp/.../data/auth/SecureStorageIos.kt` using iOS Keychain via `Security` framework
- [X] T035 [US1] Implement `AuthApiClient` in `mobile/shared/src/commonMain/data/auth/AuthApiClient.kt`: `suspend fun login(request: LoginRequest): LoginResponse`, maps HTTP errors to `AppError`
- [X] T036 [US1] Implement `AuthRepositoryImpl` in `mobile/shared/src/commonMain/data/auth/AuthRepositoryImpl.kt`: calls `AuthApiClient`, persists tokens to `TokenStore`
- [X] T037 [US1] Implement `LoginStore` (MVIKotlin) in `mobile/shared/src/commonMain/presentation/auth/LoginStore.kt`: Intent(`UpdateEmail`, `UpdateUsername`, `UpdatePassword`, `Submit`, `ToggleOtpMode`), State(`email`, `password`, `isLoading`, `error`)
- [X] T038 [US1] Implement `LoginScreen` composable in `mobile/shared/src/commonMain/presentation/auth/LoginScreen.kt`: email/username field, password field with show/hide toggle, inline validation error text, login button, "Sign in with email code" link, "Create account" link
- [X] T039 [US1] Implement `AuthComponent` (Decompose) in `mobile/shared/src/commonMain/presentation/auth/AuthComponent.kt`: manages child stack (Login → Register → VerifyEmail → OtpLogin)
- [X] T040 [US1] Implement `RootComponent` in `mobile/shared/src/commonMain/presentation/RootComponent.kt`: on app start checks `TokenStore.isLoggedIn()` → route to `AuthComponent` or `MainComponent` (home)

**Checkpoint**: Login screen renders on Android emulator; entering invalid email shows error without network call; valid credentials reach the backend and return a JWT.

---

## Phase 4: User Registration Backend + Welcome Email (User items 3 + 5)

**User Story US2**: A student registers with email, username, and password. The backend validates all fields, creates their account, sends a welcome email, and triggers email verification.

**Independent Test**: `POST /auth/register` with valid body → `201` response → welcome email appears in Swoosh local mailbox → `POST /auth/email/verify` with the OTP → `204`.

### Tests for US2 (write first, verify FAIL before T046)

- [X] T041 [P] [US2] Unit tests for password domain in `backend/svc/auth/domain/password_test.go`: valid, too short (5 chars), no uppercase, no digit, too long (256 chars), exactly 6 chars passes, exactly 255 chars passes
- [ ] T042 [P] [US2] Integration test for `POST /auth/register` in `backend/svc/auth/handler/register_test.go` using testcontainers: success 201, duplicate email 409, weak password 422, missing fields 422
- [X] T043 [P] [US2] ExUnit test for `WelcomeEmail` in `backend/svc/mail/test/preuni/emails/welcome_email_test.exs`: rendered HTML contains greeting and verification link placeholder
- [ ] T044 [P] [US2] Integration test for `POST /auth/email/verify` in `backend/svc/auth/handler/verify_email_test.go`: valid OTP 204, expired OTP 400, wrong OTP 400, already verified 409

### Implementation for US2

- [X] T045 [US2] Implement `backend/svc/auth/domain/credentials.go`: `Credentials` struct, `NewCredentials(email, password string) (*Credentials, error)` — validates password via domain rules, hashes with bcrypt (cost 12)
- [X] T046 [US2] Implement `backend/svc/auth/domain/password.go`: `ValidatePassword(password string) error` — checks length 6–255, at least 1 uppercase, 1 lowercase, 1 digit; returns typed `ValidationError` with field and message
- [X] T047 [US2] Implement `backend/svc/auth/repository/credentials.go` (pgx): `Create(ctx, creds)`, `FindByEmail(ctx, email)`, `MarkEmailVerified(ctx, id)`, `UpdatePasswordHash(ctx, id, hash)`, `Anonymize(ctx, id)`
- [X] T048 [US2] Implement `backend/svc/auth/handler/register.go`: parse body → validate fields → call `credentials.NewCredentials` → save to DB → call user-svc internal `POST /internal/students` → send welcome + verification email via mail-svc → return 201
- [X] T049 [US2] Implement user-svc internal endpoint `POST /internal/students` in `backend/svc/user/handler/create_student.go`: creates `students` row with `id` from auth-svc; auth-svc passes `student_id`, `display_name`, `email`
- [X] T050 [US2] Implement `backend/svc/user/repository/student.go` (pgx): `Create(ctx, student)`, `FindByID(ctx, id)`, `Update(ctx, id, patch)`, `Anonymize(ctx, id)`
- [X] T051 [US2] Implement OTP domain in `backend/svc/auth/domain/otp.go`: `GenerateOTP() string` (6-digit numeric), `HashOTP(raw string) string` (SHA-256 hex)
- [X] T052 [US2] Implement `backend/svc/auth/repository/otp.go` (pgx): `Create(ctx, credentialID, purpose, codeHash, expiresAt)`, `FindActiveByCredentialAndPurpose(ctx, credentialID, purpose)`, `MarkUsed(ctx, id)`
- [X] T053 [US2] Implement `backend/svc/auth/handler/verify_email.go`: look up OTP by email+purpose `EMAIL_VERIFY` → validate not expired, not used → mark used → mark credential email_verified → return 204
- [X] T054 [US2] Implement mail-svc `WelcomeEmail` in `backend/svc/mail/lib/preuni/emails/welcome_email.ex`: uses Swoosh `new_email` with `to`, `subject`, and `html_body` rendered from template `welcome_email.html.heex`
- [X] T055 [US2] Implement mail-svc `VerificationEmail` in `backend/svc/mail/lib/preuni/emails/verification_email.ex`: template includes OTP code and 15-minute expiry notice
- [X] T056 [US2] Implement mail-svc `POST /internal/email/send` in `backend/svc/mail/lib/preuni_web/controllers/email_controller.ex`: accepts `{type, to, params}` JSON; dispatches to correct email module; validates `INTERNAL_TOKEN` header
- [X] T057 [US2] Add `RegisterScreen` composable in `mobile/shared/src/commonMain/presentation/auth/RegisterScreen.kt`: display_name field, email field, username field (with `AuthValidator.validateUsername` inline), password field (with strength indicator), confirm-password field; submits to `AuthComponent`
- [X] T058 [US2] Add `VerifyEmailScreen` composable in `mobile/shared/src/commonMain/presentation/auth/VerifyEmailScreen.kt`: 6-digit OTP input, resend code link (rate-limited: 60 s), submit button

**Checkpoint**: Full register → verify email flow works end-to-end; welcome email visible in `http://localhost:4000/dev/mailbox`; `POST /auth/login` with new account returns 200.

---

## Phase 5: OTP Login with Mail Service (User item 4)

**User Story US3**: A student who forgot their password (or prefers passwordless login) requests a one-time code sent to their email and uses it to sign in.

**Independent Test**: Request OTP for a verified email → code appears in Swoosh local mailbox → POST `/auth/otp/verify` with code → 200 with JWT → KMP app navigates to home.

### Tests for US3 (write first, verify FAIL before T063)

- [ ] T059 [P] [US3] Integration test for OTP login flow in `backend/svc/auth/handler/otp_login_test.go` (testcontainers): request OTP for unknown email → 202 (no enumeration); request for known email → 202; verify valid code → 200 JWT; verify expired code → 400; verify used code → 400

### Implementation for US3

- [X] T060 [US3] Implement `backend/svc/auth/handler/otp_login_request.go` (`POST /auth/otp/request`): always returns 202 (prevents email enumeration); if email exists and is verified: generates OTP, stores with purpose `LOGIN_OTP`, calls mail-svc to send code
- [X] T061 [US3] Implement mail-svc `OtpLoginEmail` in `backend/svc/mail/lib/preuni/emails/otp_login_email.ex`: template shows code prominently, expires-in notice, security disclaimer ("if you didn't request this, ignore it")
- [X] T062 [US3] Implement `backend/svc/auth/handler/otp_login_verify.go` (`POST /auth/otp/verify`): finds OTP by email+purpose `LOGIN_OTP` → validates not expired, not used → marks used → issues access + refresh token → returns 200
- [X] T063 [US3] Add `OtpLoginScreen` composable in `mobile/shared/src/commonMain/presentation/auth/OtpLoginScreen.kt`: email input step → OTP input step (6 digits, large font); back navigation to `LoginScreen`
- [X] T064 [US3] Wire OTP login into `AuthComponent` child stack in `mobile/shared/src/commonMain/presentation/auth/AuthComponent.kt`: `LoginScreen` "Sign in with email code" link pushes `OtpLoginScreen`

**Checkpoint**: Tapping "Sign in with email code" → entering email → entering code from mailbox → arrives at home screen.

---

## Phase 6: User Management Backend (User item 6)

**User Story US4**: Authenticated student can view and update their profile (display name, username), change password, request email change, confirm email change, upload avatar, and delete their account with full GDPR erasure.

**Independent Test**: Authenticated request to `GET /students/me` returns profile; `PATCH /students/me` with new username updates it; `DELETE /students/me` anonymizes data; all confirmed via DB inspection.

### Tests for US4 (write first, verify FAIL before T072)

- [X] T065 [P] [US4] Unit tests for username validation in `backend/svc/user/domain/username_test.go`: valid (`ana_01`, `jo-se`), uppercase rejected, starts with digit rejected, too short (< 3), too long (> 30), contains `/` rejected
- [ ] T066 [P] [US4] Integration tests for `GET /students/me` and `PATCH /students/me` in `backend/svc/user/handler/student_test.go` (testcontainers): unauthenticated 401; own profile 200; update username to existing one 409; update with invalid username 422
- [ ] T067 [P] [US4] Integration tests for account deletion in `backend/svc/auth/handler/delete_account_test.go`: valid auth → 204; student row anonymized (display_name = "Deleted User", email nulled); credentials anonymized; all refresh tokens revoked

### Implementation for US4

- [X] T068 [US4] Implement `backend/svc/user/domain/username.go`: `ValidateUsername(username string) error` — regex `^[a-z][a-z0-9_-]{1,28}[a-z0-9]$`, length 3–30, returns typed `ValidationError`
- [X] T069 [P] [US4] Implement `backend/svc/user/handler/get_student.go` (`GET /students/me`): reads `X-User-ID` from context → returns student row; 404 if not found
- [X] T070 [P] [US4] Implement `backend/svc/user/handler/update_student.go` (`PATCH /students/me`): accepts `{display_name?, username?}` → validates username via domain → checks username uniqueness → updates; returns updated student
- [X] T071 [US4] Implement `backend/svc/auth/handler/change_password.go` (`POST /auth/password/change`): requires auth; accepts `{current_password, new_password}`; verifies current via bcrypt; validates new via domain; hashes + saves; revokes all refresh tokens except current
- [X] T072 [US4] Implement `backend/svc/auth/handler/change_email.go` (`POST /auth/email/change-request` + `POST /auth/email/change-confirm`): request generates OTP stored with purpose `EMAIL_CHANGE`, sends to NEW email; confirm verifies OTP → updates email in credentials → revokes all sessions
- [X] T073 [US4] Implement mail-svc `EmailChangeEmail` in `backend/svc/mail/lib/preuni/emails/email_change_email.ex`: template includes OTP code, clearly states "verify your NEW email address"
- [X] T074 [US4] Implement password reset flow in `backend/svc/auth/handler/password_reset.go` (`POST /auth/password/reset-request` + `POST /auth/password/reset`): request sends OTP with purpose `PASSWORD_RESET`; reset validates OTP → updates hash; re-use existing `VerificationEmail` or dedicated template
- [X] T075 [US4] Implement `backend/svc/user/handler/avatar.go` (`PUT /students/me/avatar`): generates S3 presigned PUT URL for `avatars/{user_id}/{uuid}.webp`; after client uploads, `POST /students/me/avatar/confirm` updates `avatar_url` in students row
- [X] T076 [US4] Implement GDPR account deletion in `backend/svc/user/handler/delete_student.go` (`DELETE /students/me`): set `display_name = 'Deleted User'`, `avatar_url = null`, `xp_total = 0`, `streak_count = 0`; does NOT delete the row (preserves referential integrity for learning/simulation data)
- [X] T077 [US4] Implement `backend/svc/auth/handler/delete_account.go` (`DELETE /auth/account`): atomically — set `credentials.email = 'deleted+{id}@preuni.com.br'`, `password_hash = ''`, `email_verified = false`; delete all `refresh_tokens` for credential; call user-svc anonymization endpoint; return 204

**Checkpoint**: Authenticated student can change their username, change their password (re-login required), receive email-change OTP, and delete account (DB shows anonymized values).

---

## Phase 7: User Management Frontend (User items 7 + 8)

**User Story US5**: Student manages their profile from within the app: edit display name, change username (validation enforced client-side), change password, change email with OTP confirmation to the new email, upload a profile photo, and delete account with GDPR-aligned disclosure.

**Independent Test**: Tap Profile tab → change username to invalid value → see inline error → change to valid value → success toast; tap Delete Account → read GDPR disclosure → confirm → redirected to login screen.

### Tests for US5 (write first, verify FAIL before T082)

- [X] T078 [P] [US5] Unit tests for `ProfileStore` in `mobile/shared/src/commonTest/presentation/profile/ProfileStoreTest.kt`: initial load emits student data; username update intent runs validator before network call; delete account intent shows confirmation state before dispatch

### Implementation for US5

- [X] T079 [P] [US5] Implement `UserRepository` interface in `mobile/shared/src/commonMain/domain/user/UserRepository.kt` + `UserApiClient` in `mobile/shared/src/commonMain/data/user/UserApiClient.kt` + `UserRepositoryImpl` in `mobile/shared/src/commonMain/data/user/UserRepositoryImpl.kt`
- [X] T080 [US5] Implement `ProfileStore` (MVIKotlin) in `mobile/shared/src/commonMain/presentation/profile/ProfileStore.kt`: Intent(`LoadProfile`, `UpdateDisplayName`, `UpdateUsername`, `ChangePassword`, `ChangeEmail`, `ConfirmEmailChange`, `UploadAvatar`, `DeleteAccount`, `ConfirmDeleteAccount`); validates username + password client-side before dispatching
- [X] T081 [US5] Implement `ProfileScreen` composable in `mobile/shared/src/commonMain/presentation/profile/ProfileScreen.kt`: avatar with edit overlay, display name, username, email; action buttons for Change Password, Change Email, Delete Account; XP and streak shown read-only
- [X] T082 [US5] Implement `EditUsernameScreen` in `mobile/shared/src/commonMain/presentation/profile/EditUsernameScreen.kt`: text field with live validation (regex `^[a-z][a-z0-9_-]{1,28}[a-z0-9]$`), error hint "Only lowercase letters, numbers, - and _", character counter (30 max)
- [X] T083 [US5] Implement `EditPasswordScreen` in `mobile/shared/src/commonMain/presentation/profile/EditPasswordScreen.kt`: current password field, new password field with strength bar, confirm field; submit disabled until all three fields pass validation
- [X] T084 [US5] Implement `ChangeEmailScreen` in `mobile/shared/src/commonMain/presentation/profile/ChangeEmailScreen.kt`: new email input → submit triggers OTP to new email; transitions to `ConfirmNewEmailScreen`
- [X] T085 [US5] Implement `ConfirmNewEmailScreen` in `mobile/shared/src/commonMain/presentation/profile/ConfirmNewEmailScreen.kt`: OTP 6-digit input, resend timer (60 s), instructional text "Enter the code sent to {newEmail}"
- [X] T086 [US5] Implement `AvatarPickerComponent` in `mobile/shared/src/commonMain/presentation/profile/AvatarPickerComponent.kt`: shows current avatar + pencil icon; tapping opens platform picker; selected image compressed to WebP before upload; calls `UserRepository.getAvatarUploadUrl()` → HTTP PUT to presigned URL → `confirmAvatarUpload()`
- [X] T087 [P] [US5] Implement `ImagePicker` expect/actual for Android in `mobile/androidApp/.../presentation/profile/ImagePickerAndroid.kt` using `ActivityResultContracts.PickVisualMedia`
- [X] T088 [P] [US5] Implement `ImagePicker` expect/actual for iOS in `mobile/iosApp/.../presentation/profile/ImagePickerIos.kt` using `PHPickerViewController`
- [X] T089 [US5] Implement `DeleteAccountScreen` in `mobile/shared/src/commonMain/presentation/profile/DeleteAccountScreen.kt`: GDPR disclosure text (what data is deleted, what is anonymized, note on legal retention), type-to-confirm field ("DELETE"), disabled confirm button until typed, on confirm calls `ProfileStore.Intent.ConfirmDeleteAccount`; on success clears `TokenStore` and navigates to `AuthComponent`
- [X] T090 [US5] Wire all profile screens into a Decompose child stack in `mobile/shared/src/commonMain/presentation/profile/ProfileComponent.kt` and register `ProfileComponent` in `MainComponent` (bottom nav)

**Checkpoint**: Full profile edit flow works on Android; avatar upload succeeds; delete account from app → app shows login screen → login with same email returns 401.

---

## Phase 8: Basic Home Screen After Login (User item 9)

**User Story US6**: After successful login (including first-time-ever login), the student sees a home screen showing their greeting, streak, XP total, and enrolled track placeholders with a bottom navigation bar.

**Independent Test**: Log in with a freshly registered account → home screen shows "Olá, {display_name}" + streak 0 + XP 0 + empty "Start learning" CTA; bottom nav shows Home, Aprender, Simular, Perfil tabs.

### Tests for US6 (write first, verify FAIL before T095)

- [X] T091 [P] [US6] Unit tests for `HomeStore` in `mobile/shared/src/commonTest/presentation/home/HomeStoreTest.kt`: successful load emits state with student data; network error emits error state; retry intent re-fetches

### Implementation for US6

- [X] T092 [US6] Implement `HomeStore` (MVIKotlin) in `mobile/shared/src/commonMain/presentation/home/HomeStore.kt`: Intent(`Load`, `Retry`); State(`student`, `isLoading`, `error`); calls `UserRepository.getMe()` on load
- [X] T093 [US6] Implement `HomeScreen` composable in `mobile/shared/src/commonMain/presentation/home/HomeScreen.kt`: greeting ("Olá, {display_name}!"), streak badge (flame icon + count), XP pill, "Continue studying" or "Start learning" CTA card (navigates to Aprender tab), readiness-score placeholder card (shown only when > 0)
- [X] T094 [US6] Implement `BottomNavigation` composable in `mobile/shared/src/commonMain/presentation/navigation/BottomNavigation.kt`: 4 tabs — Home (house icon), Aprender (book icon), Simular (pencil icon), Perfil (person icon); active tab highlighted using design system color token
- [X] T095 [US6] Implement `MainComponent` (Decompose) in `mobile/shared/src/commonMain/presentation/main/MainComponent.kt`: holds child instances for Home, Learn (placeholder), Simulate (placeholder), Profile; wires bottom nav selection to active child
- [X] T096 [US6] Implement placeholder `LearnScreen` in `mobile/shared/src/commonMain/presentation/learn/LearnScreen.kt`: "Em breve — subject tracks coming soon" empty state with illustration
- [X] T097 [US6] Implement placeholder `SimulateScreen` in `mobile/shared/src/commonMain/presentation/simulate/SimulateScreen.kt`: "Em breve — ENEM simulations coming soon" empty state with illustration
- [X] T098 [US6] Update `RootComponent` in `mobile/shared/src/commonMain/presentation/RootComponent.kt`: token check → if logged in and onboarding complete → `MainComponent`; if logged in and onboarding pending → `OnboardingComponent`; if not logged in → `AuthComponent`

**Checkpoint**: Login → home screen; bottom nav works between tabs; each placeholder tab shows "Em breve" state; profile tab shows real student data.

---

## Phase 9: First-Time Onboarding Flow (User item 10)

**User Story US7**: A first-time student sees 4 explanatory screens before reaching the home screen: what the app does, how spaced repetition works, the five subject tracks, and how dissertation practice works. On the final screen they select their starting tracks (at least one required).

**Independent Test**: Log in with a new account (onboarding_completed = false) → onboarding slides appear → select ≥ 1 track → tap "Começar" → `onboarding_completed` = true in DB → home screen reached; re-login skips onboarding.

### Tests for US7 (write first, verify FAIL before T103)

- [ ] T099 [P] [US7] Integration test for onboarding endpoint in `backend/svc/user/handler/onboarding_test.go`: unauthenticated 401; authenticated PATCH sets `onboarding_completed = true`; second call returns 200 idempotently
- [X] T100 [P] [US7] Unit tests for `OnboardingStore` in `mobile/shared/src/commonTest/presentation/onboarding/OnboardingStoreTest.kt`: next page intent advances page index; selecting tracks updates selected set; completing with 0 tracks emits validation error; completing with ≥ 1 tracks calls repository then emits Complete

### Implementation for US7

- [X] T101 [US7] Add migration `006_add_onboarding_completed.sql` in `infra/migrations/user/`: `ALTER TABLE students ADD COLUMN onboarding_completed BOOLEAN NOT NULL DEFAULT false`
- [X] T102 [US7] Implement `PATCH /students/me/onboarding` in `backend/svc/user/handler/onboarding.go`: sets `onboarding_completed = true` for the authenticated student; idempotent (returns 200 if already true); also accepts `{enrolled_track_ids: [uuid]}` and creates `track_enrollments` rows
- [X] T103 [US7] Implement `OnboardingStore` (MVIKotlin) in `mobile/shared/src/commonMain/presentation/onboarding/OnboardingStore.kt`: Intent(`NextPage`, `PreviousPage`, `ToggleTrack`, `Complete`); State(`pageIndex 0–3`, `selectedTrackIds`, `isLoading`, `error`); page 3 requires at least 1 track selected
- [X] T104 [US7] Implement `OnboardingScreen` composable in `mobile/shared/src/commonMain/presentation/onboarding/OnboardingScreen.kt`: `HorizontalPager` with 4 pages; page indicator dots; back button (hidden on page 0); next/complete button; swipe gesture supported
  - Page 0 — "Sua aprovação começa aqui": ENEM explanation, app purpose
  - Page 1 — "Aprenda de verdade": spaced repetition explanation with visual of interval growth
  - Page 2 — "Cinco trilhas, um objetivo": list of 5 ENEM areas with icons and brief descriptions
  - Page 3 — "A redação importa": dissertation scoring explained (Competências 1–5)
- [X] T105 [US7] Implement `TrackSelectionScreen` composable in `mobile/shared/src/commonMain/presentation/onboarding/TrackSelectionScreen.kt`: loads tracks from `content-svc` via `ContentRepository`; multi-select chips per track with color tokens; "Começar" button disabled if none selected
- [X] T106 [US7] Implement `ContentRepository` interface + Ktor impl in `mobile/shared/src/commonMain/domain/content/ContentRepository.kt` and `mobile/shared/src/commonMain/data/content/ContentRepositoryImpl.kt` (only `getTracks()` for now)
- [X] T107 [US7] Implement `OnboardingComponent` (Decompose) in `mobile/shared/src/commonMain/presentation/onboarding/OnboardingComponent.kt`: owns `OnboardingStore`; on Complete event navigates `RootComponent` to `MainComponent`

**Checkpoint**: Fresh login → slides appear → all 4 pages swipeable → select Mathematics track → tap "Começar" → DB shows `onboarding_completed = true` and one `track_enrollments` row → home screen reached; second login skips onboarding.

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Hardening that spans all user stories above.

- [X] T108 [P] Implement token auto-refresh in `mobile/shared/src/commonMain/data/network/TokenRefreshInterceptor.kt`: on 401 response, call `POST /auth/refresh` with stored refresh token, retry original request once; on second 401 clear `TokenStore` and emit `AppError.Unauthorized` to trigger logout
- [X] T109 [P] Implement `GlobalErrorHandler` composable in `mobile/shared/src/commonMain/presentation/common/GlobalErrorHandler.kt`: snackbar for network errors, full-screen for session-expired with "Sign in again" CTA; consumed by `RootComponent`
- [X] T110 [P] Add NGINX rate limiting to `infra/nginx/nginx.conf`: define `limit_req_zone $binary_remote_addr zone=auth_zone:10m rate=100r/m`; apply to `/v1/auth/register`, `/v1/auth/login`, `/v1/auth/otp/*`, `/v1/auth/password/*`
- [X] T111 [P] Add GDPR data export stub in `backend/svc/user/handler/data_export.go` (`GET /students/me/data-export`): returns JSON with all student rows the user owns (profile, xp_events, track_enrollments, achievements); response includes `Content-Disposition: attachment` header
- [X] T112 [P] Add `POST /auth/logout` handler in `backend/svc/auth/handler/logout.go`: revokes the specific refresh token identified by `jti` claim in the access token; returns 204
- [X] T113 Run full stack validation per `quickstart.md`: confirm all migrations apply, all services start, register → verify → login → onboarding → home flow completes without errors

**Checkpoint**: 401 with expired token → auto-refreshed seamlessly; NGINX returns 429 on rapid auth requests; data export downloads valid JSON; logout revokes token.

---

## Phase 11: Git Repository Init & Versioning

**Purpose**: Establish git history hygiene, conventional commit standards, changelog generation, and semantic versioning after the primary Sprint 1 implementation is complete. These tasks transform the working directory into a properly versioned repository ready for CI/CD.

**Dependency**: All Phase 1–10 tasks must be complete (or at a stable checkpoint) before committing history.

- [X] T114 [P] Verify `.gitignore` covers all generated artifacts: `build/`, `.gradle/`, `*.class`, `_build/`, `deps/`, `.mix/`, `*.beam`, Go binaries, `node_modules/`, `.DS_Store`, `.env*`, `infra/secrets/`; update if any patterns are missing
- [X] T115 [P] Add `commitlint.config.js` at repo root: enforce Conventional Commits (`feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `ci`) with scope rules matching service names (`auth`, `user`, `mail`, `mobile`, `infra`)
- [X] T116 [P] Add `.husky/` pre-commit hook: runs `ktlint --format` on staged `.kt` files and `golangci-lint run` on staged `.go` files; blocks commit if either fails
- [X] T117 [P] Add `.husky/commit-msg` hook: runs `commitlint` to validate message format before commit is recorded
- [X] T118 Create `CHANGELOG.md` at repo root using `git-cliff` config in `.cliff.toml`: groups by type (`feat` → Features, `fix` → Bug Fixes, `chore` → Maintenance), links to commit hashes, marks breaking changes with ⚠️
- [X] T119 Add `version.txt` at repo root with initial version `0.1.0`; add `Makefile` target `bump-version` that: reads `version.txt`, increments semver (patch/minor/major based on arg), updates `version.txt`, updates `mobile/gradle/libs.versions.toml` `appVersion` entry, and creates a git tag `v{new_version}`
- [X] T120 Create the initial git commit: stage all implementation files (exclude secrets and generated artifacts); write commit message `feat(init): bootstrap Sprint 1 — auth, onboarding, home`; this becomes the `v0.1.0` baseline
- [X] T121 [P] Document versioning workflow in `specs/001-enem-prep-platform/quickstart.md` under a new "Release Process" section: how to run `make bump-version patch`, how to generate CHANGELOG, and how to push tags

**Checkpoint**: `git log --oneline` shows a clean initial commit; `git tag` shows `v0.1.0`; `git cliff --current` generates a valid CHANGELOG entry; staging a badly formatted commit message is rejected by the commit-msg hook.

---

## Phase 12: Go Service Routing

**Goal**: Wire all implemented handlers into the chi routers so Go services actually serve real traffic (not just `/health`).

**Dependency**: Phase 10 complete (all handlers exist).

**Checkpoint**: `cd backend/svc/auth && PORT=8081 DATABASE_URL=... go run ./cmd/server`; `curl -X POST http://localhost:8081/v1/auth/register -d '{}'` returns `400` (route exists, body validation fails — proves route is registered).

- [X] T122 Wire all auth-svc handlers into chi router in `backend/svc/auth/cmd/server/main.go`: open pgx connection pool from `DATABASE_URL` env var, create Redis client from `REDIS_URL`, instantiate all repositories (`CredentialsRepository`, `OTPRepository`, `RefreshTokenRepository`), instantiate all handlers, register routes under `/v1/auth`: `POST /register`, `POST /login`, `POST /email/verify`, `POST /otp/request`, `POST /otp/verify`, `POST /token/refresh`, `POST /logout`, `POST /password/change`, `POST /password/reset/request`, `POST /password/reset/confirm`, `POST /email/change/request`, `POST /email/change/confirm`, `DELETE /account`; apply `pkg/middleware` JWT auth where required; add `/health` endpoint
- [X] T123 [P] Wire all user-svc handlers into chi router in `backend/svc/user/cmd/server/main.go`: open pgx pool, instantiate `StudentRepository`, instantiate all handlers, apply JWT middleware from `pkg/middleware`, register routes: `POST /internal/students`, `GET /v1/students/me`, `PATCH /v1/students/me`, `GET /v1/students/me/avatar/upload-url`, `POST /v1/students/me/avatar/confirm`, `PATCH /v1/students/me/onboarding`, `GET /v1/students/me/data-export`; add `/health`

---

## Phase 13: Mobile App Entry Points

**Goal**: Create the minimal Kotlin/Swift composable root and per-platform entry points so each target can boot and render the app.

**Checkpoint**: `cd mobile && ./gradlew :webApp:wasmJsBrowserDevelopmentRun` opens `http://localhost:8080` and the login screen renders.

- [X] T124 Create shared composable root in `mobile/shared/src/commonMain/kotlin/com/preuni/shared/PreuniApp.kt`: `@Composable fun PreuniApp(component: RootComponent)` that subscribes to `component.childStack` and renders `AuthContent`, `OnboardingContent`, or `MainContent` depending on the active child; each content function delegates to the corresponding screen/component composable
- [X] T125 [P] Create `mobile/webApp/src/wasmJsMain/kotlin/main.kt`: call `onWasmReady { ComposeViewport(document.body!!) { val root = remember { createDefaultRootComponent() }; PreuniApp(root) } }`; create `mobile/webApp/src/wasmJsMain/kotlin/di/WebAppModule.kt` with Koin module providing wasmJs-specific `HttpClient` (base URL from `js("window.location.origin")`) and `SqlDriver`; call `startKoin { modules(WebAppModule) }` before creating `RootComponent`
- [X] T126 [P] Create `mobile/androidApp/src/main/kotlin/com/preuni/android/MainActivity.kt`: `ComponentActivity` subclass, `setContent { val root = remember { DefaultRootComponent(defaultComponentContext()) }; PreuniApp(root) }`; create `mobile/androidApp/src/main/AndroidManifest.xml` declaring `MainActivity` as launcher; create `mobile/androidApp/src/main/res/values/themes.xml` inheriting `Theme.MaterialComponents.DayNight.NoActionBar`; create `mobile/androidApp/src/main/res/drawable/ic_launcher_background.xml` as plain color placeholder

---

## Phase 14: Dockerfiles & Docker Compose

**Goal**: Each Go microservice and the Elixir mail service can be containerized. `docker compose up` boots the full stack.

**Dependency**: Phase 12 complete (handlers wired — binaries must compile).

**Checkpoint**: `docker compose -f infra/docker-compose.yml build auth && docker compose up -d postgres redis auth && curl http://localhost:8081/health` returns `ok`.

- [X] T127 Create `backend/svc/auth/Dockerfile` — two-stage build: stage 1 `golang:1.23-alpine` AS builder, `WORKDIR /build`, `COPY go.work go.work.sum ./`, `COPY pkg/ ./pkg/`, `COPY svc/auth/ ./svc/auth/`, `RUN cd svc/auth && CGO_ENABLED=0 go build -o /bin/auth-server ./cmd/server`; stage 2 `FROM gcr.io/distroless/static-debian12`, `COPY --from=builder /bin/auth-server /auth-server`, `EXPOSE 8081`, `ENTRYPOINT ["/auth-server"]`; add `backend/svc/auth/.dockerignore` excluding `*_test.go` and `vendor/`; update `infra/docker-compose.yml` auth service to set `build.context: ../backend` and `build.dockerfile: svc/auth/Dockerfile`
- [X] T128 [P] Create `backend/svc/user/Dockerfile` — same distroless two-stage pattern, binary name `user-server`, port 8082; update docker-compose user service build config with `context: ../backend` and `build.dockerfile: svc/user/Dockerfile`
- [X] T129 [P] Create `backend/svc/content/Dockerfile` — binary `content-server`, port 8083; update docker-compose
- [X] T130 [P] Create `backend/svc/learning/Dockerfile` — binary `learning-server`, port 8084; update docker-compose
- [X] T131 [P] Create `backend/svc/simulation/Dockerfile` — binary `simulation-server`, port 8085; update docker-compose
- [X] T132 [P] Create `backend/svc/dissertation/Dockerfile` — binary `dissertation-server`, port 8086; update docker-compose
- [X] T133 [P] Create `backend/svc/notification/Dockerfile` — binary `notification-server`, port 8087; update docker-compose
- [X] T134 [P] Create `backend/svc/mail/Dockerfile`: `FROM elixir:1.17-otp-27-alpine`; `RUN mix local.hex --force && mix local.rebar --force`; `COPY mix.exs mix.lock ./`; `RUN MIX_ENV=prod mix deps.get --only prod`; `COPY lib/ ./lib/ config/ ./config/ priv/ ./priv/`; `RUN MIX_ENV=prod mix compile`; `EXPOSE 4000`; `ENV PHX_SERVER=true`; `CMD ["mix", "phx.server"]`; add `backend/svc/mail/.dockerignore`; add `mail` service entry in `infra/docker-compose.yml` with `build.context: ../backend/svc/mail`, `INTERNAL_TOKEN`, `SWOOSH_ADAPTER`, `PHX_HOST` env vars
- [X] T135 Audit and fix `infra/docker-compose.yml` completeness: (a) ensure all Go services have `depends_on` with `postgres: {condition: service_healthy}` and `redis: {condition: service_healthy}`; (b) add missing `INTERNAL_SERVICE_TOKEN` env var to every Go service; (c) add `restart: on-failure` to every service; (d) add `networks: [preuni]` to all services and define a shared bridge network `preuni` at the bottom of the file; (e) verify all port mappings match their `PORT` env var

---

## Phase 15: Elixir Local Installation

**Goal**: Elixir 1.17 is available locally on macOS so the mail service can be developed and run without Docker.

**Checkpoint**: `elixir --version` shows `Elixir 1.17.x`; `cd backend/svc/mail && mix deps.get && mix compile` exits 0; `mix phx.server` starts and `curl http://localhost:4000/health` returns 200.

- [ ] T136 Install Elixir on macOS: run `brew install elixir`; if Homebrew is absent first install it via the official script at `https://brew.sh`; after install run `elixir --version` and confirm 1.17+; run `cd backend/svc/mail && mix local.hex --force && mix local.rebar --force && mix deps.get && mix compile`; if `config/runtime.exs` is missing create it reading `INTERNAL_TOKEN`, `PORT`, `PHX_HOST`, `SECRET_KEY_BASE` from env with sensible dev defaults; verify `mix phx.server` starts without errors

---

## Phase 16: iOS Xcode Project (T005)

**Goal**: `mobile/iosApp/` becomes a buildable Xcode project that embeds the KMP shared framework so the iOS Simulator can run the app.

**Dependency**: Phase 13 complete (shared `PreuniApp` composable exists); Xcode 16+ installed.

**Checkpoint**: `cd mobile/iosApp && xcodebuild -scheme iosApp -destination 'platform=iOS Simulator,name=iPhone 16' build` exits 0 and the simulator shows the login screen.

- [X] T137 Create iOS Xcode project in `mobile/iosApp/`: (a) create `mobile/iosApp/iosApp.xcodeproj/project.pbxproj` — minimal Swift target, deployment target iOS 16, bundle ID `com.preuni.app`, embed the KMP XCFramework produced by `./gradlew :shared:assembleXCFramework` from `mobile/build/XCFrameworks/release/shared.xcframework`; (b) create `mobile/iosApp/iosApp.xcodeproj/project.xcworkspace/contents.xcworkspacedata`; (c) create `mobile/iosApp/iosApp/iOSApp.swift` with `@main struct iOSApp: App { var body: some Scene { WindowGroup { ContentView() } } }`; (d) create `mobile/iosApp/iosApp/ContentView.swift` wrapping `MainViewController` from the shared framework via `UIViewControllerRepresentable`; (e) add `embedAndSignAppleFrameworkForXcode` as a pre-build script phase in the Xcode project; (f) create `mobile/iosApp/iosApp/Info.plist` with display name `Preuni` and required fields

---

## Phase 17: Makefile Run Targets

**Goal**: Developer can start any platform target and the backend with a single `make` command.

**Checkpoint**: `make help` lists all new targets; `make run-infra` starts postgres+redis; `make run-web` opens the Kotlin/Wasm app in a browser.

- [X] T138 Add `run-infra` and `stop-infra` targets to `Makefile`: `run-infra` runs `docker compose -f infra/docker-compose.yml up -d postgres redis`; `stop-infra` runs `docker compose -f infra/docker-compose.yml stop postgres redis`; update `help` target regex to pick up new `## ` prefixed comments
- [X] T139 [P] Add `run-backend` and `stop-backend` targets to `Makefile`: `run-backend` runs `docker compose -f infra/docker-compose.yml up --build -d auth user content learning simulation dissertation notification gateway`; `stop-backend` stops those services; add `## run-backend` doc comment
- [X] T140 [P] Add `run-web` target to `Makefile`: `cd mobile && ./gradlew :webApp:wasmJsBrowserDevelopmentRun`; add prerequisite comment noting JDK 17+ and Node.js 20+ are required
- [X] T141 [P] Add `run-android` target to `Makefile`: `cd mobile && ./gradlew :androidApp:installDebug`; add comment noting a connected device or running emulator is required; add `build-android` target: `./gradlew :androidApp:assembleDebug`
- [X] T142 [P] Add `run-ios` target to `Makefile`: `cd mobile && ./gradlew :shared:assembleXCFramework && cd iosApp && xcodebuild -scheme iosApp -destination "platform=iOS Simulator,name=iPhone 16" -allowProvisioningUpdates build`; add `open-ios` shortcut that opens `mobile/iosApp/iosApp.xcodeproj` in Xcode; add comment noting Xcode 16+ is required

---

## Phase 18: Git Remote Setup & Sync

**Goal**: The local branch history is pushed to `git@github.com:dwbessa/preuni.com.br.git` and the remote mirrors local.

**Checkpoint**: `git remote -v` shows `origin  git@github.com:dwbessa/preuni.com.br.git`; `git push` succeeds; GitHub shows all commits on `001-enem-prep-platform`.

- [ ] T143 Add the git remote: `git remote add origin git@github.com:dwbessa/preuni.com.br.git`; verify SSH access with `ssh -T git@github.com` (expect "Hi dwbessa!"); push the feature branch: `git push -u origin 001-enem-prep-platform`
- [ ] T144 [P] Push `main` branch: `git push -u origin main`; if `main` does not exist locally create it pointing to the initial commit: `git branch main $(git rev-list --max-parents=0 HEAD)` then push; confirm on GitHub that both `main` and `001-enem-prep-platform` appear under branches

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 ✅ — **BLOCKS all user story phases**
- **Phase 3–9 (User Stories)**: All depend on Phase 2 completion
  - Phase 3 (US1 Login KMP) can start as soon as Phase 2 is done
  - Phase 4 (US2 Registration Backend) can run **in parallel with Phase 3** (different files)
  - Phase 5 (US3 OTP Login) depends on Phase 4 (OTP repository and mail-svc ready)
  - Phase 6 (US4 User Mgmt Backend) depends on Phase 4 (student repository ready)
  - Phase 7 (US5 User Mgmt Frontend) depends on Phase 6 (backend endpoints ready)
  - Phase 8 (US6 Home Screen) depends on Phase 3 (auth KMP) + Phase 4 (user-svc `GET /students/me`)
  - Phase 9 (US7 Onboarding) depends on Phase 8 (MainComponent exists) + Phase 6 (track enrollment endpoint)
- **Phase 10 (Polish)**: Depends on all user story phases complete
- **Phase 11 (Git Versioning)**: Depends on Phase 10 (or a stable checkpoint) — run once the primary implementation is at a good state
- **Phase 12 (Go Routing)**: Depends on Phase 10 (all handlers exist); T122 and T123 are independent [P]
- **Phase 13 (Mobile Entry Points)**: Depends on Phase 3–9 (all stores and components exist); T124 must precede T125/T126
- **Phase 14 (Dockerfiles)**: Depends on Phase 12 (services must compile with handlers wired); all Dockerfile tasks [P] once T127 pattern is established
- **Phase 15 (Elixir Install)**: No code dependency — can run any time on the developer machine
- **Phase 16 (iOS Xcode)**: Depends on Phase 13 (PreuniApp composable must exist); requires Xcode 16+ installed
- **Phase 17 (Makefile Targets)**: Depends on Phases 12–16 being complete (targets invoke these systems); T138–T142 are independent [P]
- **Phase 18 (Git Remote)**: No code dependency — can run immediately; T143 must precede T144

### Within Each Phase

- Test tasks MUST be written first and verified to FAIL before paired implementation tasks
- Models/domain before repositories before handlers
- Backend endpoints before frontend API clients
- API clients before stores before screens

### Parallel Opportunities Per Phase

**Phase 3 + Phase 4** can proceed simultaneously on two tracks:
- Track A: KMP Login UI (Phase 3)
- Track B: Go registration backend (Phase 4)

**Within Phase 4** — these are independent after T045/T046:
- `credentials.go` domain (T045, T046) → Go registration handler
- user-svc `create_student.go` (T049, T050)
- Elixir mail templates (T054, T055, T056)

---

## Parallel Example: Phase 4 (Registration)

```
# Start simultaneously after Phase 2 complete:

Track Backend-Auth:
  T041 tests → T045 domain → T046 password domain → T047 repo → T042 tests → T048 handler

Track Backend-User:
  T049 user handler → T050 user repo

Track Mail-Elixir:
  T043 tests → T054 WelcomeEmail → T055 VerificationEmail → T056 email controller

Track Frontend-KMP:
  T057 RegisterScreen → T058 VerifyEmailScreen
  (blocks on T048 handler being complete for integration)
```

---

## Implementation Strategy

### MVP (Phases 1–3 + minimal Phase 4)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: Login in KMP → student can sign in ✅
4. Minimal Phase 4: Registration backend only (no email yet) → student can register ✅
5. **Stop and validate**: Register + login works end-to-end on Android emulator

### Full Sprint 1 Delivery

After MVP: complete Phases 4–9 in order, validating each independently.

### Task Count Summary

| Phase | Tasks | User Item |
|-------|-------|-----------|
| 1 Setup | 13 | Item 1 |
| 2 Foundational | 14 | — |
| 3 US1 Login KMP | 13 | Item 2 |
| 4 US2 Registration + Welcome Email | 18 | Items 3 + 5 |
| 5 US3 OTP Login | 6 | Item 4 |
| 6 US4 User Mgmt Backend | 13 | Item 6 |
| 7 US5 User Mgmt Frontend | 13 | Items 7 + 8 |
| 8 US6 Home Screen | 8 | Item 9 |
| 9 US7 Onboarding | 9 | Item 10 |
| 10 Polish | 6 | — |
| 11 Git Versioning | 8 | — |
| 12 Go Service Routing | 2 | — |
| 13 Mobile Entry Points | 3 | — |
| 14 Dockerfiles & Docker Compose | 9 | — |
| 15 Elixir Local Install | 1 | — |
| 16 iOS Xcode Project | 1 | — |
| 17 Makefile Run Targets | 5 | — |
| 18 Git Remote & Sync | 2 | — |
| **Total** | **144** | |

---

## Notes

- `[P]` tasks touch different files with no shared state — safe to run concurrently
- `[US#]` label enables filtering tasks by story for independent delivery
- Constitution requires tests to fail before implementation — do not skip this
- Commit after each task or logical group to enable easy rollback
- Username regex: `^[a-z][a-z0-9_-]{1,28}[a-z0-9]$` — if the intent was to allow forward slash `/`, revisit `AuthValidator` and `username.go` explicitly
- GDPR deletion: row is anonymized, not hard-deleted — preserves referential integrity for learning/simulation history while removing PII
