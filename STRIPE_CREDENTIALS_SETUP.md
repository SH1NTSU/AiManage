# Stripe Credentials Setup Guide

This guide will help you set up Stripe for real payments in your application.

## 📍 Where to Get Stripe Credentials

### 1. Create a Stripe Account

1. Go to [https://stripe.com](https://stripe.com)
2. Click **"Sign up"** or **"Start now"**
3. Fill in your business information:
   - Business name
   - Email address
   - Password
   - Country/Region

### 2. Access Your API Keys

Once logged in:

1. Navigate to the **Developers** section:
   - Click your profile icon (top right)
   - Select **"Developers"** from the dropdown
   - Or go directly to: [https://dashboard.stripe.com/apikeys](https://dashboard.stripe.com/apikeys)

2. You'll see two types of keys:
   - **Publishable key** (starts with `pk_test_` for test mode, `pk_live_` for live mode)
   - **Secret key** (starts with `sk_test_` for test mode, `sk_live_` for live mode)

### 3. Test Mode vs Live Mode

**Test Mode** (Recommended for development):
- Use test keys (start with `pk_test_` and `sk_test_`)
- No real charges are made
- Use test card numbers (see below)
- Switch using the toggle in the top right of the Stripe dashboard

**Live Mode** (For production):
- Use live keys (start with `pk_live_` and `sk_live_`)
- Real charges are made
- Requires account verification
- Switch using the toggle in the top right

## 🔑 Required Environment Variables

### Frontend (`app/.env`)

Add your Stripe **Publishable Key**:

```env
VITE_STRIPE_PUBLISHABLE_KEY=pk_test_your_publishable_key_here
```

**Important**: 
- Use `pk_test_` for test mode
- Use `pk_live_` for production
- Never commit this file to git (it should be in `.gitignore`)

### Backend (`server/.env`)

Add your Stripe **Secret Key**:

```env
STRIPE_SECRET_KEY=sk_test_your_secret_key_here
FRONTEND_URL=http://localhost:5173
```

**Important**:
- Use `sk_test_` for test mode
- Use `sk_live_` for production
- **NEVER** expose this key in frontend code
- Never commit this file to git

## 🧪 Test Card Numbers

For testing payments in **Test Mode**, use these card numbers:

### Successful Payment
- **Card Number**: `4242 4242 4242 4242`
- **Expiry**: Any future date (e.g., `12/34`)
- **CVC**: Any 3 digits (e.g., `123`)
- **ZIP**: Any 5 digits (e.g., `12345`)

### Declined Payment
- **Card Number**: `4000 0000 0000 0002`
- **Expiry**: Any future date
- **CVC**: Any 3 digits
- **ZIP**: Any 5 digits

### Requires Authentication (3D Secure)
- **Card Number**: `4000 0027 6000 3184`
- **Expiry**: Any future date
- **CVC**: Any 3 digits
- **ZIP**: Any 5 digits

More test cards: [https://stripe.com/docs/testing](https://stripe.com/docs/testing)

## 📝 Setup Steps

### Step 1: Get Your Keys

1. Log in to [Stripe Dashboard](https://dashboard.stripe.com)
2. Go to **Developers** → **API keys**
3. Copy your **Publishable key** (starts with `pk_test_`)
4. Click **"Reveal test key"** and copy your **Secret key** (starts with `sk_test_`)

### Step 2: Configure Frontend

1. Open `app/.env` (create if it doesn't exist)
2. Add:
   ```env
   VITE_STRIPE_PUBLISHABLE_KEY=pk_test_your_actual_key_here
   ```
3. Replace `pk_test_your_actual_key_here` with your actual publishable key
4. **Restart your frontend dev server** (Vite doesn't hot-reload `.env` changes)

### Step 3: Configure Backend

1. Open `server/.env` (create if it doesn't exist)
2. Add:
   ```env
   STRIPE_SECRET_KEY=sk_test_your_actual_key_here
   FRONTEND_URL=http://localhost:5173
   ```
3. Replace `sk_test_your_actual_key_here` with your actual secret key
4. **Restart your backend server**

### Step 4: Verify Setup

1. Start both frontend and backend servers
2. Try to purchase a paid model or upgrade subscription
3. Use test card `4242 4242 4242 4242` for testing
4. Check backend logs for Stripe API calls

## 🔒 Security Best Practices

1. **Never commit `.env` files** to git
2. **Use test keys** during development
3. **Rotate keys** if accidentally exposed
4. **Use environment variables** in production (not hardcoded)
5. **Secret key** should only be on the backend
6. **Publishable key** is safe to expose in frontend code

## 🚀 Going Live (Production)

When ready for production:

1. **Complete Stripe account verification**:
   - Add business information
   - Verify identity
   - Add bank account for payouts

2. **Switch to Live Mode**:
   - Toggle "Test mode" off in Stripe dashboard
   - Copy your **live** publishable key (`pk_live_...`)
   - Copy your **live** secret key (`sk_live_...`)

3. **Update environment variables**:
   - Update `app/.env` with live publishable key
   - Update `server/.env` with live secret key
   - Update `FRONTEND_URL` to your production domain

4. **Test with real card** (small amount):
   - Use your own card for a small test purchase
   - Verify payment appears in Stripe dashboard
   - Verify webhook events are received (if configured)

## 📚 Additional Resources

- [Stripe Documentation](https://stripe.com/docs)
- [Stripe Testing Guide](https://stripe.com/docs/testing)
- [Stripe Dashboard](https://dashboard.stripe.com)
- [Stripe API Reference](https://stripe.com/docs/api)

## ❓ Troubleshooting

### "Payment processing not configured"
- Check that `STRIPE_SECRET_KEY` is set in `server/.env`
- Restart backend server after adding the key

### "Stripe is not defined" or "Stripe Elements not loading"
- Check that `VITE_STRIPE_PUBLISHABLE_KEY` is set in `app/.env`
- Restart frontend dev server after adding the key
- Verify the key starts with `pk_test_` or `pk_live_`

### Payment fails with "Your card was declined"
- Make sure you're using test mode with test cards
- Use card number `4242 4242 4242 4242` for successful test payments
- Check Stripe dashboard logs for detailed error messages

### "Invalid API Key"
- Verify you copied the entire key (no spaces or line breaks)
- Check if you're mixing test and live keys
- Ensure keys match the mode (test/live) you're using

## 📞 Support

- Stripe Support: [https://support.stripe.com](https://support.stripe.com)
- Stripe Status: [https://status.stripe.com](https://status.stripe.com)

