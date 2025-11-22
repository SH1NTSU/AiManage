# AiManage Deployment Guide

This guide will help you deploy AiManage to your VPS using GitHub Actions for automated CI/CD.

## Overview

The deployment uses:
- **Docker** for containerization
- **GitHub Actions** for CI/CD automation
- **GitHub Container Registry** for storing Docker images
- **SSH** for deploying to your VPS

## Prerequisites

1. A VPS server (Ubuntu 20.04+ recommended)
2. Domain name pointed to your VPS (optional but recommended)
3. GitHub account with this repository
4. SSH access to your VPS

## Step 1: Prepare Your VPS

### 1.1 Install Docker and Docker Compose

```bash
# Update packages
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Verify installation
docker --version
docker compose version
```

### 1.2 Create Deployment Directory

```bash
# Create deployment directory
sudo mkdir -p /opt/aimanage
sudo chown $USER:$USER /opt/aimanage
cd /opt/aimanage

# Clone your repository
git clone https://github.com/YOUR_USERNAME/YOUR_REPO.git .
```

### 1.3 Set Up Environment Variables

```bash
# Copy the example env file
cp .env.example .env

# Edit with your actual values
nano .env
```

Fill in all required values in `.env`:

```env
# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=YOUR_SECURE_PASSWORD
POSTGRES_DB=ai_db

# JWT (generate a secure random string)
JWT_SECRET=your_jwt_secret_key_minimum_32_characters

# Stripe
STRIPE_SECRET_KEY=sk_live_your_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
STRIPE_MOCK_MODE=false

# OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# HuggingFace
HUGGINGFACE_TOKEN=your_token

# CORS
ALLOWED_ORIGINS=https://yourdomain.com

# GitHub (for docker-compose.prod.yml)
GITHUB_REPOSITORY=YOUR_USERNAME/YOUR_REPO
```

### 1.4 Set Up Nginx Reverse Proxy (Optional but Recommended)

```bash
# Install Nginx
sudo apt install nginx -y

# Create Nginx configuration
sudo nano /etc/nginx/sites-available/aimanage
```

Add this configuration:

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/aimanage /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 1.5 Set Up SSL with Let's Encrypt (Recommended)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Obtain SSL certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Auto-renewal is set up automatically
```

## Step 2: Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add the following secrets:

### Required Secrets

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `VPS_HOST` | Your VPS IP address or domain | `123.45.67.89` |
| `VPS_USERNAME` | SSH username | `ubuntu` |
| `VPS_SSH_KEY` | Private SSH key for authentication | Contents of `~/.ssh/id_rsa` |
| `VPS_PORT` | SSH port (optional, default: 22) | `22` |
| `DEPLOY_PATH` | Path to deployment directory | `/opt/aimanage` |
| `GITHUB_TOKEN` | Automatically provided by GitHub | Auto-generated |

### Frontend Environment Secrets

| Secret Name | Description |
|-------------|-------------|
| `VITE_API_URL` | Backend API URL |
| `VITE_STRIPE_PUBLISHABLE_KEY` | Stripe publishable key |
| `VITE_GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `VITE_GITHUB_CLIENT_ID` | GitHub OAuth client ID |
| `VITE_GITHUB_REDIRECT_URI` | GitHub OAuth redirect URI |

### 2.1 Generate SSH Key for GitHub Actions

On your VPS:

```bash
# Generate SSH key pair
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github-actions

# Add public key to authorized_keys
cat ~/.ssh/github-actions.pub >> ~/.ssh/authorized_keys

# Display private key (copy this to GitHub secrets as VPS_SSH_KEY)
cat ~/.ssh/github-actions
```

## Step 3: Enable GitHub Container Registry

### 3.1 Make Repository Public or Configure Access

For private repositories:

```bash
# On your VPS, create a Personal Access Token (PAT)
# GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
# Generate new token with: read:packages, write:packages

# Login to GitHub Container Registry on VPS
echo YOUR_PAT | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
```

### 3.2 Update Package Visibility

1. Go to your GitHub profile → Packages
2. Find your `server` and `app` packages
3. Make them public OR grant access to your VPS

## Step 4: Initial Manual Deployment

Before using automated deployment, do an initial manual deployment:

```bash
cd /opt/aimanage

# Build and start services locally (development)
docker compose up -d

# OR use production compose file with pre-built images
export GITHUB_REPOSITORY=YOUR_USERNAME/YOUR_REPO
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d

# Check logs
docker compose logs -f

# Run database migrations if needed
docker compose exec server ./migrate -help
```

## Step 5: Test Automated Deployment

Now push changes to your repository:

```bash
git add .
git commit -m "Set up deployment"
git push origin main
```

GitHub Actions will:
1. Build Docker images for server and app
2. Push images to GitHub Container Registry
3. SSH into your VPS
4. Pull latest code and images
5. Restart services
6. Verify deployment

Monitor the deployment:
- Go to your repository → Actions tab
- Watch the workflow run

## Step 6: Database Migrations

For database migrations on your VPS:

```bash
cd /opt/aimanage

# Run migrations manually
docker compose exec server sh -c "cd migrations && migrate -help"

# Or if you have migrate CLI installed in the container
docker compose exec server ./server migrate up
```

## Troubleshooting

### Check Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f server
docker compose logs -f app
docker compose logs -f postgres
```

### Restart Services

```bash
cd /opt/aimanage
docker compose restart
```

### Clean Rebuild

```bash
cd /opt/aimanage
docker compose down -v  # WARNING: This removes volumes (database data)
docker compose up -d --build
```

### Check Container Status

```bash
docker compose ps
docker compose top
```

### Network Issues

```bash
# Check if ports are listening
sudo netstat -tlnp | grep -E '(3000|8081|5432)'

# Test backend
curl http://localhost:8081/health

# Test frontend
curl http://localhost:3000
```

### Disk Space

```bash
# Check disk usage
df -h

# Clean up Docker
docker system prune -a --volumes
```

## Updating the Application

Simply push to the main branch:

```bash
git add .
git commit -m "Update feature"
git push origin main
```

GitHub Actions will automatically:
- Build new images
- Deploy to your VPS
- Restart services

## Rollback

If deployment fails:

```bash
cd /opt/aimanage

# Check previous images
docker images

# Use specific version
docker compose down
# Edit docker-compose.prod.yml to use specific image tag
docker compose up -d
```

## Security Checklist

- [ ] Change default database password
- [ ] Use strong JWT secret (32+ characters)
- [ ] Enable SSL/TLS with Let's Encrypt
- [ ] Configure firewall (UFW)
- [ ] Set up fail2ban
- [ ] Regular backups of database
- [ ] Keep Docker and system updated
- [ ] Use environment variables for secrets
- [ ] Restrict SSH access (key-only, disable root)
- [ ] Configure CORS properly

## Backup Strategy

```bash
# Backup database
docker compose exec -T postgres pg_dump -U postgres ai_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup uploads
tar -czf uploads_backup_$(date +%Y%m%d_%H%M%S).tar.gz uploads/

# Backup .env
cp .env .env.backup
```

## Monitoring

Consider setting up:
- **Uptime monitoring**: UptimeRobot, Pingdom
- **Error tracking**: Sentry
- **Logs**: Loki, ELK stack
- **Metrics**: Prometheus + Grafana

## Support

For issues:
1. Check logs: `docker compose logs -f`
2. Verify GitHub Actions workflow
3. Check GitHub repository issues
4. Review this documentation

---

**Happy Deploying! 🚀**
