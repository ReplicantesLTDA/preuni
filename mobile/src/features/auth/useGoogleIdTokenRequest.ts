import * as Google from 'expo-auth-session/providers/google';
import * as WebBrowser from 'expo-web-browser';

// Required once per app so the browser-based auth flow can hand control
// back to the app after Google redirects -- see expo-auth-session docs.
WebBrowser.maybeCompleteAuthSession();

// expo-auth-session throws at render time if its resolved client ID is
// undefined -- this placeholder keeps the hook (and Rules of Hooks) happy
// when issue #39's real Google Cloud client IDs haven't been configured
// yet. `configured` tells callers to hide the "Continue with Google"
// button entirely in that case, so the placeholder is never actually used.
const UNCONFIGURED_PLACEHOLDER = 'google-oauth-not-configured';

/**
 * Wraps expo-auth-session's Google ID-token flow with this app's env vars.
 * `configured` is false until at least one of
 * EXPO_PUBLIC_GOOGLE_{IOS,ANDROID,WEB}_CLIENT_ID is set -- callers should
 * not render the Google sign-in button while it's false.
 */
export function useGoogleIdTokenRequest() {
  const iosClientId = process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID;
  const androidClientId = process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID;
  const webClientId = process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID;
  const configured = Boolean(iosClientId || androidClientId || webClientId);

  const [request, response, promptAsync] = Google.useIdTokenAuthRequest({
    iosClientId: iosClientId || UNCONFIGURED_PLACEHOLDER,
    androidClientId: androidClientId || UNCONFIGURED_PLACEHOLDER,
    webClientId: webClientId || UNCONFIGURED_PLACEHOLDER,
  });

  return { request, response, promptAsync, configured };
}
