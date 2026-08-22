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
  // (Principle VI: small PRs). Raised significantly this pass: added
  // contract tests for essay/streak/social/gamification api.ts + hooks.ts,
  // plus screen-level tests for redacao/{index,write,[id]}.tsx,
  // perfil/friends.tsx, and perfil/ranking.tsx (using the previously-unused
  // tests/lib/queryWrapper.tsx harness). Also fixed collectCoverageFrom,
  // which excluded every file literally named index.tsx -- that's Expo
  // Router's convention for a tab's home screen, so it was hiding
  // redacao/index.tsx, trilha/index.tsx, etc. from measurement entirely.
  // Real number after both fixes: 39.6%. Floor set to 35 for headroom.
  coverageThreshold: {
    global: {
      lines: 35,
      statements: 35,
      functions: 35,
      branches: 35,
    },
  },
};
