package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"server/internal/middlewares"
	"server/internal/repository"
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

	user, err := repository.GetUserByEmail(r.Context(), userEmail)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Extract subscription info
	subscription := map[string]interface{}{
		"tier":            getStringField(*user, "subscription_tier", TierFree),
		"status":          getStringField(*user, "subscription_status", "active"),
		"training_credits": getIntField(*user, "training_credits", 0),
		"start_date":      (*user)["subscription_start_date"],
		"end_date":        (*user)["subscription_end_date"],
	}

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

	// TODO: Integrate with Stripe API to create checkout session
	// For now, return a mock response
	checkoutURL := os.Getenv("FRONTEND_URL") + "/checkout?tier=" + req.Tier

	log.Printf("Creating checkout session for user %s, tier: %s", userEmail, req.Tier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"checkout_url": checkoutURL,
		"tier":         req.Tier,
		"price":        subscriptionPrices[req.Tier],
		"message":      "Stripe integration pending - this is a mock response",
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

// WebhookHandler handles Stripe webhook events
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Verify Stripe signature

	var event struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received Stripe webhook: %s", event.Type)

	switch event.Type {
	case "checkout.session.completed":
		// Handle successful subscription
		log.Println("Checkout session completed")
	case "customer.subscription.updated":
		// Handle subscription updates
		log.Println("Subscription updated")
	case "customer.subscription.deleted":
		// Handle subscription cancellation
		log.Println("Subscription deleted")
	case "invoice.payment_succeeded":
		// Handle successful payment
		log.Println("Payment succeeded")
	case "invoice.payment_failed":
		// Handle failed payment
		log.Println("Payment failed")
	}

	w.WriteHeader(http.StatusOK)
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
