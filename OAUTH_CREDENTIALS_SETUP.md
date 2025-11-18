# OAuth Credentials Setup Guide

This guide will walk you through obtaining OAuth credentials for Google and GitHub to enable social login in your application.

## Google OAuth Setup

### Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click on the project dropdown at the top
3. Click **"New Project"**
4. Enter a project name (e.g., "AiManage")
5. Click **"Create"**

### Step 2: Enable Google+ API

1. In the Google Cloud Console, go to **"APIs & Services"** > **"Library"**
2. Search for **"Google+ API"** or **"Google Identity Services"**
3. Click on it and click **"Enable"**

### Step 3: Configure OAuth Consent Screen

1. Go to **"APIs & Services"** > **"OAuth consent screen"**
2. Choose **"External"** (unless you have a Google Workspace account)
3. Click **"Create"**
4. Fill in the required information:
   - **App name**: Your application name (e.g., "AI Model Manager")
   - **User support email**: Your email address
   - **Developer contact information**: Your email address
5. Click **"Save and Continue"**
6. On the **"Scopes"** page, click **"Add or Remove Scopes"**
   - Add: `openid`, `email`, `profile`
7. Click **"Save and Continue"**
8. On the **"Test users"** page (if in testing mode), add test email addresses
9. Click **"Save and Continue"** and then **"Back to Dashboard"**

### Step 4: Create OAuth 2.0 Credentials

1. Go to **"APIs & Services"** > **"Credentials"**
2. Click **"+ CREATE CREDENTIALS"** > **"OAuth client ID"**
3. Select **"Web application"** as the application type
4. Fill in the details:
   - **Name**: "AiManage Web Client" (or any name you prefer)
   - **Authorized JavaScript origins**:
     - `http://localhost:5173` (for development)
     - `http://localhost:8080` (if using different port)
     - Your production URL (e.g., `https://yourdomain.com`)
   - **Authorized redirect URIs**:
     - `http://localhost:5173/auth/callback/google` (for development)
     - `http://localhost:8080/auth/callback/google` (if using different port)
     - `https://yourdomain.com/auth/callback/google` (for production)
5. Click **"Create"**
6. **Copy the Client ID and Client Secret** - you'll need these!

### Step 5: Configure Your Application

Add the credentials to your `.env` files:

**Frontend (`app/.env`):**
```env
VITE_GOOGLE_CLIENT_ID=your_google_client_id_here
```

**Backend (`server/.env`):**
```env
GOOGLE_CLIENT_ID=your_google_client_id_here
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
GOOGLE_REDIRECT_URI=http://localhost:5173/auth/callback/google
```

---

## GitHub OAuth Setup

### Step 1: Create a GitHub OAuth App

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click **"OAuth Apps"** in the left sidebar
3. Click **"New OAuth App"** (or **"Register a new application"**)

### Step 2: Fill in Application Details

1. **Application name**: Your app name (e.g., "AI Model Manager")
2. **Homepage URL**: 
   - Development: `http://localhost:5173`
   - Production: `https://yourdomain.com`
3. **Authorization callback URL**:
   - Development: `http://localhost:5173/auth/callback/github`
   - Production: `https://yourdomain.com/auth/callback/github`
4. Click **"Register application"**

### Step 3: Get Your Credentials

1. After creating the app, you'll see the **Client ID** (public)
2. Click **"Generate a new client secret"** to get your **Client Secret**
3. **Copy both** - you'll need them!

**Important**: The client secret is only shown once. Make sure to copy it immediately!

### Step 4: Configure Your Application

Add the credentials to your `.env` files:

**Frontend (`app/.env`):**
```env
VITE_GITHUB_CLIENT_ID=your_github_client_id_here
VITE_GITHUB_REDIRECT_URI=http://localhost:5173/auth/callback/github
```

**Backend (`server/.env`):**
```env
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
```

---

## Environment Variables Summary

### Frontend (`app/.env`)
```env
# Google OAuth
VITE_GOOGLE_CLIENT_ID=your_google_client_id_here

# GitHub OAuth
VITE_GITHUB_CLIENT_ID=your_github_client_id_here
VITE_GITHUB_REDIRECT_URI=http://localhost:5173/auth/callback/github
```

### Backend (`server/.env`)
```env
# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id_here
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
GOOGLE_REDIRECT_URI=http://localhost:5173/auth/callback/google

# GitHub OAuth
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
```

---

## Important Notes

### Redirect URIs Must Match Exactly

⚠️ **Critical**: The redirect URIs you configure in Google/GitHub must **exactly match** what your application uses, including:
- Protocol (`http://` vs `https://`)
- Domain (`localhost` vs your production domain)
- Port number (`:5173`, `:8080`, etc.)
- Path (`/auth/callback/google` or `/auth/callback/github`)

### Development vs Production

For **development**:
- Use `http://localhost:5173` (or your dev port)
- Add both to authorized redirect URIs

For **production**:
- Use your production domain (e.g., `https://yourdomain.com`)
- Make sure to add the production redirect URI to both providers

### Testing

1. Restart your development servers after adding environment variables
2. Try logging in with Google/GitHub
3. Check browser console and server logs for any errors
4. Common issues:
   - **"redirect_uri_mismatch"**: Check that redirect URIs match exactly
   - **"invalid_client"**: Check that client ID/secret are correct
   - **"access_denied"**: User cancelled the OAuth flow

### Security Best Practices

1. **Never commit** `.env` files to version control
2. Add `.env` to `.gitignore`
3. Use different credentials for development and production
4. Rotate secrets periodically
5. Use environment variables, never hardcode credentials

---

## Troubleshooting

### Google OAuth Issues

- **"redirect_uri_mismatch"**: 
  - Verify redirect URI in Google Console matches exactly
  - Check for trailing slashes, ports, protocols
  
- **"access_denied"**:
  - Check OAuth consent screen is published (or add test users)
  - Verify scopes are correctly configured

### GitHub OAuth Issues

- **"redirect_uri_mismatch"**:
  - Verify callback URL in GitHub OAuth app settings
  - Must match exactly including protocol and port

- **"bad_verification_code"**:
  - Usually means the authorization code expired (they expire quickly)
  - Try the flow again

### General Issues

- **Environment variables not loading**:
  - Restart your dev server
  - Check `.env` file is in the correct directory
  - Verify variable names match exactly (case-sensitive)

---

## Quick Reference Links

- [Google Cloud Console](https://console.cloud.google.com/)
- [Google OAuth Documentation](https://developers.google.com/identity/protocols/oauth2)
- [GitHub Developer Settings](https://github.com/settings/developers)
- [GitHub OAuth Documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)

---

## Next Steps

After setting up credentials:

1. ✅ Add credentials to `.env` files
2. ✅ Restart your development servers
3. ✅ Test OAuth login flows
4. ✅ Configure production credentials when deploying
5. ✅ Set up proper error handling and user feedback

Happy coding! 🚀

