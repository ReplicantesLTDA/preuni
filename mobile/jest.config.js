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
  // 0% before -- 128 lines) and (tabs)/perfil/edit.tsx's save/cancel flow.
  // Floor set to 79 for headroom. app/_layout.tsx and the per-tab
  // _layout.tsx Stack wrappers intentionally left untested -- native-module
  // bootstrap wiring / trivial one-liners with low unit-test ROI.
  coverageThreshold: {
    global: {
      lines: 79,
      statements: 79,
      functions: 75,
      branches: 66,
    },
  },
};
