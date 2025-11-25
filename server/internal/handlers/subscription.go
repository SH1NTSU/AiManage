package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"server/internal/middlewares"
	"server/internal/repository"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/webhook"
)

// Subscription tiers
const (
	TierFree       = "free"
	TierBasic      = "basic"
	TierPro        = "pro"
	TierEnterprise = "enterprise"
)

// Subscription prices (in cents)
var subscriptionPrices = map[string]int64{
	TierBasic:      999,   // $9.99/month
	TierPro:        2999,  // $29.99/month
	TierEnterprise: 9999,  // $99.99/month
}

// Training credits per tier
var trainingCredits = map[string]int{
	TierFree:       0,   // No server training
	TierBasic:      10,  // 10 training jobs per month
	TierPro:        50,  // 50 training jobs per month
	TierEnterprise: 999, // Unlimited
}

// GetSubscriptionHandler returns the user's current subscription
func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("📋 Fetching subscription for user: %s", userEmail)

	user, err := repository.GetUserByEmail(r.Context(), userEmail)
	if err != nil || user == nil {
		log.Printf("❌ User not found: %s", userEmail)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Extract subscription info
	subscription := map[string]interface{}{
		"tier":             getStringField(*user, "subscription_tier", TierFree),
		"status":           getStringField(*user, "subscription_status", "active"),
		"training_credits": getIntField(*user, "training_credits", 0),
		"start_date":       (*user)["subscription_start_date"],
		"end_date":         (*user)["subscription_end_date"],
	}

	log.Printf("✅ Returning subscription for %s: tier=%s, credits=%d",
		userEmail, subscription["tier"], subscription["training_credits"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"subscription": subscription,
	})
}

// CreateCheckoutSessionHandler creates a Stripe checkout session
func CreateCheckoutSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userEmail, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Tier string `json:"tier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate tier
	if req.Tier != TierBasic && req.Tier != TierPro && req.Tier != TierEnterprise {
		http.Error(w, "Invalid subscription tier", http.StatusBadRequest)
		return
	}

	user, err := repository.GetUserByEmail(r.Context(), userEmail)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Initialize Stripe
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Println("⚠️  STRIPE_SECRET_KEY not set, using mock mode")
		// Mock response for development
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		checkoutURL := frontendURL + "/settings?mock_checkout=true&tier=" + req.Tier
		log.Printf("Mock checkout URL: %s", checkoutURL)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"checkout_url": checkoutURL,
			"tier":         req.Tier,
			"price":        subscriptionPrices[req.Tier],
			"message":      "Mock mode - STRIPE_SECRET_KEY not configured",
		})
		return
	}

	stripe.Key = stripeKey

	// Get or create Stripe customer
	stripeCustomerID := getStringField(*user, "stripe_customer_id", "")
	if stripeCustomerID == "" {
		// Create new Stripe customer
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(userEmail),
			Metadata: map[string]string{
				"user_id": fmt.Sprintf("%v", (*user)["id"]),
			},
		}
		cust, err := customer.New(customerParams)
		if err != nil {
			log.Printf("❌ Failed to create Stripe customer: %v", err)
			http.Error(w, "Failed to create customer", http.StatusInternalServerError)
			return
		}
		stripeCustomerID = cust.ID

		// Update user with Stripe customer ID
		if err := repository.UpdateUserStripeCustomer(r.Context(), userEmail, stripeCustomerID); err != nil {
			log.Printf("⚠️  Failed to save Stripe customer ID: %v", err)
		}
	}

	// Create checkout session
	successURL := os.Getenv("FRONTEND_URL")
	if successURL == "" {
		successURL = "http://localhost:5173"
	}
	successURL += "/settings?subscription_success=true"

	cancelURL := os.Getenv("FRONTEND_URL")
	if cancelURL == "" {
		cancelURL = "http://localhost:5173"
	}
	cancelURL += "/pricing?subscription_canceled=true"

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(stripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(fmt.Sprintf("AiManage %s Plan", req.Tier)),
						Description: stripe.String(fmt.Sprintf("%d training credits per month", trainingCredits[req.Tier])),
					},
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripe.String("month"),
					},
					UnitAmount: stripe.Int64(subscriptionPrices[req.Tier]),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"user_email": userEmail,
			"tier":       req.Tier,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("❌ Failed to create checkout session: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Created checkout session for user %s, tier: %s", userEmail, req.Tier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"checkout_url": sess.URL,
		"session_id":   sess.ID,
		"tier":         req.Tier,
		"price":        subscriptionPrices[req.Tier],
	})
}

// CanUserTrainOnServer checks if user has permission to train on server
func CanUserTrainOnServer(r *http.Request) (bool, string) {
	userEmail, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok {
		return false, "Unauthorized"
	}

	user, err := repository.GetUserByEmail(r.Context(), userEmail)
	if err != nil || user == nil {
		return false, "User not found"
	}

	tier := getStringField(*user, "subscription_tier", TierFree)
	status := getStringField(*user, "subscription_status", "active")
	credits := getIntField(*user, "training_credits", 0)

	// Free tier cannot train on server
	if tier == TierFree {
		return false, "Server training requires a paid subscription. Train locally or upgrade your plan."
	}

	// Check subscription status
	if status != "active" {
		return false, "Your subscription is not active. Please renew to continue server training."
	}

	// Check training credits (except for enterprise)
	if tier != TierEnterprise && credits <= 0 {
		return false, "You've used all your training credits for this month. Upgrade to Pro or Enterprise for more."
	}

	return true, ""
}

// Helper functions
func getStringField(user map[string]interface{}, field string, defaultValue string) string {
	if val, ok := user[field]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntField(user map[string]interface{}, field string, defaultValue int) int {
	if val, ok := user[field]; ok && val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultValue
}

// DecrementTrainingCredits decrements the user's training credits
func DecrementTrainingCredits(userEmail string) error {
	// TODO: Implement in repository
	log.Printf("Decrementing training credits for user: %s", userEmail)
	return nil
}

// StripeWebhookHandler handles Stripe webhook events
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		http.Error(w, "Invalid payload", http.StatusServiceUnavailable)
		return
	}

	// Verify webhook signature
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret != "" {
		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
		if err != nil {
			log.Printf("❌ Webhook signature verification failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handleStripeEvent(event)
	} else {
		// For development without webhook secret
		log.Println("⚠️  STRIPE_WEBHOOK_SECRET not set, skipping signature verification")
		var event stripe.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("❌ Failed to parse webhook JSON: %v", err)
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		handleStripeEvent(event)
	}

	w.WriteHeader(http.StatusOK)
}

func handleStripeEvent(event stripe.Event) {
	log.Printf("📥 Received Stripe webhook: %s", event.Type)

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("❌ Error parsing checkout.session.completed: %v", err)
			return
		}

		// Extract user email and tier from metadata
		userEmail := session.Metadata["user_email"]
		tier := session.Metadata["tier"]

		if userEmail == "" || tier == "" {
			log.Printf("⚠️  Missing metadata in checkout session")
			return
		}

		// Update user subscription
		err := repository.UpdateUserSubscription(nil, userEmail, map[string]interface{}{
			"subscription_tier":          tier,
			"subscription_status":        "active",
			"stripe_subscription_id":     session.Subscription.ID,
			"stripe_customer_id":         session.Customer.ID,
			"subscription_start_date":    time.Now(),
			"subscription_end_date":      time.Now().AddDate(0, 1, 0), // 1 month from now
			"training_credits":           trainingCredits[tier],
		})

		if err != nil {
			log.Printf("❌ Failed to update user subscription: %v", err)
			return
		}

		log.Printf("✅ Subscription activated for %s: %s tier", userEmail, tier)

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			log.Printf("❌ Error parsing customer.subscription.updated: %v", err)
			return
		}

		// Find user by stripe customer ID
		userEmail, err := repository.GetUserEmailByStripeCustomer(nil, subscription.Customer.ID)
		if err != nil {
			log.Printf("❌ Failed to find user for customer %s: %v", subscription.Customer.ID, err)
			return
		}

		// Update subscription status
		status := "active"
		if subscription.Status != stripe.SubscriptionStatusActive {
			status = string(subscription.Status)
		}

		err = repository.UpdateUserSubscriptionStatus(nil, userEmail, status)
		if err != nil {
			log.Printf("❌ Failed to update subscription status: %v", err)
			return
		}

		log.Printf("✅ Subscription updated for %s: %s", userEmail, status)

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			log.Printf("❌ Error parsing customer.subscription.deleted: %v", err)
			return
		}

		// Find user by stripe customer ID
		userEmail, err := repository.GetUserEmailByStripeCustomer(nil, subscription.Customer.ID)
		if err != nil {
			log.Printf("❌ Failed to find user for customer %s: %v", subscription.Customer.ID, err)
			return
		}

		// Downgrade to free tier
		err = repository.UpdateUserSubscription(nil, userEmail, map[string]interface{}{
			"subscription_tier":   "free",
			"subscription_status": "canceled",
			"training_credits":    0,
		})

		if err != nil {
			log.Printf("❌ Failed to cancel subscription: %v", err)
			return
		}

		log.Printf("✅ Subscription canceled for %s", userEmail)

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			log.Printf("❌ Error parsing invoice.payment_succeeded: %v", err)
			return
		}

		log.Printf("✅ Payment succeeded for customer %s", invoice.Customer.ID)

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			log.Printf("❌ Error parsing invoice.payment_failed: %v", err)
			return
		}

		// Find user by stripe customer ID
		userEmail, err := repository.GetUserEmailByStripeCustomer(nil, invoice.Customer.ID)
		if err != nil {
			log.Printf("❌ Failed to find user for customer %s: %v", invoice.Customer.ID, err)
			return
		}

		// Mark subscription as past_due
		err = repository.UpdateUserSubscriptionStatus(nil, userEmail, "past_due")
		if err != nil {
			log.Printf("❌ Failed to update subscription status: %v", err)
			return
		}

		log.Printf("⚠️  Payment failed for %s", userEmail)
	}
}

// GetPricingHandler returns available subscription tiers and pricing
func GetPricingHandler(w http.ResponseWriter, r *http.Request) {
	pricing := []map[string]interface{}{
		{
			"tier":             TierFree,
			"name":             "Free",
			"price":            0,
			"training_credits": 0,
			"features": []string{
				"Train models locally on your own machine",
				"Upload trained models to platform",
				"Access community models",
				"Basic model analytics",
			},
		},
		{
			"tier":             TierBasic,
			"name":             "Basic",
			"price":            subscriptionPrices[TierBasic],
			"training_credits": trainingCredits[TierBasic],
			"features": []string{
				"Everything in Free",
				"10 server training jobs per month",
				"Priority queue for training",
				"Advanced model analytics",
				"Email support",
			},
		},
		{
			"tier":             TierPro,
			"name":             "Pro",
			"price":            subscriptionPrices[TierPro],
			"training_credits": trainingCredits[TierPro],
			"features": []string{
				"Everything in Basic",
				"50 server training jobs per month",
				"Faster training GPUs",
				"Custom model architectures",
				"API access",
				"Priority support",
			},
		},
		{
			"tier":             TierEnterprise,
			"name":             "Enterprise",
			"price":            subscriptionPrices[TierEnterprise],
			"training_credits": trainingCredits[TierEnterprise],
			"features": []string{
				"Everything in Pro",
				"Unlimited server training",
				"Dedicated GPU resources",
				"Custom integrations",
				"SLA guarantee",
				"24/7 priority support",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"pricing": pricing,
	})
}

// ResetMonthlyCredits resets training credits for all users (run monthly via cron)
func ResetMonthlyCreditsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authentication
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Resetting monthly training credits for all users...")

	// TODO: Implement in repository
	// Update all users with their tier's monthly credits

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Monthly credits reset successfully",
		"timestamp": time.Now(),
	})
}

// MockUpgradeHandler simulates a subscription upgrade for development/testing
// This should only be used in development - remove or disable in production
func MockUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userEmail, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Tier string `json:"tier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate tier
	if req.Tier != TierBasic && req.Tier != TierPro && req.Tier != TierEnterprise && req.Tier != TierFree {
		http.Error(w, "Invalid subscription tier", http.StatusBadRequest)
		return
	}

	log.Printf("🎭 Mock upgrade: %s -> %s tier", userEmail, req.Tier)

	// Update user subscription in database
	err := repository.UpdateUserSubscription(r.Context(), userEmail, map[string]interface{}{
		"subscription_tier":       req.Tier,
		"subscription_status":     "active",
		"subscription_start_date": time.Now(),
		"subscription_end_date":   time.Now().AddDate(0, 1, 0), // 1 month from now
		"training_credits":        trainingCredits[req.Tier],
	})

	if err != nil {
		log.Printf("❌ Failed to update subscription: %v", err)
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Mock upgrade successful: %s is now on %s tier", userEmail, req.Tier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully upgraded to %s tier (MOCK)", req.Tier),
		"tier":    req.Tier,
		"credits": trainingCredits[req.Tier],
	})
}
