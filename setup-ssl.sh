#!/bin/bash

# Wait for DNS propagation and then run this script to set up SSL

echo "Checking DNS resolution..."
if ! nslookup aimanage.online | grep -q "109.199.115.1"; then
    echo "ERROR: DNS is not yet pointing to this server (109.199.115.1)"
    echo "Please wait for DNS propagation and try again."
    echo "You can check DNS status with: nslookup aimanage.online"
    exit 1
fi

echo "DNS is configured correctly!"
echo "Setting up SSL certificate with Let's Encrypt..."

sudo certbot --nginx -d aimanage.online -d www.aimanage.online

echo ""
echo "Setup complete! Your site should now be accessible at:"
echo "  https://aimanage.online"
echo "  https://www.aimanage.online"
