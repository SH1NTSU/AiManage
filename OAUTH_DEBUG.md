# OAuth Debugging Guide

## Common Issues and Solutions

### 1. Check Environment Variables

**Frontend (`app/.env`):**
```bash
# Check if variables are loaded
echo $VITE_GOOGLE_CLIENT_ID
echo $VITE_GITHUB_CLIENT_ID
```

**Backend (`server/.env`):**
```bash
# Check if variables are loaded
echo $GOOGLE_CLIENT_ID
echo $GOOGLE_CLIENT_SECRET
echo $GITHUB_CLIENT_ID
echo $GITHUB_CLIENT_SECRET
```

**Important**: After updating `.env` files, you MUST restart both frontend and backend servers!

### 2. Check Browser Console

Open browser DevTools (F12) and check:
- **Console tab**: Look for errors
- **Network tab**: Check OAuth requests/responses
- **Application tab**: Check sessionStorage for `oauth_state_*`

### 3. Check Server Logs

Look for error messages starting with `❌` in your backend logs:
- Token exchange errors
- OAuth errors
- Missing credentials

### 4. Verify Redirect URIs Match Exactly

**Google:**
- In Google Cloud Console → Credentials → Your OAuth Client
- Check "Authorized redirect URIs" includes: `http://localhost:5173/auth/callback/google`
- Must match EXACTLY (including protocol, port, path)

**GitHub:**
- In GitHub → Settings → Developer settings → OAuth Apps → Your App
- Check "Authorization callback URL" is: `http://localhost:5173/auth/callback/github`
- Must match EXACTLY

### 5. Common Error Messages

#### "redirect_uri_mismatch"
- **Cause**: Redirect URI in OAuth provider doesn't match what's being sent
- **Fix**: Update redirect URI in Google/GitHub settings to match exactly

#### "invalid_client"
- **Cause**: Client ID or Client Secret is incorrect
- **Fix**: Double-check credentials in `.env` files

#### "Invalid state parameter"
- **Cause**: State mismatch (CSRF protection)
- **Fix**: Clear browser sessionStorage and try again

#### "No access token received"
- **Cause**: OAuth provider didn't return access token
- **Fix**: Check server logs for detailed error from provider

### 6. Test OAuth Flow Manually

1. Click "Sign in with Google/GitHub"
2. Check browser redirects to provider login
3. After login, check redirect back to `/auth/callback/google` or `/auth/callback/github`
4. Check URL parameters: `?code=...&state=...`
5. Check browser console for errors
6. Check server logs for backend errors

### 7. Verify Port Numbers

Make sure your frontend is running on the port specified in redirect URIs:
- Default: `http://localhost:5173` (Vite default)
- If using different port, update all redirect URIs accordingly

### 8. Check CORS Issues

If you see CORS errors:
- Make sure backend allows requests from frontend origin
- Check backend CORS configuration

### 9. Debug Steps

1. **Clear browser storage**:
   ```javascript
   // In browser console
   localStorage.clear();
   sessionStorage.clear();
   ```

2. **Check if credentials are loaded**:
   ```javascript
   // In browser console (on Auth page)
   console.log('Google Client ID:', import.meta.env.VITE_GOOGLE_CLIENT_ID);
   console.log('GitHub Client ID:', import.meta.env.VITE_GITHUB_CLIENT_ID);
   ```

3. **Test OAuth URL generation**:
   ```javascript
   // In browser console
   import { signIn } from './lib/auth';
   // Check the generated URL before redirect
   ```

4. **Check backend environment variables**:
   ```bash
   # In server directory
   cd server
   go run cmd/server/main.go
   # Check startup logs for OAuth config
   ```

### 10. Quick Fix Checklist

- [ ] Restarted frontend server after updating `.env`
- [ ] Restarted backend server after updating `.env`
- [ ] Redirect URIs match exactly in OAuth provider settings
- [ ] Client IDs and Secrets are correct
- [ ] No typos in environment variable names
- [ ] Browser cache cleared
- [ ] Checked browser console for errors
- [ ] Checked server logs for errors
- [ ] Port numbers match in all configurations

### 11. Still Not Working?

1. **Enable verbose logging**:
   - Check server logs for detailed error messages
   - Check browser network tab for failed requests

2. **Test with curl**:
   ```bash
   # Test Google OAuth (replace with your values)
   curl -X POST http://localhost:8081/v1/auth/google \
     -H "Content-Type: application/json" \
     -d '{"code":"test_code","redirect_uri":"http://localhost:5173/auth/callback/google"}'
   ```

3. **Check OAuth provider status**:
   - Google: https://status.cloud.google.com/
   - GitHub: https://www.githubstatus.com/

4. **Verify OAuth app is active**:
   - Google: Check OAuth consent screen is published (or add test users)
   - GitHub: Check OAuth app is not suspended

