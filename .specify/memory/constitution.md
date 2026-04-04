# preuni.com.br Constitution

## Core Principles

### I. Code Quality (NON-NEGOTIABLE)

Every piece of code merged to `main` must meet these standards:

- **Readability first**: Code is written once but read many times; clarity beats cleverness
- **Single responsibility**: Functions and modules do one thing well; no god objects or catch-all utilities
- **No dead code**: Unused imports, variables, functions, and feature flags must be removed before merge
- **Explicit over implicit**: Avoid magic values, undocumented side effects, and surprise mutations
- **Consistent naming**: Follow the conventions already present in the file/module; do not mix styles
- **No premature abstraction**: Three similar lines of code are better than a wrong abstraction; extract only when a third use case appears

### II. Testing Standards (NON-NEGOTIABLE)

- **Test-first for new features**: Write failing tests → get approval → implement → pass tests (Red-Green-Refactor)
- **Unit tests required** for all business logic, utilities, and pure functions
- **Integration tests required** for: API endpoints, database interactions, authentication flows, and third-party service boundaries
- **No test skipping**: `skip`, `xit`, `xtest`, or equivalent are forbidden in CI — comment them with a tracked issue instead
- **Test names must describe behavior**: `it("returns 404 when user does not exist")` not `it("works")`
- **Coverage floor**: Maintain ≥ 80% line coverage; new code must not lower the project average
- **Real dependencies over mocks** at integration boundaries: mock only what you own or what is external and unreliable

### III. User Experience Consistency

- **Design system adherence**: All UI components must use tokens from the design system (colors, spacing, typography); raw hex/px values are not allowed in component styles
- **Consistent feedback patterns**: Loading states, error messages, and empty states must follow the established patterns for the application; no one-off spinners or ad-hoc error strings
- **Accessible by default**: Every interactive element requires keyboard navigability and ARIA labels; WCAG AA is the minimum bar
- **Mobile-first**: Layouts are designed and reviewed on mobile before desktop; no feature ships without a responsive implementation
- **Copy consistency**: User-facing strings follow the established tone (friendly, direct, encouraging for a pre-university audience); no technical jargon exposed to end users
- **Predictable navigation**: URLs are stable and bookmarkable; back-button behavior must work as expected

### IV. Performance Requirements

- **Page load (LCP)**: ≤ 2.5 s on a simulated mid-tier mobile device (Lighthouse, throttled 4G)
- **Interaction responsiveness (INP)**: ≤ 200 ms for all user interactions
- **Bundle size budget**: No single route chunk > 150 kB (gzipped); new dependencies must be evaluated for size impact before adoption
- **Image optimization**: All images served in next-gen formats (WebP/AVIF) with explicit `width`/`height` to prevent layout shift (CLS ≤ 0.1)
- **API response time**: p95 < 500 ms for all user-facing endpoints; queries hitting the database must have an EXPLAIN plan reviewed before merge
- **No N+1 queries**: Every new data-fetching path must be reviewed for N+1 patterns; use batching or eager loading where appropriate

## Quality Gates

Every pull request must pass all of the following before merge:

- [ ] CI is green (lint, type-check, unit tests, integration tests)
- [ ] Coverage has not decreased
- [ ] Lighthouse CI score does not regress on LCP, INP, or CLS
- [ ] Bundle size budget is not exceeded
- [ ] Design system tokens used (no raw values in styles)
- [ ] Accessibility: no new axe-core violations introduced
- [ ] Reviewer has verified mobile layout

## Governance

- This Constitution supersedes all other practices and informal agreements
- Amendments require: written proposal, team discussion, documented rationale, and update to this file
- All code reviews must verify compliance with these principles — "it works" is not sufficient approval
- Exceptions must be documented inline with a comment referencing a tracked issue and an expiry plan
- Complexity must be justified; the burden of proof is on the author adding it

**Version**: 1.0.0 | **Ratified**: 2026-04-03 | **Last Amended**: 2026-04-03
