#!/bin/bash

# AiManage VPS Setup Script
# This script automates the initial VPS setup for AiManage deployment

set -e

echo "🚀 AiManage VPS Setup Script"
echo "============================"
echo ""

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "❌ Please don't run this script as root"
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check system
echo "📋 System Information:"
echo "OS: $(lsb_release -d | cut -f2)"
echo "Kernel: $(uname -r)"
echo ""

# Update system
print_info "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install Docker
if ! command -v docker &> /dev/null; then
    print_info "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
    print_info "Docker installed successfully"
else
    print_warning "Docker is already installed"
fi

# Install Docker Compose plugin
if ! docker compose version &> /dev/null; then
    print_info "Installing Docker Compose..."
    sudo apt install docker-compose-plugin -y
    print_info "Docker Compose installed successfully"
else
    print_warning "Docker Compose is already installed"
fi

# Install other useful tools
print_info "Installing additional tools..."
sudo apt install -y git curl wget nginx certbot python3-certbot-nginx ufw

# Configure firewall
print_info "Configuring firewall..."
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw --force enable

# Create deployment directory
DEPLOY_DIR="/opt/aimanage"
print_info "Creating deployment directory at $DEPLOY_DIR..."
sudo mkdir -p $DEPLOY_DIR
sudo chown $USER:$USER $DEPLOY_DIR

# Prompt for GitHub repository
read -p "Enter your GitHub repository (username/repo): " GITHUB_REPO

if [ -z "$GITHUB_REPO" ]; then
    print_error "GitHub repository is required"
    exit 1
fi

# Clone repository
print_info "Cloning repository..."
cd $DEPLOY_DIR
git clone "https://github.com/$GITHUB_REPO.git" .

# Setup environment file
if [ ! -f .env ]; then
    print_info "Creating .env file..."
    cp .env.example .env

    # Generate random secrets
    JWT_SECRET=$(openssl rand -base64 48)
    DB_PASSWORD=$(openssl rand -base64 32)

    # Update .env with generated secrets
    sed -i "s/your_jwt_secret_key_here_minimum_32_characters/$JWT_SECRET/" .env
    sed -i "s/your_secure_password_here/$DB_PASSWORD/" .env
    sed -i "s|YOUR_USERNAME/YOUR_REPO|$GITHUB_REPO|" .env

    print_warning "Please edit .env file and add your API keys and secrets"
    echo "  - Stripe keys"
    echo "  - OAuth credentials"
    echo "  - Domain configuration"
    echo ""
    echo "Edit with: nano $DEPLOY_DIR/.env"
else
    print_warning ".env file already exists"
fi

# Generate SSH key for GitHub Actions
SSH_KEY_PATH="$HOME/.ssh/github-actions"
if [ ! -f "$SSH_KEY_PATH" ]; then
    print_info "Generating SSH key for GitHub Actions..."
    ssh-keygen -t ed25519 -C "github-actions" -f "$SSH_KEY_PATH" -N ""
    cat "${SSH_KEY_PATH}.pub" >> "$HOME/.ssh/authorized_keys"
    chmod 600 "$HOME/.ssh/authorized_keys"

    print_info "SSH key generated. Add this private key to GitHub Secrets as VPS_SSH_KEY:"
    echo ""
    cat "$SSH_KEY_PATH"
    echo ""
else
    print_warning "SSH key already exists at $SSH_KEY_PATH"
fi

# Setup Nginx configuration
read -p "Enter your domain name (e.g., example.com) or press Enter to skip: " DOMAIN

if [ ! -z "$DOMAIN" ]; then
    NGINX_CONF="/etc/nginx/sites-available/aimanage"

    if [ ! -f "$NGINX_CONF" ]; then
        print_info "Creating Nginx configuration..."

        sudo tee $NGINX_CONF > /dev/null <<EOF
server {
    listen 80;
    server_name $DOMAIN www.$DOMAIN;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # WebSocket
    location /ws {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

        sudo ln -s $NGINX_CONF /etc/nginx/sites-enabled/
        sudo nginx -t && sudo systemctl restart nginx

        print_info "Nginx configured successfully"

        # Setup SSL
        read -p "Do you want to setup SSL with Let's Encrypt? (y/n): " SETUP_SSL
        if [ "$SETUP_SSL" = "y" ]; then
            print_info "Setting up SSL..."
            sudo certbot --nginx -d $DOMAIN -d www.$DOMAIN
        fi
    else
        print_warning "Nginx configuration already exists"
    fi
fi

# Summary
echo ""
echo "======================================"
print_info "Setup completed successfully!"
echo "======================================"
echo ""
echo "📝 Next Steps:"
echo ""
echo "1. Edit environment variables:"
echo "   nano $DEPLOY_DIR/.env"
echo ""
echo "2. Add these GitHub Secrets:"
echo "   - VPS_HOST: $(curl -s ifconfig.me)"
echo "   - VPS_USERNAME: $USER"
echo "   - VPS_SSH_KEY: (shown above)"
echo "   - DEPLOY_PATH: $DEPLOY_DIR"
echo ""
echo "3. Configure GitHub Container Registry access:"
echo "   Create a Personal Access Token with read:packages, write:packages"
echo "   echo YOUR_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
echo ""
echo "4. Start the application:"
echo "   cd $DEPLOY_DIR"
echo "   export GITHUB_REPOSITORY=$GITHUB_REPO"
echo "   docker compose up -d"
echo ""
echo "5. View logs:"
echo "   docker compose logs -f"
echo ""
echo "📚 Full documentation: $DEPLOY_DIR/DEPLOYMENT.md"
echo ""
