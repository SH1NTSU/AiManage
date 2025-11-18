// Auth.js-style authentication wrapper
// Compatible with existing Go backend

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8081";

// Helper to check if value is a placeholder
function isPlaceholder(value: string): boolean {
  const placeholders = [
    'your_github_client_id_here',
    'your_google_client_id_here',
    'your_apple_client_id_here',
    'your_client_id_here',
    'dummy-client-id',
  ];
  return placeholders.some(placeholder => 
    value.toLowerCase().includes(placeholder.toLowerCase())
  );
}

// OAuth Provider Configuration (Auth.js style)
export const providers = {
  google: {
    id: "google",
    name: "Google",
    type: "oauth" as const,
    authorization: {
      url: "https://accounts.google.com/o/oauth2/v2/auth",
      params: {
        scope: "openid email profile",
        response_type: "code",
        access_type: "offline",
        prompt: "consent",
      },
    },
    clientId: (() => {
      const id = import.meta.env.VITE_GOOGLE_CLIENT_ID || "";
      if (id && isPlaceholder(id)) {
        console.warn("⚠️ Google Client ID appears to be a placeholder. Please update VITE_GOOGLE_CLIENT_ID in your .env file and restart the dev server.");
      }
      return id;
    })(),
    redirectUri: `${window.location.origin}/auth/callback/google`,
  },
  github: {
    id: "github",
    name: "GitHub",
    type: "oauth" as const,
    authorization: {
      url: "https://github.com/login/oauth/authorize",
      params: {
        scope: "read:user user:email",
      },
    },
    clientId: (() => {
      const id = import.meta.env.VITE_GITHUB_CLIENT_ID || "";
      if (id && isPlaceholder(id)) {
        console.warn("⚠️ GitHub Client ID appears to be a placeholder. Please update VITE_GITHUB_CLIENT_ID in your .env file and restart the dev server.");
      }
      return id;
    })(),
    redirectUri: import.meta.env.VITE_GITHUB_REDIRECT_URI || `${window.location.origin}/auth/callback/github`,
  },
} as const;

// Generate random state for CSRF protection
function generateState(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, "0")).join("");
}

// Auth.js-style signIn function
export async function signIn(provider: "google" | "github") {
  const providerConfig = providers[provider];
  
  // Debug logging
  console.log(`🔐 Attempting ${provider} OAuth sign in`);
  console.log(`📋 Client ID: ${providerConfig.clientId ? providerConfig.clientId.substring(0, 10) + '...' : 'NOT SET'}`);
  console.log(`🔗 Redirect URI: ${providerConfig.redirectUri}`);
  
  if (!providerConfig.clientId) {
    const errorMsg = `${provider} OAuth is not configured. Please check VITE_${provider.toUpperCase()}_CLIENT_ID in your .env file.`;
    console.error(`❌ ${errorMsg}`);
    throw new Error(errorMsg);
  }

  const state = generateState();
  sessionStorage.setItem(`oauth_state_${provider}`, state);
  console.log(`✅ State generated and stored: ${state.substring(0, 10)}...`);

  const params = new URLSearchParams({
    client_id: providerConfig.clientId,
    redirect_uri: providerConfig.redirectUri,
    response_type: "code",
    state,
    ...providerConfig.authorization.params,
  });

  const authUrl = `${providerConfig.authorization.url}?${params.toString()}`;
  console.log(`🚀 Redirecting to: ${providerConfig.authorization.url}`);
  window.location.href = authUrl;
}

// Handle OAuth callback (Auth.js style)
export async function handleCallback(
  provider: "google" | "github",
  code: string,
  state: string
): Promise<{ token: string; refresh_token: string }> {
  console.log(`🔐 Verifying OAuth callback for ${provider}...`);
  
  // Verify state
  const storedState = sessionStorage.getItem(`oauth_state_${provider}`);
  console.log(`📋 Stored state: ${storedState ? storedState.substring(0, 10) + '...' : 'NOT FOUND'}`);
  console.log(`📋 Received state: ${state.substring(0, 10)}...`);

  if (!storedState || storedState !== state) {
    console.error(`❌ State mismatch! Stored: ${storedState}, Received: ${state}`);
    throw new Error("Invalid state parameter - possible CSRF attack or session expired");
  }
  
  // Remove state after verification
  sessionStorage.removeItem(`oauth_state_${provider}`);
  console.log(`✅ State verified successfully`);

  // Exchange code for token via backend (Auth.js style)
  const providerConfig = providers[provider];
  console.log(`🔄 Exchanging ${provider} authorization code for token...`);
  console.log(`📡 Backend URL: ${API_URL}/v1/auth/${provider}`);
  console.log(`🔗 Redirect URI: ${providerConfig.redirectUri}`);
  
  const response = await fetch(`${API_URL}/v1/auth/${provider}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ 
      code,
      redirect_uri: providerConfig.redirectUri, // Send redirect URI for backend validation
    }),
  });

  if (!response.ok) {
    const error = await response.text();
    console.error(`❌ OAuth callback failed: ${response.status} ${response.statusText}`);
    console.error(`❌ Error details: ${error}`);
    throw new Error(`OAuth callback failed: ${error}`);
  }
  
  console.log(`✅ Token exchange successful`);

  const data = await response.json();
  return {
    token: data.token,
    refresh_token: data.refresh_token || "",
  };
}

// Auth.js-style signOut
export function signOut() {
  localStorage.removeItem("token");
  sessionStorage.clear();
  window.location.href = "/auth";
}

