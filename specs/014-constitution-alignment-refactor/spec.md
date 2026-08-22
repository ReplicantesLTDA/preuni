# Feature Specification: Constitution Alignment Refactor

**Feature Branch**: `014-constitution-alignment-refactor`
**Created**: 2026-08-22
**Status**: Draft
**Input**: User description: "refactor to achieve some goals: 1- app fully covers the constitution (essay driven, friends, rankings etc) 2- essay corrector and backend communicate correctly 3- fully working CI/CD with unit tests covering 90% of the code"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Submit an essay and see it graded (Priority: P1)

A student opens the app, picks (or is given) a prompt theme, submits an essay,
and later sees a structured grade broken into the five ENEM competencies with
explanations. Submitting also advances their daily streak.

**Why this priority**: This is the core loop the whole product pivot exists
for. Nothing else (friends, ranking, medals) has value without a working
submit → grade → streak path.

**Independent Test**: Can be fully tested by submitting one essay as a
free-tier user and confirming a graded result appears with all five
competencies and the user's streak count increments by one.

**Acceptance Scenarios**:

1. **Given** a free-tier user has not submitted today, **When** they submit an
   essay against a prompt theme, **Then** the submission is accepted
   immediately, their streak increments by one, and a "grading in progress"
   state is shown.
2. **Given** a submitted essay has finished grading, **When** the user opens
   it, **Then** they see a score plus, for each of the 5 ENEM competencies, a
   justification in Portuguese and at least one excerpt quoted from their own
   essay.
3. **Given** a free-tier user has already submitted once today, **When** they
   try to submit a second essay, **Then** the system blocks the submission
   and explains they've used today's free submission (with a Pro upsell).
4. **Given** a Pro user has already submitted once today, **When** they submit
   again, **Then** the system accepts it.
5. **Given** the correction service is slow or temporarily unavailable,
   **When** a user submits an essay, **Then** the submission is still
   accepted and queued — the user is never blocked or shown an error because
   of grading latency.

---

### User Story 2 - Add friends and see their progress (Priority: P2)

A user finds and adds another user as a friend, then sees that friend's
current streak and most recent essay grade on a shared view.

**Why this priority**: Social visibility is the mechanic that turns a solo
habit into a shared one; it depends on Story 1 existing (there must be
streaks/grades to show) but doesn't depend on ranking or medals.

**Independent Test**: Can be fully tested by two test accounts adding each
other and confirming each can see the other's streak count and latest grade
without seeing anything from non-friends.

**Acceptance Scenarios**:

1. **Given** two users are not yet friends, **When** user A sends a friend
   request and user B accepts, **Then** each appears in the other's friends
   list.
2. **Given** two users are friends, **When** user A opens their friends list,
   **Then** they see user B's current streak and most recent essay grade.
3. **Given** two users are not friends, **When** user A looks for user B's
   profile, **Then** user A cannot see user B's streak or grade history.
4. **Given** a user removes a friend, **When** the removal completes,
   **Then** neither party can see the other's streak/grades anymore.

---

### User Story 3 - Compete on the weekly ranking and earn medals (Priority: P3)

A user's graded essays from the current week accumulate into a global weekly
score. At the end of each week, users move up or down a league tier based on
their rank, and streak/ranking milestones award medals.

**Why this priority**: This is the retention/competitive layer on top of the
core loop and social graph — valuable, but the app is still usable without it
while Stories 1–2 are being hardened.

**Independent Test**: Can be fully tested by having several test accounts
submit graded essays across a week, then confirming the leaderboard orders
them correctly and the top/bottom bands are promoted/demoted at week-end.

**Acceptance Scenarios**:

1. **Given** a user has graded essays within the current week, **When** they
   view the leaderboard, **Then** they see their global rank and score
   alongside other users, grouped by league tier.
2. **Given** a weekly period ends, **When** the ranking is finalized,
   **Then** users in the top band of their tier are promoted to the next
   tier and users in the bottom band are demoted, matching a published rule.
3. **Given** a user reaches a streak milestone (e.g., 7-day, 30-day) or a
   ranking milestone (e.g., promoted to a new tier), **When** the milestone
   is reached, **Then** a medal is awarded and visible on their profile.
4. **Given** a new week begins, **When** the leaderboard resets, **Then**
   each user's weekly score starts at zero while their all-time streak is
   unaffected.

---

### User Story 4 - Ship changes safely and continuously (Priority: P1)

A contributor (human or AI-assisted) opens a pull request. Before it can
merge to the main line, automated checks run the full test suite across the
Go backend, the Python correction service, and the mobile app, confirm
coverage meets the required floor, and a human approves the change. Nothing
that fails these checks reaches the main line.

**Why this priority**: This underwrites every other story — without a
trustworthy CI/CD gate, regressions in the essay-grading pipeline, streak
math, or social/ranking logic can silently reach users. It's marked P1
because it's a precondition for safely shipping Stories 1–3, but it's listed
last because it has no independent end-user-visible behavior of its own.

**Independent Test**: Can be fully tested by opening a pull request that
introduces a failing test or drops coverage below the required floor, and
confirming the pull request is blocked from merging until fixed, and
separately by opening a fully passing pull request and confirming it cannot
merge without a recorded human approval.

**Acceptance Scenarios**:

1. **Given** a contributor commits code locally, **When** the commit is
   created, **Then** fast checks (unit tests, lint, type-check) run
   automatically and block the commit if they fail.
2. **Given** a pull request is opened against the main line, **When** the
   automated pipeline runs, **Then** it executes the full unit and
   integration test suites for every codebase touched by the change.
3. **Given** a pull request's test suite passes, **When** measured coverage
   for any touched codebase is below 90% or lower than before the change,
   **Then** the pipeline fails and merging is blocked.
4. **Given** a pull request has a failing required check, **When** someone
   attempts to merge it, **Then** the merge is blocked regardless of who
   authored the change.
5. **Given** a pull request has all checks green, **When** no human has
   approved it, **Then** the merge is still blocked.

---

### Edge Cases

- What happens when a user's essay fails grading entirely (e.g., unreadable
  input, correction service error after retries)? The user must be notified
  the grading failed and be able to resubmit without losing their streak
  credit for that day (streak already advanced on submission).
- What happens when a Pro subscription lapses mid-week? Remaining submissions
  for the current day beyond the free limit are not allowed once downgraded;
  already-submitted essays and their grades/streak/ranking contributions are
  unaffected.
- What happens at the UTC day boundary if a user submits at 23:59 and again
  at 00:01? These count as two different days for streak/quota purposes even
  if only two minutes apart.
- How does the system handle a friend request to a user who already removed
  the requester previously? Treated as a new request.
- What happens if two users are tied in weekly score at a promotion/demotion
  boundary? Tie-break rule must be defined and applied consistently (see
  Assumptions).
- What happens when the correction service is down during a CI run's
  integration tests? Integration tests must use a controlled test double or
  test instance, not the live service, so CI is not flaky due to external
  availability.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST let an authenticated user submit one essay per day
  against a prompt theme on the free tier, and multiple per day on the Pro
  tier, enforcing this quota before the essay reaches grading.
- **FR-002**: System MUST accept an essay submission immediately (never
  blocking on grading time) and asynchronously produce a structured grade
  covering all 5 ENEM competencies, each with a justification and at least
  one verbatim excerpt from the essay.
- **FR-003**: System MUST increment a user's daily streak at the moment of a
  valid, quota-accepted submission, using a fixed UTC day boundary, and MUST
  reset the streak to zero if a UTC day passes with no submission.
- **FR-004**: System MUST let a user send, accept, and remove friend
  connections, and MUST only expose another user's streak and latest grade
  to accepted friends — never to non-friends.
- **FR-005**: System MUST compute a weekly score per user from that week's
  graded essays and MUST expose a global leaderboard ranking users by that
  score within league tiers.
- **FR-006**: System MUST, at the end of each weekly period, promote the
  top band and demote the bottom band of each league tier according to a
  documented, deterministic rule, and MUST reset weekly scores to zero for
  the new period without affecting all-time streak counts.
- **FR-007**: System MUST award medals for defined streak milestones and
  ranking milestones (promotion, top finishes) and MUST make earned medals
  visible on a user's profile.
- **FR-008**: System MUST route every essay submission from the core
  application to the AI correction capability and back through a reliable,
  asynchronous hand-off (submission is durably recorded before grading
  starts, and a grading result is durably recorded and reconciled back to
  the originating submission) such that no submission is silently lost if
  the correction step is temporarily slow or unavailable.
- **FR-009**: System MUST run an automated pipeline that, on every proposed
  change to the main line, executes the full unit and integration test
  suites for each affected codebase and fails the pipeline on any test
  failure.
- **FR-010**: System MUST measure test coverage on every proposed change and
  fail the pipeline if coverage for any affected codebase falls below 90% or
  drops relative to the main line.
- **FR-011**: System MUST run a fast subset of checks (unit tests, lint,
  type-check) automatically before a change is committed locally, blocking
  the commit on failure.
- **FR-012**: System MUST prevent any change from reaching the main line
  without both (a) all required automated checks passing and (b) a recorded
  human approval, regardless of who or what authored the change.
- **FR-013**: System MUST NOT allow direct pushes or force-pushes to the main
  line — all changes arrive exclusively via reviewable pull requests.

### Key Entities

- **User**: A student account; holds profile info, subscription tier
  (free/Pro), current streak count, all-time longest streak, and league tier.
- **Essay Submission**: One essay + prompt theme submitted by a user on a
  given UTC day; tracks submission time, quota-day, and grading status
  (pending/graded/failed).
- **Grade**: The structured result of correcting one submission; holds an
  overall score and, per ENEM competency (5 total), a sub-score,
  justification text, and excerpt(s) from the essay.
- **Friendship**: A bidirectional, accepted connection between two users
  that gates visibility of streak and latest grade.
- **Weekly Ranking Entry**: A user's aggregated score for a given week and
  their resulting position/tier within the global leaderboard.
- **Medal**: An awarded achievement tied to a streak or ranking milestone,
  associated with a user and an earned date.
- **Pipeline Run**: A CI/CD execution for a proposed change; tracks which
  test suites ran, pass/fail status, and measured coverage per codebase.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A submitted essay always receives either a complete structured
  grade or an explicit failure notice — no submission is left in an
  indefinite "pending" state beyond a defined grading timeout.
- **SC-002**: 100% of free-tier users are blocked from a second same-day
  submission, and 100% of Pro-tier users are able to submit more than once
  the same day, verified across automated tests.
- **SC-003**: A user can only ever see streak/grade details for accounts
  they are actively friends with — zero cross-account data leaks in testing.
- **SC-004**: The weekly leaderboard and tier promotion/demotion produce the
  same result when re-run against the same input data (fully deterministic).
- **SC-005**: No change reaches the main line without a fully green pipeline
  and a recorded human approval — verified by attempting and observing a
  blocked merge for both a failing-check case and a no-approval case.
- **SC-006**: Combined unit test coverage across the backend, correction
  service, and mobile app is at or above 90% on the main line at all times
  after this refactor lands.
- **SC-007**: A local commit with a failing fast check (test/lint/type-check)
  is rejected before it enters version control, in 100% of tested cases.

## Assumptions

- "Fully covers the constitution" is scoped to the product/gamification
  principles ratified in this pivot (essay-driven core loop, streaks,
  friends, weekly ranking with league tiers, medals) plus the engineering
  workflow and coverage principles — not a re-litigation of pre-existing
  auth/onboarding behavior, which is assumed to already comply.
- Streak day boundary is fixed UTC (per prior constitution clarification),
  not per-user local time.
- A submission counts toward the streak immediately on acceptance, not on
  completed grading (per prior constitution clarification).
- The weekly ranking is global (all users in a tier), not friends-only, with
  Duolingo-style league tiers that promote/demote at week boundaries (per
  prior constitution clarification).
- Tie-break rule at a promotion/demotion boundary defaults to "earlier
  submission timestamp wins the higher slot" unless a product owner
  specifies otherwise before planning.
- The correction service and backend communicate asynchronously via a
  durable hand-off (e.g., a persisted job/queue record), consistent with the
  constitution's "async by default" correction principle — the exact
  transport is an implementation decision for the planning phase.
- The correction service keeps its own database/schema, separate from the
  Go monolith's, per the constitution's architecture section and prior
  clarification.
- "Fully working CI/CD" means: pre-commit fast checks + a full pipeline on
  every pull request covering all three codebases, gating merge on tests,
  coverage, and human review — the specific CI platform/tooling is an
  implementation decision for the planning phase.
- Medal criteria (which milestones, how many medal tiers) beyond "streak
  milestones" and "ranking milestones" are left to planning-phase design;
  this spec only requires that such milestones exist and are visible.
- Grading timeout duration referenced in SC-001 is an implementation
  decision for planning, not fixed here.
