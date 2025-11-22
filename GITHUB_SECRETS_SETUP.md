# GitHub Secrets Setup Guide

This guide will walk you through setting up all required GitHub secrets for automated deployment with GitHub Actions.

## Prerequisites

1. A GitHub repository with this project
2. A VPS server ready for deployment
3. SSH access to your VPS

## How to Add GitHub Secrets

1. Go to your GitHub repository
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Enter the secret name and value
5. Click **Add secret**

## Required Secrets

### 1. VPS Connection Secrets

These secrets allow GitHub Actions to connect to your VPS and deploy the application.

#### `VPS_HOST`
- **Description**: Your VPS IP address or domain name
- **Example**: `123.45.67.89` or `myserver.example.com`
- **How to get**: This is provided by your VPS hosting provider

#### `VPS_USERNAME`
- **Description**: SSH username for your VPS
- **Example**: `ubuntu`, `root`, or your custom username
- **Default**: Usually `ubuntu` for Ubuntu servers, `root` for other distributions

#### `VPS_SSH_KEY`
- **Description**: Private SSH key for authentication
- **How to generate**:

  ```bash
  # On your VPS, generate a new SSH key pair
  ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github-actions

  # Add the public key to authorized_keys
  cat ~/.ssh/github-actions.pub >> ~/.ssh/authorized_keys

  # Set proper permissions
  chmod 600 ~/.ssh/authorized_keys

  # Display the private key (copy this entire output to GitHub secret)
  cat ~/.ssh/github-actions
  ```

- **Important**: Copy the ENTIRE private key including the `-----BEGIN OPENSSH PRIVATE KEY-----` and `-----END OPENSSH PRIVATE KEY-----` lines

#### `VPS_PORT` (Optional)
- **Description**: SSH port number
- **Default**: `22`
- **When to set**: Only if you're using a custom SSH port

#### `DEPLOY_PATH`
- **Description**: Absolute path to deployment directory on VPS
- **Example**: `/opt/aimanage`
- **Default**: `/opt/aimanage`
- **Note**: Make sure this directory exists on your VPS and your user has write permissions

### 2. Frontend Environment Secrets

These secrets are used to build the frontend application with proper configuration.

#### `VITE_API_URL`
- **Description**: URL where your backend API is accessible
- **Examples**:
  - Production: `https://api.yourdomain.com` or `https://yourdomain.com`
  - If using same domain: `https://yourdomain.com`
- **Note**: Do NOT include trailing slash or `/api` path

#### `VITE_STRIPE_PUBLISHABLE_KEY`
- **Description**: Stripe publishable key for frontend
- **Format**: `pk_live_...` (production) or `pk_test_...` (development)
- **How to get**:
  1. Go to https://dashboard.stripe.com/apikeys
  2. Copy the "Publishable key"
- **Important**: Use test keys for development, live keys for production

#### `VITE_GOOGLE_CLIENT_ID`
- **Description**: Google OAuth client ID
- **Format**: Ends with `.apps.googleusercontent.com`
- **How to get**:
  1. Go to https://console.cloud.google.com/apis/credentials
  2. Create OAuth 2.0 Client ID (or use existing)
  3. Copy the "Client ID"
- **Note**: Must match the client ID in server environment

#### `VITE_GITHUB_CLIENT_ID`
- **Description**: GitHub OAuth client ID
- **How to get**:
  1. Go to https://github.com/settings/developers
  2. Click "OAuth Apps" → "New OAuth App" (or use existing)
  3. Copy the "Client ID"
- **Note**: Must match the client ID in server environment

#### `VITE_GITHUB_REDIRECT_URI`
- **Description**: GitHub OAuth callback URL
- **Format**: `https://yourdomain.com/auth/callback/github`
- **Important**: Must match the callback URL registered in your GitHub OAuth app

## Secret Verification Checklist

Before deploying, verify you have set up:

- [ ] `VPS_HOST` - Your VPS IP or domain
- [ ] `VPS_USERNAME` - SSH username
- [ ] `VPS_SSH_KEY` - Complete private SSH key
- [ ] `VPS_PORT` - SSH port (if not 22)
- [ ] `DEPLOY_PATH` - Deployment directory path
- [ ] `VITE_API_URL` - Backend API URL
- [ ] `VITE_STRIPE_PUBLISHABLE_KEY` - Stripe publishable key
- [ ] `VITE_GOOGLE_CLIENT_ID` - Google client ID
- [ ] `VITE_GITHUB_CLIENT_ID` - GitHub client ID
- [ ] `VITE_GITHUB_REDIRECT_URI` - GitHub OAuth callback URL

## VPS Environment Variables

In addition to GitHub secrets, you need to create a `.env` file on your VPS with backend configuration:

```bash
# On your VPS
cd /opt/aimanage  # or your DEPLOY_PATH
cp .env.example .env
nano .env  # Edit with your actual values
```

Required variables in VPS `.env`:
- `POSTGRES_USER` and `POSTGRES_PASSWORD` - Database credentials
- `JWT_SECRET` - JWT signing key (generate with `openssl rand -base64 32`)
- `STRIPE_SECRET_KEY` - Stripe secret key (sk_live_... or sk_test_...)
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook signing secret
- `STRIPE_MOCK_MODE` - Set to `false` for production
- `GOOGLE_CLIENT_SECRET` - Google OAuth secret
- `GITHUB_CLIENT_SECRET` - GitHub OAuth secret
- `HUGGINGFACE_TOKEN` - HuggingFace API token
- `ALLOWED_ORIGINS` - Your domain (https://yourdomain.com)
- `GITHUB_REPOSITORY` - Format: `username/repo`

## Testing Your Setup

### 1. Test SSH Connection

```bash
# On your local machine
ssh -i ~/.ssh/github-actions username@your-vps-host
```

If this works, your SSH key is correctly set up.

### 2. Test Deployment Workflow

1. Make a small change to your repository
2. Commit and push to main branch:
   ```bash
   git add .
   git commit -m "Test deployment"
   git push origin main
   ```
3. Go to your repository → **Actions** tab
4. Watch the deployment workflow run
5. Check for any errors in the workflow logs

### 3. Verify Deployment

After successful deployment, SSH into your VPS:

```bash
cd /opt/aimanage
docker compose ps  # Check if all containers are running
docker compose logs -f  # Check logs for errors
```

## Troubleshooting

### SSH Connection Failed

**Error**: `Permission denied (publickey)`

**Solution**:
1. Verify the private key is copied completely
2. Check that the public key is in `~/.ssh/authorized_keys` on VPS
3. Verify SSH service is running: `sudo systemctl status ssh`

### Docker Login Failed

**Error**: `Error response from daemon: Get "https://ghcr.io/v2/": unauthorized`

**Solution**:
1. Ensure your repository packages are public OR
2. Create a Personal Access Token (PAT) with `read:packages` scope
3. On VPS, login: `echo YOUR_PAT | docker login ghcr.io -u YOUR_USERNAME --password-stdin`

### Build Failed - Missing Environment Variables

**Error**: Build fails with missing VITE_ variables

**Solution**:
1. Double-check all `VITE_*` secrets are set in GitHub
2. Verify secret names match exactly (case-sensitive)
3. Make sure there are no spaces in secret values

### Containers Not Starting

**Error**: Containers exit immediately after starting

**Solution**:
1. Check VPS `.env` file exists and has correct values
2. Verify database credentials
3. Check logs: `docker compose logs server`
4. Ensure `GITHUB_REPOSITORY` is set correctly in VPS `.env`

## Security Best Practices

1. **Never commit secrets** to your repository
2. **Use different credentials** for development and production
3. **Rotate secrets regularly** (every 90 days recommended)
4. **Use strong passwords** for database and JWT secret
5. **Enable 2FA** on GitHub account
6. **Restrict SSH** to key-only authentication (disable password login)
7. **Use HTTPS** for production (set up SSL with Let's Encrypt)
8. **Monitor GitHub Actions logs** for any suspicious activity

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Deployment Guide](./DEPLOYMENT.md)

## Need Help?

If you encounter issues:

1. Check the GitHub Actions logs for detailed error messages
2. Review the DEPLOYMENT.md guide
3. Verify all secrets are set correctly
4. Test SSH connection manually
5. Check VPS logs: `docker compose logs -f`

---

**Ready to deploy!** Once all secrets are configured, push to main branch and watch your application deploy automatically! 🚀
