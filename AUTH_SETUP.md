# Authentication Setup Guide

This guide explains the authentication features that have been implemented and how to configure them.

## Features Implemented

### 1. Password Validation
- Minimum 8 characters required
- Must contain at least one letter (a-z or A-Z)
- Must contain at least one number (0-9)
- Real-time visual feedback showing which requirements are met
- Validation occurs both on frontend (for UX) and can be added to backend

**Location:**
- Frontend validation: `app/src/lib/validation.ts`
- UI implementation: `app/src/pages/Auth.tsx`

### 2. OAuth Authentication
Support for three OAuth providers:
- **Google OAuth 2.0**
- **GitHub OAuth**
- **Apple Sign In**

**Locations:**
- Frontend OAuth utilities: `app/src/lib/oauth.ts`
- Backend OAuth handlers: `server/internal/handlers/oauth.go`
- Auth context: `app/src/context/authContext.tsx`

## Setup Instructions

### Frontend Setup

1. **Copy environment variables:**
   ```bash
   cd app
   cp .env.example .env
   ```

2. **Configure OAuth credentials in `app/.env`:**
   ```env
   # Google OAuth
   VITE_GOOGLE_CLIENT_ID=your_google_client_id_here

   # GitHub OAuth
   VITE_GITHUB_CLIENT_ID=your_github_client_id_here
   VITE_GITHUB_REDIRECT_URI=http://localhost:5173/auth/callback/github

   # Apple Sign In
   VITE_APPLE_CLIENT_ID=your_apple_client_id_here
   VITE_APPLE_REDIRECT_URI=http://localhost:5173/auth/callback/apple
   ```

### Backend Setup

1. **Copy environment variables:**
   ```bash
   cd server
   cp .env.example .env
   ```

2. **Configure OAuth credentials in `server/.env`:**
   ```env
   # JWT Secret (required)
   JWT_SECRET=your_strong_random_string_min_32_chars

   # Google OAuth
   GOOGLE_CLIENT_ID=your_google_client_id_here
   GOOGLE_CLIENT_SECRET=your_google_client_secret_here
   GOOGLE_REDIRECT_URI=http://localhost:8081/v1/auth/google/callback

   # GitHub OAuth
   GITHUB_CLIENT_ID=your_github_client_id_here
   GITHUB_CLIENT_SECRET=your_github_client_secret_here

   # Apple Sign In
   APPLE_CLIENT_ID=your_apple_client_id_here
   APPLE_CLIENT_SECRET=your_apple_client_secret_here
   APPLE_REDIRECT_URI=http://localhost:8081/v1/auth/apple/callback
   ```

## Getting OAuth Credentials

### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project or select an existing one
3. Enable "Google+ API"
4. Create OAuth 2.0 credentials
5. Add authorized JavaScript origins: `http://localhost:5173`
6. Add authorized redirect URIs: `http://localhost:8081/v1/auth/google/callback`
7. Copy the Client ID and Client Secret

### GitHub OAuth

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in:
   - Application name: Your app name
   - Homepage URL: `http://localhost:5173`
   - Authorization callback URL: `http://localhost:5173/auth/callback/github`
4. Click "Register application"
5. Copy the Client ID and generate a Client Secret

### Apple Sign In

1. Go to [Apple Developer Account](https://developer.apple.com/account/resources/identifiers/list/serviceId)
2. Create a new Services ID
3. Enable "Sign in with Apple"
4. Configure:
   - Domains: `localhost`
   - Return URLs: `http://localhost:5173/auth/callback/apple`
5. Create a private key for Sign in with Apple
6. Download and configure the credentials

**Note:** Apple Sign In requires additional JWT configuration for the client secret. The current implementation has basic support but requires proper JWT token generation for production use.

## API Endpoints

The following OAuth endpoints are available:

- `POST /v1/auth/google` - Google OAuth callback
- `POST /v1/auth/github` - GitHub OAuth callback
- `POST /v1/auth/apple` - Apple Sign In callback

## Testing

1. Start the backend server:
   ```bash
   cd server
   go run cmd/server/main.go
   ```

2. Start the frontend dev server:
   ```bash
   cd app
   npm run dev
   ```

3. Navigate to `http://localhost:5173/auth`
4. Try registering with a password that doesn't meet requirements - you'll see real-time validation
5. Try logging in with OAuth providers (requires proper credentials configured)

## Security Notes

- Password validation is implemented in the frontend for user experience
- Consider adding backend password validation as well for additional security
- OAuth tokens are exchanged securely through the backend
- JWT tokens are used for session management
- Refresh tokens are stored securely with HTTP-only cookies
- Always use HTTPS in production
- Keep your OAuth secrets secure and never commit them to version control

## Production Checklist

Before deploying to production:

- [ ] Change JWT_SECRET to a strong random string (32+ characters)
- [ ] Update all redirect URIs to production URLs
- [ ] Enable HTTPS for all OAuth callbacks
- [ ] Implement proper error logging
- [ ] Add rate limiting to auth endpoints
- [ ] Implement CSRF protection
- [ ] Complete Apple Sign In JWT implementation
- [ ] Review and test all OAuth flows
- [ ] Set up proper environment variable management
