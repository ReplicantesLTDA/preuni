module.exports = {
  preset: 'jest-expo',
  setupFiles: ['<rootDir>/tests/patch-rn-mocks.js'],
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],
  testMatch: ['**/tests/**/*.test.{ts,tsx}'],
  transformIgnorePatterns: [
    'node_modules/(?!.*(react-native|@react-native|expo|@expo|expo-router|expo-modules-core|@react-navigation|@testing-library|react-native-svg|react-native-reanimated|react-native-gesture-handler|react-native-safe-area-context|react-native-screens|@expo/vector-icons|nanoid))',
  ],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^msw/node$': '<rootDir>/node_modules/msw/lib/node/index.js',
    '^msw$': '<rootDir>/node_modules/msw/lib/core/index.js',
  },
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    'app/**/*.{ts,tsx}',
    '!**/*.d.ts',
    // Barrel re-export files only (e.g. src/features/auth/index.ts) -- NOT
    // app/**/index.tsx, which is Expo Router's convention for every tab's
    // home screen (redacao/index.tsx, trilha/index.tsx, ...). The old
    // blanket '!**/index.{ts,tsx}' excluded those route screens from
    // coverage entirely -- a real measurement bug found while adding
    // screen tests for redacao/index.tsx (2026-08-22).
    '!src/**/index.{ts,tsx}',
  ],
  // Constitution Principle II: coverage floor is 90%, reached incrementally
  // (Principle VI: small PRs). 77.0% -> 78.4% after adding tests for
  // app/index.tsx and app/+not-found.tsx (root-level redirects), then
  // 80.9% after covering (tabs)/perfil/index.tsx (the profile hub screen,
  // 0% before -- 128 lines) and (tabs)/perfil/edit.tsx's save/cancel flow,
  // then 81.1% after trilha CTA/subject-tile interactions, and now 82.9%
  // after a batched round: features/perfil/hooks.ts's useUploadAvatar
  // (dev-stub path)/useDeleteMe/useChangeEmailConfirm (0% before),
  // errorMessages.ts's fallback + authErrorByCode (0% before, no test
  // file existed), and change-email.tsx's request-error/confirm-success/
  // confirm-error branches, and now 86.2% after a large batch: lib/auth/
  // tokenStore.ts (native SecureStore + web localStorage branches),
  // stores/sessionStore.ts, lib/query/{keys,client}.ts, lib/api/context.tsx,
  // features/auth/hooks.ts (persistSession's embedded-student branch,
  // useStudentMe, useVerifyEmail), default-prop branches across several
  // shared components (Card, ScreenContainer, Skeleton, ErrorState, Toast,
  // Button, TopStatusBar, NextActivityCard, SubjectGrid), and more
  // onError-toast branches in perfil/social screens, and now 87.3% after
  // redacao/index.tsx (pending/failed status labels, essay-item nav,
  // sample-prompt nav), write.tsx (429 quota-exceeded + generic-error
  // toast branches), perfil/index.tsx (all 6 remaining menu links), and
  // useFonts.ts (mocked expo-font, both loaded/error branches), and now
  // 89.83% (constitution's 90% floor reached) after trilha/index.tsx's
  // RefreshControl.onRefresh + computeGreeting's 3 time-of-day branches,
  // and server-side field-validation-error branches (422 with a `field`)
  // across the (auth) screens, plus resend-code and weak-password-on-
  // confirm branches, and now 90.41% -- constitution's 90% floor
  // reached -- after useUploadAvatar's real-presigned-URL PUT path
  // (success + S3-failure branches; only the dev-stub skip-PUT path was
  // tested before), ApiProvider's development-mode request logger, and
  // t.trilha.streak's singular/plural ternary (no i18n test file
  // existed at all). Floor set to 90 for headroom. app/_layout.tsx and
  // the per-tab _layout.tsx Stack wrappers intentionally left untested --
  // native-module bootstrap wiring / trivial one-liners with low
  // unit-test ROI. Remaining real gaps: perfil/edit.tsx's and
  // (onboarding)/profile.tsx's avatar-picker flows (native ImagePicker,
  // out of scope all session), and Button/NextActivityCard's `pressed`-
  // state style callbacks (RTL's fireEvent.press doesn't capture the
  // active-touch frame).
  coverageThreshold: {
    global: {
      lines: 90,
      statements: 90,
      functions: 91,
      branches: 81,
    },
  },
};
