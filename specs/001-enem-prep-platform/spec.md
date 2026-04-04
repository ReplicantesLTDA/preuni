# Feature Specification: preuni.com.br – ENEM Prep Platform

**Feature Branch**: `001-enem-prep-platform`
**Created**: 2026-04-03
**Status**: Draft
**Input**: User description: "I want to build a Duolingo like a mobile-first app to help students to achieve Higher Education in Brazil via ENEM. This app will have the principles of Duolingo in what it stands about how to learn a thing, using spaced-repetition methods to help the student to learn some concept or thing. In this way students can complement you knowledge with non-grade-punishments methods. This app will have separated tracks per subject and ENEM simulates in it (with Dissertations)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 – Subject Track Progression (Priority: P1)

A student opens the app, picks a subject track (e.g., Mathematics), and works through a sequence of bite-sized lessons organized around ENEM concepts. Each lesson covers one concept through short exercises with immediate, non-punishing feedback. Completing a lesson marks the concept as "learned" and schedules it for future spaced-repetition review.

**Why this priority**: This is the core value proposition — a student who can pick a subject and make measurable progress already derives benefit even without any other feature present.

**Independent Test**: Can be fully tested by creating an account, selecting the Mathematics track, completing three lessons end-to-end, and verifying those concepts appear in the review queue.

**Acceptance Scenarios**:

1. **Given** a newly registered student, **When** they open the app for the first time, **Then** they are prompted to enroll in one or more subject tracks from the five ENEM subject areas.
2. **Given** a student who has chosen a track, **When** they start a lesson, **Then** the lesson presents exercises covering exactly one concept in a sequence of at most ten interactions.
3. **Given** a student who answers incorrectly during a lesson, **When** the feedback is shown, **Then** they receive a clear explanation and can retry without losing progress or receiving any score penalty.
4. **Given** a student who completes all exercises in a lesson, **When** the lesson ends, **Then** the concept is marked as "learned" and automatically added to the spaced-repetition review queue.
5. **Given** a student who exits mid-lesson, **When** they return to the app, **Then** the lesson resumes from the point where they left off.

---

### User Story 2 – Spaced-Repetition Daily Review (Priority: P2)

A student returns to the app on any given day and is presented with a personalized review session. The session surfaces concepts due for review based on their individual performance history, prioritizing those at risk of being forgotten. Completing the review reinforces those concepts and pushes the next scheduled review further into the future.

**Why this priority**: Spaced repetition is the mechanism that makes learning stick long-term. Without it the app is a simple quiz rather than a system that builds durable knowledge.

**Independent Test**: Can be tested by completing a lesson, waiting the minimum interval for that concept to become due, returning, and verifying the concept appears in the review session with higher priority than recently reviewed concepts.

**Acceptance Scenarios**:

1. **Given** a student with at least one learned concept that is due for review, **When** they open the app, **Then** a review session is surfaced prominently on the home screen.
2. **Given** a student who answers a concept correctly in a review session, **When** the answer is submitted, **Then** the concept's next review is scheduled further in the future (interval increases).
3. **Given** a student who answers a concept incorrectly in a review session, **When** the answer is submitted, **Then** the concept's review interval resets to a shorter period without affecting their streak, XP, or any visible punishment.
4. **Given** a student who completes all due reviews in a session, **When** the last concept is reviewed, **Then** a summary screen shows how many concepts were strengthened.
5. **Given** a student who has missed several days, **When** they return, **Then** overdue reviews are batched into manageable sessions of no more than 20 concepts rather than shown all at once.

---

### User Story 3 – ENEM Full Simulation (Priority: P3)

A student tests their overall readiness by taking a full ENEM simulation. The simulation mirrors the real exam: 180 multiple-choice questions across all five subject areas plus one dissertation prompt. The student can complete it in one sitting or pause and resume. After submission they receive a performance report broken down by subject area and ENEM competency.

**Why this priority**: Simulations connect daily learning to the real exam goal and provide the concrete readiness signal that motivates sustained study effort.

**Independent Test**: Can be tested independently by starting a simulation, answering all 180 multiple-choice questions, submitting a dissertation draft, and verifying the result report shows scores per subject area.

**Acceptance Scenarios**:

1. **Given** a registered student, **When** they choose to start a new ENEM simulation, **Then** they are presented with 180 multiple-choice questions in the standard ENEM subject-area order.
2. **Given** a student mid-simulation, **When** they close the app or navigate away, **Then** all answers entered so far are saved and the simulation is marked as "in progress."
3. **Given** a student returning to an in-progress simulation, **When** they open it, **Then** they can continue from the last unanswered question or navigate freely among questions already answered.
4. **Given** a student who has answered all multiple-choice questions, **When** they proceed to the dissertation section, **Then** they are presented with a contextualized ENEM-style writing prompt with supporting texts.
5. **Given** a student who submits the complete simulation, **When** results are processed, **Then** a report shows their score per subject area and highlights which ENEM competencies need the most improvement.
6. **Given** a student who has completed multiple simulations, **When** they view their simulation history, **Then** scores are displayed over time so improvement can be tracked.

---

### User Story 4 – Standalone Dissertation Practice (Priority: P4)

A student wants to improve their dissertation writing independently of a full simulation. They browse a library of ENEM-style prompts, choose one, write their response in the app, and receive structured feedback aligned to the five official ENEM dissertation competency criteria (Competências 1–5).

**Why this priority**: The dissertation accounts for up to 1,000 of the 1,000-point ENEM scale and is the most neglected skill in available study tools. Standalone practice addresses a high-stakes, underserved need.

**Independent Test**: Can be tested independently by selecting a dissertation prompt from the library, writing and submitting a response, and verifying that feedback covering all five ENEM competency criteria is returned.

**Acceptance Scenarios**:

1. **Given** a student accessing the dissertation practice section, **When** they browse available prompts, **Then** they see a library organized by thematic area and recency, with at least ten available prompts.
2. **Given** a student who selects a prompt, **When** the writing view opens, **Then** they see the prompt text, any supporting materials, and a distraction-free text editor.
3. **Given** a student who submits a completed dissertation, **When** feedback is generated, **Then** it is structured around the five official ENEM competency criteria with a qualitative rating and improvement notes for each.
4. **Given** a student viewing their feedback, **When** they read the competency breakdown, **Then** each criterion includes specific, actionable improvement suggestions rather than generic comments.
5. **Given** a student who revises and resubmits a dissertation, **When** new feedback is displayed, **Then** it is shown alongside the previous submission and feedback so the student can compare their improvement.

---

### User Story 5 – Progress Dashboard & Motivational Gamification (Priority: P5)

A student checks their progress and motivation signals. They access a dashboard showing their daily streak, total XP, per-subject track completion, simulation score history, and a composite ENEM readiness indicator. Reaching milestones unlocks achievement badges.

**Why this priority**: Sustained daily engagement is what produces learning outcomes. Gamification is the primary lever for that engagement, but it depends on the core learning loops (P1–P4) existing first.

**Independent Test**: Can be tested by completing a week of daily activity and verifying the dashboard displays streaks, XP, per-subject progress bars, and at least one earned achievement.

**Acceptance Scenarios**:

1. **Given** a student who has completed at least one activity, **When** they open the dashboard, **Then** they see their streak count, total XP, and a progress bar for each enrolled subject track.
2. **Given** a student who completes a lesson or review session, **When** the activity ends, **Then** XP is awarded with a brief, non-intrusive animated celebration.
3. **Given** a student who has not completed any activity today and whose streak is still active, **When** they open the app, **Then** a friendly reminder (not a punishment) encourages them to keep their streak alive.
4. **Given** a student who reaches a milestone (e.g., 7-day streak, first simulation completed, first track finished), **When** the milestone is triggered, **Then** an achievement badge is unlocked and displayed on their profile.
5. **Given** a student who has progress across tracks and simulations, **When** they view the readiness indicator, **Then** a single composite metric estimates their current ENEM readiness based on track mastery and simulation performance history.

---

### Edge Cases

- What happens when a student has no due reviews and no remaining new lessons in any enrolled track?
- How does the system handle a dissertation draft that was started but never submitted — is it auto-saved, and for how long is it retained?
- What happens when a student exits a simulation after completing only some subject areas — can partial results be shown for completed sections?
- How does spaced repetition handle a concept the student has never answered correctly (risk of perpetual short-interval cycling)?
- What happens if a student unenrolls from a subject track mid-way — is their progress preserved if they re-enroll later?
- How does the app behave when the student's device is offline?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST organize learning content into five subject tracks corresponding to the official ENEM subject areas: Languages & Codes, Human Sciences, Natural Sciences, Mathematics, and Writing/Dissertation.
- **FR-002**: Each subject track MUST contain lessons, where each lesson covers exactly one concept through a sequence of short exercises.
- **FR-003**: The system MUST implement a spaced-repetition scheduling algorithm that determines when each learned concept is due for review based on each student's individual performance history.
- **FR-004**: The system MUST provide immediate feedback after every exercise answer without removing progress, ending the session, or applying any numeric penalty for incorrect answers.
- **FR-005**: The system MUST offer a full ENEM simulation mode containing 180 multiple-choice questions and one dissertation prompt, structured in the standard ENEM subject-area order.
- **FR-006**: Students MUST be able to pause a simulation at any point and resume it later with all previously entered answers preserved.
- **FR-007**: The system MUST provide a standalone dissertation practice section with a library of ENEM-style writing prompts, each including supporting texts.
- **FR-008**: Dissertation submissions MUST receive structured feedback mapped explicitly to the five official ENEM dissertation competency criteria (Competências 1–5).
- **FR-009**: The system MUST track and display per-student progress for each enrolled subject track, including concepts learned, concepts mastered, and concepts due for review.
- **FR-010**: The system MUST track a daily activity streak and notify students in a non-punishing manner when their streak is at risk of being lost.
- **FR-011**: The system MUST award experience points (XP) for completed lessons, review sessions, and simulations, and display cumulative XP on the student's profile.
- **FR-012**: The system MUST display a composite ENEM readiness indicator that combines subject track mastery levels and simulation score history.
- **FR-013**: The system MUST be fully functional on mobile devices as the primary form factor, with all interactions optimized for touch and small screens.
- **FR-014**: Students MUST be able to create an account and log in, with all progress and history persisted and accessible across multiple devices.

### Key Entities

- **Student**: A registered user with an enrolled subject track list, spaced-repetition state per concept, streak count, XP total, and profile.
- **Subject Track**: One of the five ENEM subject areas, containing an ordered sequence of Lessons progressing from foundational to advanced topics.
- **Lesson**: A single learning unit covering one Concept, composed of a sequence of Exercises, with a completion status per Student.
- **Exercise**: An individual interaction (multiple-choice, fill-in-the-blank, matching, etc.) with a correct answer, an explanation, and a difficulty level.
- **Concept**: The atomic unit tracked by spaced repetition; each Concept has a per-Student review schedule (interval, ease factor, next due date).
- **Review Session**: A set of Concepts due for reinforcement surfaced to a Student, generated by the spaced-repetition algorithm.
- **Simulation**: A full mock ENEM exam instance linked to a Student, containing saved answers per question, a start timestamp, a completion status, and a result report.
- **Dissertation Prompt**: A writing theme with a contextualized scenario and supporting texts, aligned to ENEM-style themes.
- **Dissertation Submission**: A Student's written response to a Dissertation Prompt, with a submission timestamp and associated Feedback.
- **Feedback**: A structured assessment of a Dissertation Submission, providing per-criterion scores and actionable improvement suggestions aligned to the five ENEM competency criteria.
- **Achievement**: A milestone badge with a defined unlock condition (e.g., "7-day streak", "completed first simulation"), tracked per Student.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new student can enroll in a subject track and begin their first lesson within 60 seconds of completing account registration.
- **SC-002**: A daily review session covering 20 concepts can be completed in under 10 minutes.
- **SC-003**: Students who complete at least 30 days of consistent app use show a measurable improvement in their second simulation score compared to their first.
- **SC-004**: At least 70% of students who start a lesson complete it within the same session.
- **SC-005**: A full simulation (180 questions + dissertation) can be paused and resumed across multiple sessions without any loss of previously entered answers.
- **SC-006**: Dissertation feedback covering all five competency criteria is delivered within 60 seconds of submission.
- **SC-007**: Among students who have used the app for more than 7 days, at least 60% return for activity on any given day (daily active rate target for the engaged cohort).
- **SC-008**: The full student journey — login → daily review → new lesson → dashboard — is completable on a mobile device without horizontal scrolling, layout overflow, or inaccessible touch targets.

## Assumptions

- Students are the primary users; teacher, tutor, and administrator roles are out of scope for v1.
- Dissertation feedback in v1 is AI-generated and not human-reviewed; a human-review tier may be introduced in a later version.
- The five ENEM subject areas and the five dissertation competency criteria follow the official INEP specification for the current exam format.
- The spaced-repetition algorithm in v1 is based on an established, proven algorithm (e.g., SM-2 or equivalent); a custom proprietary algorithm is deferred to a later version.
- Learning content (questions, exercises, prompts, explanations) is authored or curated internally by the product team; a content management or authoring interface for third parties is out of scope for v1.
- The app is mobile-first (iOS and Android); a dedicated desktop web experience is a future scope item and not required for v1 launch.
- Students are expected to have a smartphone and a data connection for the core learning experience; offline support is a future enhancement, not a v1 requirement.
- User authentication uses standard secure practices (email/password with verified email); social login and institutional SSO are out of scope for v1.
