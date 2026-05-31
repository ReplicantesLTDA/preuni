# Feature Specification: Backend Monolith Refactor

**Feature Branch**: `009-backend-monolith-refactor`
**Created**: 2026-05-25
**Status**: Draft
**Input**: User description: "Refactor the backend into a monolith. Integrate email sending into the unified backend. Decide packaging approach later."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single backend deployment (Priority: P1)

As a platform operator, I can deploy and roll back all backend capabilities as one coordinated unit, so releases are simpler and incidents are easier to contain.

**Why this priority**: This reduces operational risk and coordination overhead across multiple backend components, improving reliability and release speed.

**Independent Test**: Deploy the unified backend to a staging environment and validate the platform end-to-end (sign-in, core learning flow, and email delivery) without coordinating multiple backend deployments.

**Acceptance Scenarios**:

1. **Given** the platform is running on the current backend setup, **When** the unified backend is deployed to a staging environment, **Then** core user flows work end-to-end without requiring separate deployments of multiple backend services.
2. **Given** a unified-backend release causes an incident in staging, **When** operators roll back to the previous release, **Then** core user flows recover and no user data is lost.

---

### User Story 2 - Faster local development (Priority: P2)

As a backend developer, I can run the backend locally as a single system for day-to-day development and debugging, so I spend less time orchestrating multiple services and more time shipping changes safely.

**Why this priority**: Developer productivity and onboarding speed are primary drivers for consolidating the backend.

**Independent Test**: A new developer can follow the local setup documentation to start the backend and complete at least one basic end-to-end flow (sign-in + fetch content).

**Acceptance Scenarios**:

1. **Given** a clean workstation, **When** a developer follows the local setup instructions, **Then** they can get the unified backend running and complete a basic end-to-end flow without manual coordination of multiple backend services.

---

### User Story 3 - Email continues to work (Priority: P3)

As an end user, I continue to receive transactional emails (such as verification, password reset, and notifications) as expected, even after the backend is consolidated.

**Why this priority**: Email is critical for account access and user trust; regressions are highly visible and costly.

**Independent Test**: Trigger each supported email type in staging and confirm delivery and content match current expectations.

**Acceptance Scenarios**:

1. **Given** a user requests a password reset, **When** the request is made against the unified backend, **Then** a password reset email is delivered and the reset flow can be completed successfully.
2. **Given** external email delivery is temporarily unavailable, **When** the unified backend attempts to send an email, **Then** the user-facing flow is not blocked and the failure is visible to operators for follow-up.

### Edge Cases

- Partial rollout where old and new backend versions coexist during a transition.
- Email delivery delays or duplicates caused by retries during outages.
- Load spikes that were previously spread across multiple backend components.
- One business capability becomes unavailable while others must remain usable.
- Permission regressions where a user gains or loses access due to consolidation errors.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a single deployable backend unit that delivers the full set of existing backend capabilities used by current clients (authentication, user management, content delivery, learning/simulation flows, essay/dissertation flows, notifications, and email sending). This is accepted when the platform can be validated end-to-end in a test environment using only the unified backend release.
- **FR-002**: Public behavior relied upon by existing clients MUST remain backward compatible so current released clients continue to function without requiring client updates. This is accepted when the latest released clients can sign in and complete a representative set of core learning flows without changes.
- **FR-003**: Transactional email sending MUST be provided by the unified backend, with all current email types supported and triggered under the same business conditions as today. This is accepted when each supported email type can be triggered and delivered successfully in a test environment.
- **FR-004**: Operators MUST be able to roll out and roll back the unified backend with minimal downtime and without user data loss. This is accepted when a rollout and rollback drill restores core user flows within the time limits in Success Criteria.
- **FR-005**: The system MUST preserve existing authentication, authorization, and permission rules across all backend capabilities. This is accepted when a predefined regression set of permission scenarios produces the same allow/deny outcomes as before.
- **FR-006**: The system MUST provide operational visibility for troubleshooting through clear error reporting and the ability to correlate user-impacting incidents to server-side diagnostics. This is accepted when operators can trace a reported failure from user impact to a concrete server-side error signal using existing operational tooling.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A full backend release to staging can be completed through a single coordinated deployment step and validated end-to-end within 30 minutes.
- **SC-002**: Rollback to the previous backend release can be completed in under 10 minutes with restoration of core user flows.
- **SC-003**: A new backend engineer can get the backend running locally and complete a basic end-to-end flow within 45 minutes using documented instructions.
- **SC-004**: Transactional email delivery success rate is at least 99% in staging validation runs and at least 99% during the first week after production rollout (excluding invalid recipient addresses).

## Assumptions

- No new user-facing features are introduced; the change is focused on consolidation and operational simplification.
- Existing client applications (mobile/web) are the compatibility baseline and must continue working as-is.
- Data storage and existing schemas remain unchanged for this feature unless a migration is strictly required for compatibility.
- The packaging format for the unified backend is out of scope for this specification and will be decided separately.
- The existing email delivery integration remains available and can be exercised in non-production environments for validation.
