# Environment Variables Setup Instructions

## Important: Vite Environment Variables

Vite only loads environment variables that:
1. Start with `VITE_` prefix
2. Are in the `.env` file in the `app` directory
3. Require **server restart** after changes

## Steps to Fix OAuth

### 1. Update `.env` file

Edit `app/.env` and replace the placeholder values:

```env
# GitHub OAuth - Replace with your actual Client ID
VITE_GITHUB_CLIENT_ID=your_actual_github_client_id_here

# Google OAuth - Replace with your actual Client ID  
VITE_GOOGLE_CLIENT_ID=your_actual_google_client_id_here

# Redirect URI (usually correct as-is)
VITE_GITHUB_REDIRECT_URI=http://localhost:5173/auth/callback/github
```

### 2. **CRITICAL: Restart Vite Dev Server**

After updating `.env`:
1. **Stop** your Vite dev server (Ctrl+C)
2. **Start** it again: `npm run dev`

**Vite does NOT hot-reload `.env` changes!** You MUST restart the server.

### 3. Verify Variables are Loaded

Open browser console (F12) and check:
- Look for `🔍 OAuth Configuration Check` log
- Should show your actual Client ID (first 10 chars), not "your_github_client_id_here"

### 4. Check Backend `.env`

Also update `server/.env`:

```env
# GitHub OAuth
GITHUB_CLIENT_ID=your_actual_github_client_id_here
GITHUB_CLIENT_SECRET=your_actual_github_client_secret_here

# Google OAuth
GOOGLE_CLIENT_ID=your_actual_google_client_id_here
GOOGLE_CLIENT_SECRET=your_actual_google_client_secret_here
GOOGLE_REDIRECT_URI=http://localhost:5173/auth/callback/google
```

**Restart backend server** after updating `server/.env` too!

## Quick Test

After restarting both servers:

1. Open browser console (F12)
2. Go to Auth page
3. Look for: `🔍 OAuth Configuration Check`
4. Should see: `Google: ✅ Configured` and `GitHub: ✅ Configured`
5. Client IDs should show actual values, not placeholders

## Troubleshooting

If still showing placeholder values:

1. **Double-check `.env` file location**: Must be in `app/` directory
2. **Check for typos**: Variable names are case-sensitive
3. **No spaces**: `VITE_GITHUB_CLIENT_ID=value` (no spaces around `=`)
4. **No quotes needed**: Just `VITE_GITHUB_CLIENT_ID=abc123`, not `VITE_GITHUB_CLIENT_ID="abc123"`
5. **Restart required**: Vite must be restarted after `.env` changes

