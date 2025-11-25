# Domain Setup Guide for AI Model Manager

## Overview
This guide will help you transition from IP-based access to a proper domain with SSL/HTTPS.

## Prerequisites
- A domain name (e.g., `aimanage.com`)
- Access to domain registrar's DNS settings
- SSH access to your VPS

---

## Option 1: Cloudflare (RECOMMENDED - Easiest)

### Why Cloudflare?
- ✅ Free SSL/TLS certificates (automatic)
- ✅ DDoS protection
- ✅ CDN (faster loading)
- ✅ Easy DNS management
- ✅ No manual certificate renewal needed

### Steps:

#### 1. Add Domain to Cloudflare
1. Sign up at https://cloudflare.com (free plan)
2. Click "Add a Site"
3. Enter your domain name
4. Choose the Free plan

#### 2. Update Nameservers
Cloudflare will give you 2 nameservers like:
```
ns1.cloudflare.com
ns2.cloudflare.com
```

Go to your domain registrar and update nameservers to these.

#### 3. Add DNS Records in Cloudflare
Add these A records:

| Type | Name | Content | Proxy | TTL |
|------|------|---------|-------|-----|
| A | @ | 109.199.115.1 | ✅ Proxied | Auto |
| A | www | 109.199.115.1 | ✅ Proxied | Auto |
| A | api | 109.199.115.1 | ✅ Proxied | Auto |

**Note:** Orange cloud (Proxied) = Cloudflare handles SSL automatically

#### 4. SSL/TLS Settings in Cloudflare
1. Go to SSL/TLS tab
2. Set SSL/TLS encryption mode to: **"Full"** (not "Full (strict)" yet)
3. Enable "Always Use HTTPS"

#### 5. Update Nginx Configuration on Server

SSH to your server and update nginx:

```bash
# Install certbot for origin certificates (optional but recommended)
sudo apt update
sudo apt install certbot python3-certbot-nginx -y

# Or use Cloudflare's Origin Certificates (easier)
```

**Update your nginx configuration:**

```nginx
# /etc/nginx/sites-available/aimanage

# Redirect HTTP to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name aimanage.com www.aimanage.com;
    return 301 https://$server_name$request_uri;
}

# Main app (HTTPS)
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name aimanage.com www.aimanage.com;

    # SSL configuration (Cloudflare handles this when proxied)
    ssl_certificate /etc/ssl/certs/cloudflare_origin.pem;
    ssl_certificate_key /etc/ssl/private/cloudflare_origin.key;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}

# API backend (HTTPS)
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.aimanage.com;

    ssl_certificate /etc/ssl/certs/cloudflare_origin.pem;
    ssl_certificate_key /etc/ssl/private/cloudflare_origin.key;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

**Or simpler - use same domain with /api path:**

```nginx
server {
    listen 443 ssl http2;
    server_name aimanage.com www.aimanage.com;

    # Frontend
    location / {
        proxy_pass http://127.0.0.1:3000;
        # ... proxy headers
    }

    # Backend API
    location /api/ {
        rewrite ^/api/(.*) /$1 break;
        proxy_pass http://127.0.0.1:8081;
        # ... proxy headers
    }
}
```

#### 6. Update Environment Variables

Update `.env` file:
```bash
# Backend .env
ALLOWED_ORIGINS=https://aimanage.com,https://www.aimanage.com

# Update OAuth redirect URIs
GOOGLE_REDIRECT_URI=https://aimanage.com/auth/callback/google
GITHUB_REDIRECT_URI=https://aimanage.com/auth/callback/github
```

Update GitHub Secrets for app build:
```bash
VITE_API_URL=https://api.aimanage.com
# or
VITE_API_URL=https://aimanage.com/api

VITE_GITHUB_REDIRECT_URI=https://aimanage.com/auth/callback/github
```

#### 7. Update OAuth Providers

**Google OAuth:**
1. Go to https://console.cloud.google.com/apis/credentials
2. Edit your OAuth 2.0 Client ID
3. Update Authorized JavaScript origins:
   - `https://aimanage.com`
4. Update Authorized redirect URIs:
   - `https://aimanage.com/auth/callback/google`

**GitHub OAuth:**
1. Go to https://github.com/settings/developers
2. Edit your OAuth App
3. Update Homepage URL: `https://aimanage.com`
4. Update Authorization callback URL: `https://aimanage.com/auth/callback/github`

**Stripe:**
1. Go to https://dashboard.stripe.com/webhooks
2. Update webhook endpoint: `https://api.aimanage.com/webhooks/stripe`
   (or `https://aimanage.com/api/webhooks/stripe`)

#### 8. Deploy Changes

```bash
# Update environment variables
git add .env
git commit -m "chore: Update environment variables for domain"
git push origin main

# Or rebuild and deploy manually
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
```

---

## Option 2: Let's Encrypt with Certbot (Free, Manual)

If you don't want to use Cloudflare:

### Steps:

#### 1. Point Domain to Server
Add A record in your DNS:
```
@ -> 109.199.115.1
www -> 109.199.115.1
```

#### 2. Install Certbot
```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx -y
```

#### 3. Get SSL Certificate
```bash
sudo certbot --nginx -d aimanage.com -d www.aimanage.com
```

Certbot will:
- Get SSL certificate
- Update nginx config automatically
- Set up auto-renewal

#### 4. Test Auto-Renewal
```bash
sudo certbot renew --dry-run
```

#### 5. Follow steps 6-8 from Option 1

---

## Quick Checklist

When you get your domain:

- [ ] Add domain to Cloudflare (or update DNS)
- [ ] Add A records pointing to 109.199.115.1
- [ ] Update nginx configuration
- [ ] Get SSL certificate (automatic with Cloudflare or use Certbot)
- [ ] Update `.env` with new domain
- [ ] Update GitHub repository secrets
- [ ] Update Google OAuth redirect URIs
- [ ] Update GitHub OAuth redirect URIs
- [ ] Update Stripe webhook URLs
- [ ] Test the deployment
- [ ] Enable "Always Use HTTPS" in Cloudflare

---

## Recommended Domain Structure

### Option A: Subdomain for API
```
Frontend: https://aimanage.com
Backend:  https://api.aimanage.com
```

### Option B: Path-based API (Simpler)
```
Frontend: https://aimanage.com
Backend:  https://aimanage.com/api
```

**Recommendation:** Use Option B (path-based) - it's simpler and requires only one domain.

---

## Testing After Setup

1. Visit `https://aimanage.com` - should load with HTTPS
2. Try Google login - should redirect properly
3. Try GitHub login - should work
4. Check browser console - no CORS errors
5. Test API calls - should use HTTPS

---

## Troubleshooting

**Issue: "Your connection is not private"**
- Wait for DNS propagation (up to 48 hours, usually 1-2 hours)
- Check SSL certificate is installed
- Check Cloudflare SSL/TLS mode is "Full"

**Issue: CORS errors after domain change**
- Update `ALLOWED_ORIGINS` in backend .env
- Rebuild and restart server container

**Issue: OAuth redirects to old URL**
- Update OAuth settings in Google/GitHub consoles
- Clear browser cache
- Update GitHub secrets and redeploy

---

## Estimated Time: 30 minutes - 2 hours
(Most time is waiting for DNS propagation)
