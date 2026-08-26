// Silence Reanimated logger warnings in tests
jest.mock('react-native-reanimated', () => require('react-native-reanimated/mock'));

// expo-auth-session's Google provider calls expo-linking's createURL at
// render time, which needs a real app.config.ts scheme unavailable under
// jest's expo-constants mock -- stub it out so screens using
// useGoogleIdTokenRequest can still render in tests.
jest.mock('expo-auth-session/providers/google', () => ({
  useIdTokenAuthRequest: () => [null, null, jest.fn()],
}));
jest.mock('expo-web-browser', () => ({
  maybeCompleteAuthSession: jest.fn(),
}));
