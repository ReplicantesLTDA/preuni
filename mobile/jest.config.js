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
    '!**/index.{ts,tsx}',
  ],
  // Constitution Principle II: coverage floor is 90%, reached incrementally
  // (Principle VI: small PRs). 25/30/20/25 is today's *provisional*
  // baseline (measured ~28.5% lines pre-014) — specs/014-constitution-
  // alignment-refactor/tasks.md T060 (Polish) raises this to 90 once the
  // essay/streak/social/gamification screens (US1-US3) land their tests.
  coverageThreshold: {
    global: {
      lines: 25,
      statements: 25,
      functions: 20,
      branches: 30,
    },
  },
};
