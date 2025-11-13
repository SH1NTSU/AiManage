package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"server/internal/models"
)

// UpdateUserStripeCustomer updates the Stripe customer ID for a user
func UpdateUserStripeCustomer(ctx context.Context, userEmail, stripeCustomerID string) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		UPDATE users
		SET stripe_customer_id = $1, updated_at = $2
		WHERE email = $3
	`

	_, err := models.Pool.Exec(ctx, query, stripeCustomerID, time.Now(), userEmail)
	if err != nil {
		return fmt.Errorf("failed to update stripe customer ID: %w", err)
	}

	log.Printf("✅ Updated Stripe customer ID for user: %s", userEmail)
	return nil
}

// UpdateUserSubscription updates user subscription details
func UpdateUserSubscription(ctx context.Context, userEmail string, fields map[string]interface{}) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// Build dynamic UPDATE query
	query := "UPDATE users SET updated_at = $1"
	args := []interface{}{time.Now()}
	argIndex := 2

	for field, value := range fields {
		query += fmt.Sprintf(", %s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE email = $%d", argIndex)
	args = append(args, userEmail)

	_, err := models.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user subscription: %w", err)
	}

	log.Printf("✅ Updated subscription for user: %s", userEmail)
	return nil
}

// UpdateUserSubscriptionStatus updates only the subscription status
func UpdateUserSubscriptionStatus(ctx context.Context, userEmail, status string) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		UPDATE users
		SET subscription_status = $1, updated_at = $2
		WHERE email = $3
	`

	_, err := models.Pool.Exec(ctx, query, status, time.Now(), userEmail)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	log.Printf("✅ Updated subscription status for user %s: %s", userEmail, status)
	return nil
}

// GetUserEmailByStripeCustomer retrieves user email by Stripe customer ID
func GetUserEmailByStripeCustomer(ctx context.Context, stripeCustomerID string) (string, error) {
	if models.Pool == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT email
		FROM users
		WHERE stripe_customer_id = $1
	`

	var email string
	err := models.Pool.QueryRow(ctx, query, stripeCustomerID).Scan(&email)
	if err != nil {
		return "", fmt.Errorf("failed to get user by stripe customer ID: %w", err)
	}

	return email, nil
}

// DecrementUserTrainingCredits decrements training credits for a user
func DecrementUserTrainingCredits(ctx context.Context, userEmail string) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		UPDATE users
		SET training_credits = GREATEST(training_credits - 1, 0), updated_at = $1
		WHERE email = $2 AND training_credits > 0
	`

	result, err := models.Pool.Exec(ctx, query, time.Now(), userEmail)
	if err != nil {
		return fmt.Errorf("failed to decrement training credits: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no credits to decrement or user not found")
	}

	log.Printf("✅ Decremented training credits for user: %s", userEmail)
	return nil
}

// ResetMonthlyCreditsForAllUsers resets training credits for all users based on their tier
func ResetMonthlyCreditsForAllUsers(ctx context.Context) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		UPDATE users
		SET training_credits = CASE
			WHEN subscription_tier = 'basic' THEN 10
			WHEN subscription_tier = 'pro' THEN 50
			WHEN subscription_tier = 'enterprise' THEN 999
			ELSE 0
		END,
		updated_at = $1
		WHERE subscription_tier != 'free'
	`

	result, err := models.Pool.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to reset monthly credits: %w", err)
	}

	rowsAffected := result.RowsAffected()
	log.Printf("✅ Reset monthly credits for %d users", rowsAffected)
	return nil
}
