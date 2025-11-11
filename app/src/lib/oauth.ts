// OAuth Configuration
export const OAUTH_CONFIG = {
  google: {
    clientId: import.meta.env.VITE_GOOGLE_CLIENT_ID || '',
  },
  github: {
    clientId: import.meta.env.VITE_GITHUB_CLIENT_ID || '',
    redirectUri: import.meta.env.VITE_GITHUB_REDIRECT_URI || 'http://localhost:5173/auth/callback/github',
    scope: 'read:user user:email',
  },
  apple: {
    clientId: import.meta.env.VITE_APPLE_CLIENT_ID || '',
    redirectUri: import.meta.env.VITE_APPLE_REDIRECT_URI || 'http://localhost:5173/auth/callback/apple',
    scope: 'name email',
  },
};

// Generate random string for state and code verifier
function generateRandomString(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  let result = '';
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);
  for (let i = 0; i < length; i++) {
    result += chars[randomValues[i] % chars.length];
  }
  return result;
}

// Generate code challenge for PKCE
async function generateCodeChallenge(codeVerifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(codeVerifier);
  const hash = await crypto.subtle.digest('SHA-256', data);
  const base64 = btoa(String.fromCharCode(...new Uint8Array(hash)));
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

// GitHub OAuth
export async function initiateGitHubOAuth() {
  const state = generateRandomString(32);
  const codeVerifier = generateRandomString(64);
  const codeChallenge = await generateCodeChallenge(codeVerifier);

  // Store state and code verifier in session storage
  sessionStorage.setItem('github_oauth_state', state);
  sessionStorage.setItem('github_code_verifier', codeVerifier);

  const params = new URLSearchParams({
    client_id: OAUTH_CONFIG.github.clientId,
    redirect_uri: OAUTH_CONFIG.github.redirectUri,
    scope: OAUTH_CONFIG.github.scope,
    state: state,
    response_type: 'code',
  });

  window.location.href = `https://github.com/login/oauth/authorize?${params.toString()}`;
}

// Apple OAuth
export async function initiateAppleOAuth() {
  const state = generateRandomString(32);
  const nonce = generateRandomString(32);

  // Store state in session storage
  sessionStorage.setItem('apple_oauth_state', state);
  sessionStorage.setItem('apple_nonce', nonce);

  const params = new URLSearchParams({
    client_id: OAUTH_CONFIG.apple.clientId,
    redirect_uri: OAUTH_CONFIG.apple.redirectUri,
    response_type: 'code id_token',
    response_mode: 'form_post',
    scope: OAUTH_CONFIG.apple.scope,
    state: state,
    nonce: nonce,
  });

  window.location.href = `https://appleid.apple.com/auth/authorize?${params.toString()}`;
}

// Verify OAuth state to prevent CSRF attacks
export function verifyOAuthState(provider: 'github' | 'apple', receivedState: string): boolean {
  const storedState = sessionStorage.getItem(`${provider}_oauth_state`);
  sessionStorage.removeItem(`${provider}_oauth_state`);
  return storedState === receivedState;
}

// Get code verifier for token exchange
export function getCodeVerifier(provider: 'github'): string | null {
  const verifier = sessionStorage.getItem(`${provider}_code_verifier`);
  sessionStorage.removeItem(`${provider}_code_verifier`);
  return verifier;
}
