package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"server/helpers"
	"server/internal/models"
)

// GetModelsByUserID retrieves all models for a specific user
func GetModelsByUserID(ctx context.Context, userID int) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, name, picture, folder, training_script, trained_model_path, trained_at, accuracy_score, created_at, updated_at
		FROM models
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := models.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fieldDescriptions := rows.FieldDescriptions()
		row := make(map[string]interface{})
		for i, v := range values {
			fieldName := string(fieldDescriptions[i].Name)
			
			// Convert accuracy_score to float64 if it exists
			if fieldName == "accuracy_score" && v != nil {
				var acc float64
				switch val := v.(type) {
				case float64:
					acc = val
				case float32:
					acc = float64(val)
				case int64:
					acc = float64(val)
				case int32:
					acc = float64(val)
				case int:
					acc = float64(val)
				case string:
					// Try to parse string to float64
					if parsed, err := strconv.ParseFloat(val, 64); err == nil {
						acc = parsed
					} else {
						row[fieldName] = nil
						continue
					}
				default:
					// Try to convert via fmt.Sprintf and parse
					if str := fmt.Sprintf("%v", val); str != "" && str != "<nil>" {
						if parsed, err := strconv.ParseFloat(str, 64); err == nil {
							acc = parsed
						} else {
							row[fieldName] = nil
							continue
						}
					} else {
						row[fieldName] = nil
						continue
					}
				}
				row[fieldName] = acc
			} else {
				row[fieldName] = v
			}

			// Convert picture path from "./uploads/..." to "/uploads/..."
			if fieldName == "picture" && v != nil {
				if picturePath, ok := v.(string); ok && picturePath != "" {
					row[fieldName] = strings.TrimPrefix(picturePath, ".")
				}
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d models for user %d", len(results), userID)
	return results, nil
}

// GetAllModels retrieves all models from the database
func GetAllModels(ctx context.Context) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, name, picture, folder, training_script, trained_model_path, trained_at, created_at, updated_at
		FROM models
		ORDER BY created_at DESC
	`

	rows, err := models.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fieldDescriptions := rows.FieldDescriptions()
		row := make(map[string]interface{})
		for i, v := range values {
			fieldName := string(fieldDescriptions[i].Name)
			row[fieldName] = v

			// Convert picture path from "./uploads/..." to "/uploads/..."
			if fieldName == "picture" && v != nil {
				if picturePath, ok := v.(string); ok && picturePath != "" {
					row[fieldName] = strings.TrimPrefix(picturePath, ".")
				}
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d models", len(results))
	return results, nil
}

// InsertModel inserts a new model into the database
func InsertModel(ctx context.Context, userID int, name, picture string, folder []string, trainingScript string) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	// Use default if training_script is empty
	if trainingScript == "" {
		trainingScript = "train.py"
	}

	query := `
		INSERT INTO models (user_id, name, picture, folder, training_script)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int
	err := models.Pool.QueryRow(ctx, query, userID, name, picture, folder, trainingScript).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	log.Printf("Inserted model with ID: %d (training_script: %s)", id, trainingScript)
	return id, nil
}

// Query executes a generic SELECT query and returns results as maps
func Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	rows, err := models.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fieldDescriptions := rows.FieldDescriptions()
		row := make(map[string]interface{})
		for i, v := range values {
			row[string(fieldDescriptions[i].Name)] = v
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// QueryRow executes a query that returns a single row
func QueryRow(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	rows, err := models.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return row, nil
}

// Exec executes a query without returning rows (INSERT, UPDATE, DELETE)
func Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	result, err := models.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("exec failed: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(ctx context.Context, email string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, password, username, api_key, created_at, updated_at,
		subscription_tier, subscription_status, training_credits,
		stripe_customer_id, stripe_subscription_id, subscription_start_date, subscription_end_date,
		email_verified, verification_token, verification_token_expires_at
		FROM users WHERE email = $1`

	rows, err := models.Pool.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // User not found
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}

// DeleteModel deletes a model by ID and userID (for security)
func DeleteModel(ctx context.Context, modelID int, userID int) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	// Security: Make sure the model belongs to this user
	query := `
		DELETE FROM models
		WHERE id = $1 AND user_id = $2
		RETURNING id
	`

	var id int
	err := models.Pool.QueryRow(ctx, query, modelID, userID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("model not found or you don't have permission to delete it")
		}
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	log.Printf("Deleted model ID: %d for user: %d", id, userID)
	return id, nil
}

// UpdateTrainedModelPath updates the trained_model_path for a specific model
func UpdateTrainedModelPath(ctx context.Context, modelName string, modelPath string) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE models
		SET trained_model_path = $1, trained_at = NOW()
		WHERE name = $2
	`

	result, err := models.Pool.Exec(ctx, query, modelPath, modelName)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Warning: No model found with name '%s' to update trained path", modelName)
	} else {
		log.Printf("Updated trained_model_path for model '%s' to '%s'", modelName, modelPath)
	}

	return nil
}

// UpdateModelAccuracy updates the accuracy_score for a specific model
// accuracy parameter should be in percentage format (e.g., 95.50 for 95.5%)
func UpdateModelAccuracy(ctx context.Context, modelName string, accuracy float64) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE models
		SET accuracy_score = $1
		WHERE name = $2
	`

	result, err := models.Pool.Exec(ctx, query, accuracy, modelName)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️  Warning: No model found with name '%s' to update accuracy", modelName)
	} else {
		log.Printf("✅ Updated accuracy_score for model '%s' to %.2f%%", modelName, accuracy)
	}

	return nil
}

// UpdateTrainedModelPathAndAccuracy updates both trained_model_path and accuracy_score for a specific model
// accuracy parameter should be in percentage format (e.g., 95.50 for 95.5%)
func UpdateTrainedModelPathAndAccuracy(ctx context.Context, modelName string, modelPath string, accuracy *float64) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	var query string
	var err error
	var result interface{}

	if accuracy != nil {
		query = `
			UPDATE models
			SET trained_model_path = $1, trained_at = NOW(), accuracy_score = $2
			WHERE name = $3
		`
		result, err = models.Pool.Exec(ctx, query, modelPath, *accuracy, modelName)
	} else {
		query = `
			UPDATE models
			SET trained_model_path = $1, trained_at = NOW()
			WHERE name = $2
		`
		result, err = models.Pool.Exec(ctx, query, modelPath, modelName)
	}

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Extract rows affected from result (pgx v5 returns CommandTag)
	var rowsAffected int64
	if tag, ok := result.(interface{ RowsAffected() int64 }); ok {
		rowsAffected = tag.RowsAffected()
	}

	if rowsAffected == 0 {
		log.Printf("⚠️  Warning: No model found with name '%s' to update", modelName)
	} else {
		if accuracy != nil {
			log.Printf("✅ Updated trained_model_path and accuracy_score for model '%s' (accuracy: %.2f%%)", modelName, *accuracy)
		} else {
			log.Printf("✅ Updated trained_model_path for model '%s'", modelName)
		}
	}

	return nil
}

// GetModelByFolderPath retrieves a model by its folder path
func GetModelByFolderPath(ctx context.Context, folderPath string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, name, picture, folder, training_script, trained_model_path, trained_at, accuracy_score, created_at, updated_at
		FROM models
		WHERE $1 = ANY(folder)
		LIMIT 1
	`

	rows, err := models.Pool.Query(ctx, query, folderPath)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var id, userID int32
		var name, picture, trainingScript string
		var folder []string
		var trainedModelPath, accuracyScore *string
		var trainedAt, createdAt, updatedAt *time.Time

		err := rows.Scan(&id, &userID, &name, &picture, &folder, &trainingScript, &trainedModelPath, &trainedAt, &accuracyScore, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		result := make(map[string]interface{})
		result["id"] = id
		result["user_id"] = userID
		result["name"] = name
		result["picture"] = picture
		result["folder"] = folder
		result["training_script"] = trainingScript
		if trainedModelPath != nil {
			result["trained_model_path"] = *trainedModelPath
		}
		if trainedAt != nil {
			result["trained_at"] = *trainedAt
		}
		if accuracyScore != nil {
			result["accuracy_score"] = *accuracyScore
		}
		if createdAt != nil {
			result["created_at"] = *createdAt
		}
		if updatedAt != nil {
			result["updated_at"] = *updatedAt
		}

		return &result, nil
	}

	return nil, fmt.Errorf("no model found with folder path: %s", folderPath)
}

// GetModelByName retrieves a model by its name (useful for training completion)
func GetModelByName(ctx context.Context, name string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, name, picture, folder, training_script, trained_model_path, trained_at, accuracy_score, created_at, updated_at
		FROM models
		WHERE name = $1
		LIMIT 1
	`

	rows, err := models.Pool.Query(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}

// GetModelByID retrieves a model by its ID
func GetModelByID(ctx context.Context, modelID int) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, name, picture, folder, training_script, trained_model_path, trained_at, accuracy_score, created_at, updated_at
		FROM models
		WHERE id = $1
		LIMIT 1
	`

	rows, err := models.Pool.Query(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}

// InsertPublishedModel inserts a new published model into the marketplace
func InsertPublishedModel(ctx context.Context, req map[string]interface{}) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO published_models (
			model_id, publisher_id, name, picture, trained_model_path, training_script,
			description, price, license_type, category, tags, model_type, framework, accuracy_score
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	var id int
	err := models.Pool.QueryRow(ctx, query,
		req["model_id"],
		req["publisher_id"],
		req["name"],
		req["picture"],
		req["trained_model_path"],
		req["training_script"],
		req["description"],
		req["price"],
		req["license_type"],
		req["category"],
		req["tags"],
		req["model_type"],
		req["framework"],
		req["accuracy_score"],
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("insert published model failed: %w", err)
	}

	log.Printf("Published model with ID: %d", id)
	return id, nil
}

// GetPublishedModels retrieves all active published models for community marketplace
func GetPublishedModels(ctx context.Context) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT
			pm.id, pm.model_id, pm.publisher_id, pm.name, pm.picture, pm.trained_model_path, pm.training_script,
			pm.description, pm.short_description, pm.price, pm.category, pm.tags, pm.model_type, pm.framework,
			pm.file_size, pm.accuracy_score, pm.license_type, pm.downloads_count, pm.views_count,
			pm.rating_average, pm.rating_count, pm.is_active, pm.is_featured, pm.published_at, pm.updated_at,
			u.username as publisher_username
		FROM published_models pm
		LEFT JOIN users u ON pm.publisher_id = u.id
		WHERE pm.is_active = true
		ORDER BY pm.published_at DESC
	`

	rows, err := models.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fieldDescriptions := rows.FieldDescriptions()
		row := make(map[string]interface{})
		for i, v := range values {
			fieldName := string(fieldDescriptions[i].Name)
			row[fieldName] = v

			// Convert picture path from "./uploads/..." to "/uploads/..."
			if fieldName == "picture" && v != nil {
				if picturePath, ok := v.(string); ok && picturePath != "" {
					row[fieldName] = strings.TrimPrefix(picturePath, ".")
				}
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d published models", len(results))
	return results, nil
}

// GetPublishedModelByID retrieves a single published model by ID
func GetPublishedModelByID(ctx context.Context, modelID int) (map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT
			pm.id, pm.model_id, pm.publisher_id, pm.name, pm.picture, pm.trained_model_path, pm.training_script,
			pm.description, pm.short_description, pm.price, pm.category, pm.tags, pm.model_type, pm.framework,
			pm.file_size, pm.accuracy_score, pm.license_type, pm.downloads_count, pm.views_count,
			pm.rating_average, pm.rating_count, pm.is_active, pm.is_featured, pm.published_at, pm.updated_at,
			u.username as publisher_username
		FROM published_models pm
		LEFT JOIN users u ON pm.publisher_id = u.id
		WHERE pm.id = $1
		LIMIT 1
	`

	rows, err := models.Pool.Query(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		fieldName := string(fieldDescriptions[i].Name)
		row[fieldName] = v

		// Convert picture path from "./uploads/..." to "/uploads/..."
		if fieldName == "picture" && v != nil {
			if picturePath, ok := v.(string); ok && picturePath != "" {
				row[fieldName] = strings.TrimPrefix(picturePath, ".")
			}
		}
	}

	log.Printf("Retrieved published model ID: %d", modelID)
	return row, nil
}

// IncrementModelViews increments the view count for a published model (one view per user)
// userID can be nil for anonymous users, ipAddress is used as fallback
func IncrementModelViews(ctx context.Context, modelID int, userID *int, ipAddress string) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Start a transaction to ensure atomicity
	tx, err := models.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Try to insert a new view record
	// If it already exists (user already viewed), it will fail silently
	var insertQuery string
	var args []interface{}

	if userID != nil {
		// Authenticated user - track by user_id
		insertQuery = `
			INSERT INTO model_views (model_id, user_id, ip_address)
			VALUES ($1, $2, $3)
			ON CONFLICT (model_id, user_id) DO NOTHING
		`
		args = []interface{}{modelID, *userID, ipAddress}
	} else {
		// Anonymous user - track by IP address
		insertQuery = `
			INSERT INTO model_views (model_id, user_id, ip_address)
			VALUES ($1, NULL, $2)
			ON CONFLICT (model_id, ip_address) WHERE user_id IS NULL DO NOTHING
		`
		args = []interface{}{modelID, ipAddress}
	}

	result, err := tx.Exec(ctx, insertQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}

	// Check if a new row was inserted
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// User already viewed this model, don't increment counter
		log.Printf("User already viewed model ID: %d (skipping increment)", modelID)
		return nil
	}

	// New view - increment the counter
	updateQuery := `
		UPDATE published_models
		SET views_count = views_count + 1
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, updateQuery, modelID)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Incremented views for published model ID: %d (new view)", modelID)
	return nil
}

// IncrementModelDownloads increments the download count for a published model
func IncrementModelDownloads(ctx context.Context, modelID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE published_models
		SET downloads_count = downloads_count + 1
		WHERE id = $1
	`

	_, err := models.Pool.Exec(ctx, query, modelID)
	if err != nil {
		return fmt.Errorf("failed to increment downloads: %w", err)
	}

	log.Printf("Incremented downloads for published model ID: %d", modelID)
	return nil
}

// RecordModelDownload records a download in the model_purchases table for history
func RecordModelDownload(ctx context.Context, userID int, modelID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Check if this user has already downloaded this model
	checkQuery := `
		SELECT id FROM model_purchases
		WHERE user_id = $1 AND published_model_id = $2
		LIMIT 1
	`

	rows, err := models.Pool.Query(ctx, checkQuery, userID, modelID)
	if err != nil {
		return fmt.Errorf("failed to check existing download: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		// Already downloaded, don't record again
		log.Printf("User %d already downloaded model %d", userID, modelID)
		return nil
	}

	// Record new download
	insertQuery := `
		INSERT INTO model_purchases (user_id, published_model_id, purchase_type, amount_paid, purchased_at)
		VALUES ($1, $2, 'download', 0, NOW())
	`

	_, err = models.Pool.Exec(ctx, insertQuery, userID, modelID)
	if err != nil {
		return fmt.Errorf("failed to record download: %w", err)
	}

	log.Printf("Recorded download for user %d, model %d", userID, modelID)
	return nil
}

// ======= LIKES =======

// LikeModel adds a like to a published model
func LikeModel(ctx context.Context, userID int, modelID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO model_likes (user_id, published_model_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, published_model_id) DO NOTHING
	`

	_, err := models.Pool.Exec(ctx, query, userID, modelID)
	if err != nil {
		return fmt.Errorf("failed to like model: %w", err)
	}

	log.Printf("User %d liked model %d", userID, modelID)
	return nil
}

// UnlikeModel removes a like from a published model
func UnlikeModel(ctx context.Context, userID int, modelID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		DELETE FROM model_likes
		WHERE user_id = $1 AND published_model_id = $2
	`

	result, err := models.Pool.Exec(ctx, query, userID, modelID)
	if err != nil {
		return fmt.Errorf("failed to unlike model: %w", err)
	}

	rowsAffected := result.RowsAffected()
	log.Printf("User %d unliked model %d (rows affected: %d)", userID, modelID, rowsAffected)
	return nil
}

// GetModelLikesCount gets the total number of likes for a model
func GetModelLikesCount(ctx context.Context, modelID int) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT COUNT(*) FROM model_likes WHERE published_model_id = $1`

	var count int
	err := models.Pool.QueryRow(ctx, query, modelID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get likes count: %w", err)
	}

	return count, nil
}

// HasUserLikedModel checks if a user has liked a specific model
func HasUserLikedModel(ctx context.Context, userID int, modelID int) (bool, error) {
	if models.Pool == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM model_likes
			WHERE user_id = $1 AND published_model_id = $2
		)
	`

	var exists bool
	err := models.Pool.QueryRow(ctx, query, userID, modelID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user liked model: %w", err)
	}

	return exists, nil
}

// ======= COMMENTS =======

// AddComment adds a comment to a published model
func AddComment(ctx context.Context, userID int, modelID int, commentText string, parentCommentID *int) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO model_comments (user_id, published_model_id, comment_text, parent_comment_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var commentID int
	err := models.Pool.QueryRow(ctx, query, userID, modelID, commentText, parentCommentID).Scan(&commentID)
	if err != nil {
		return 0, fmt.Errorf("failed to add comment: %w", err)
	}

	log.Printf("User %d added comment %d to model %d", userID, commentID, modelID)
	return commentID, nil
}

// GetModelComments retrieves all comments for a model (with user info)
func GetModelComments(ctx context.Context, modelID int) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT
			c.id, c.user_id, c.published_model_id, c.parent_comment_id,
			c.comment_text, c.edited, c.created_at, c.updated_at,
			u.username, u.email
		FROM model_comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.published_model_id = $1
		ORDER BY c.created_at ASC
	`

	rows, err := models.Pool.Query(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fieldDescriptions := rows.FieldDescriptions()
		row := make(map[string]interface{})
		for i, v := range values {
			row[string(fieldDescriptions[i].Name)] = v
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d comments for model %d", len(results), modelID)
	return results, nil
}

// DeleteComment deletes a comment (only by the comment author)
func DeleteComment(ctx context.Context, commentID int, userID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Security: ensure the comment belongs to this user
	query := `
		DELETE FROM model_comments
		WHERE id = $1 AND user_id = $2
	`

	result, err := models.Pool.Exec(ctx, query, commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("comment not found or you don't have permission to delete it")
	}

	log.Printf("User %d deleted comment %d", userID, commentID)
	return nil
}

// GetPublishedModelsByPublisher retrieves all published models by a specific publisher
func GetPublishedModelsByPublisher(ctx context.Context, publisherID int) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT
			pm.id, pm.model_id, pm.publisher_id, pm.name, pm.picture, pm.trained_model_path, pm.training_script,
			pm.description, pm.short_description, pm.price, pm.category, pm.tags, pm.model_type, pm.framework,
			pm.file_size, pm.accuracy_score, pm.license_type, pm.downloads_count, pm.views_count,
			pm.rating_average, pm.rating_count, pm.is_active, pm.is_featured, pm.published_at, pm.updated_at,
			u.username as publisher_username
		FROM published_models pm
		LEFT JOIN users u ON pm.publisher_id = u.id
		WHERE pm.publisher_id = $1
		ORDER BY pm.published_at DESC
	`

	rows, err := models.Pool.Query(ctx, query, publisherID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fieldDescriptions := rows.FieldDescriptions()
		row := make(map[string]interface{})
		for i, v := range values {
			fieldName := string(fieldDescriptions[i].Name)
			row[fieldName] = v

			// Convert picture path from "./uploads/..." to "/uploads/..."
			if fieldName == "picture" && v != nil {
				if picturePath, ok := v.(string); ok && picturePath != "" {
					row[fieldName] = strings.TrimPrefix(picturePath, ".")
				}
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d published models for publisher %d", len(results), publisherID)
	return results, nil
}

// UnpublishModel sets is_active to false for a published model
func UnpublishModel(ctx context.Context, publishedModelID int, publisherID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE published_models
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND publisher_id = $2
	`

	result, err := models.Pool.Exec(ctx, query, publishedModelID, publisherID)
	if err != nil {
		return fmt.Errorf("failed to unpublish model: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("model not found or you don't have permission to unpublish it")
	}

	log.Printf("Model %d unpublished by publisher %d", publishedModelID, publisherID)
	return nil
}

// GetUserByUsername retrieves a user by username
func GetUserByApiKey(ctx context.Context, apiKey string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, username, api_key, subscription_tier, subscription_status, training_credits FROM users WHERE api_key = $1`

	rows, err := models.Pool.Query(ctx, query, apiKey)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var user map[string]interface{} = make(map[string]interface{})
	var id int
	var email, username, apiKeyDb string
	var subscriptionTier, subscriptionStatus *string
	var trainingCredits *int

	if err := rows.Scan(&id, &email, &username, &apiKeyDb, &subscriptionTier, &subscriptionStatus, &trainingCredits); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	user["id"] = id
	user["email"] = email
	user["username"] = username
	user["api_key"] = apiKeyDb
	if subscriptionTier != nil {
		user["subscription_tier"] = *subscriptionTier
	}
	if subscriptionStatus != nil {
		user["subscription_status"] = *subscriptionStatus
	}
	if trainingCredits != nil {
		user["training_credits"] = *trainingCredits
	}

	return &user, nil
}

func GetUserByUsername(ctx context.Context, username string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, email, password, username, created_at
		FROM users
		WHERE username = $1
		LIMIT 1
	`

	rows, err := models.Pool.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // User not found
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}

// InsertUser inserts a new user
func InsertUser(ctx context.Context, email, password, username string) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	// Generate API key for new user
	apiKey, err := helpers.GenerateAPIKey(email)
	if err != nil {
		log.Printf("⚠️  Failed to generate API key for user %s: %v", email, err)
		// Continue without API key - it can be generated later
		apiKey = ""
	}

	query := `
		INSERT INTO users (email, password, username, api_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err = models.Pool.QueryRow(ctx, query, email, password, username, apiKey).Scan(&id)
	if err != nil {
		// If insertion fails due to unique constraint on api_key, retry with a new key
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			log.Printf("⚠️  API key collision, retrying with new key...")
			apiKey, retryErr := helpers.GenerateAPIKey(email + time.Now().String())
			if retryErr == nil {
				err = models.Pool.QueryRow(ctx, query, email, password, username, apiKey).Scan(&id)
			}
		}
		if err != nil {
			return 0, fmt.Errorf("insert failed: %w", err)
		}
	}

	if apiKey != "" {
		log.Printf("Inserted user with ID: %d (username: %s, api_key: sk_live_...)", id, username)
	} else {
		log.Printf("Inserted user with ID: %d (username: %s, no API key generated)", id, username)
	}
	return id, nil
}

// RegenerateAPIKey generates and updates a user's API key
func RegenerateAPIKey(ctx context.Context, userID int) (string, error) {
	if models.Pool == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Get user email for key generation
	user, err := GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	email, ok := (*user)["email"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user email")
	}

	// Generate new API key
	apiKey, err := helpers.GenerateAPIKey(email)
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Retry logic for unique constraint violations
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		query := `UPDATE users SET api_key = $1 WHERE id = $2 RETURNING api_key`
		var updatedKey string
		err = models.Pool.QueryRow(ctx, query, apiKey, userID).Scan(&updatedKey)
		
		if err == nil {
			log.Printf("✅ Regenerated API key for user ID: %d", userID)
			return updatedKey, nil
		}

		// If unique constraint violation, generate a new key and retry
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			log.Printf("⚠️  API key collision (attempt %d/%d), generating new key...", i+1, maxRetries)
			apiKey, err = helpers.GenerateAPIKey(email + time.Now().String() + fmt.Sprintf("%d", i))
			if err != nil {
				return "", fmt.Errorf("failed to generate retry API key: %w", err)
			}
			continue
		}

		// Other errors
		return "", fmt.Errorf("failed to update API key: %w", err)
	}

	return "", fmt.Errorf("failed to regenerate API key after %d attempts", maxRetries)
}

// EnsureUserHasAPIKey ensures a user has an API key, generating one if missing
func EnsureUserHasAPIKey(ctx context.Context, userID int) (string, error) {
	if models.Pool == nil {
		return "", fmt.Errorf("database connection not initialized")
	}

	// Check if user already has an API key
	user, err := GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	existingKey, ok := (*user)["api_key"].(string)
	if ok && existingKey != "" {
		return existingKey, nil
	}

	// User doesn't have an API key, generate one
	log.Printf("⚠️  User %d doesn't have an API key, generating one...", userID)
	return RegenerateAPIKey(ctx, userID)
}

// GetUserByID retrieves a user by ID
func GetUserByID(ctx context.Context, userID int) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, username, api_key, created_at, updated_at FROM users WHERE id = $1`

	rows, err := models.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}

// InsertSession inserts a new session
func InsertSession(ctx context.Context, userID int, email, refreshToken string, expiresAt interface{}) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO sessions (user_id, email, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := models.Pool.QueryRow(ctx, query, userID, email, refreshToken, expiresAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	log.Printf("Inserted session with ID: %d", id)
	return id, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
func GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, email, refresh_token, expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1 AND expires_at > NOW()
	`

	rows, err := models.Pool.Query(ctx, query, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // Session not found or expired
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}

// SetVerificationToken sets the verification token and expiry for a user
func SetVerificationToken(ctx context.Context, email, token string, expiresAt time.Time) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE users
		SET verification_token = $1, verification_token_expires_at = $2
		WHERE email = $3
	`

	_, err := models.Pool.Exec(ctx, query, token, expiresAt, email)
	if err != nil {
		return fmt.Errorf("failed to set verification token: %w", err)
	}

	log.Printf("✅ Set verification token for user: %s", email)
	return nil
}

// VerifyEmailByToken verifies a user's email using the verification token
func VerifyEmailByToken(ctx context.Context, token string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// First, check if the token is valid and not expired
	query := `
		SELECT id, email, username, verification_token_expires_at
		FROM users
		WHERE verification_token = $1 AND verification_token_expires_at > NOW()
	`

	rows, err := models.Pool.Query(ctx, query, token)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("invalid or expired verification token")
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	// Update the user to mark email as verified and clear the token
	email, ok := row["email"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid email in user record")
	}

	updateQuery := `
		UPDATE users
		SET email_verified = true, verification_token = NULL, verification_token_expires_at = NULL
		WHERE email = $1
	`

	_, err = models.Pool.Exec(ctx, updateQuery, email)
	if err != nil {
		return nil, fmt.Errorf("failed to update email verification status: %w", err)
	}

	log.Printf("✅ Email verified for user: %s", email)
	return &row, nil
}

// GetUserByVerificationToken retrieves a user by verification token
func GetUserByVerificationToken(ctx context.Context, token string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, email, username, email_verified, verification_token_expires_at
		FROM users
		WHERE verification_token = $1
	`

	rows, err := models.Pool.Query(ctx, query, token)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // Token not found
	}

	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	fieldDescriptions := rows.FieldDescriptions()
	row := make(map[string]interface{})
	for i, v := range values {
		row[string(fieldDescriptions[i].Name)] = v
	}

	return &row, nil
}
