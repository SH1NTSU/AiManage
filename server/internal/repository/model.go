package repository

import ( "context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
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

	query := `SELECT id, email, password, username, api_key, created_at, updated_at FROM users WHERE email = $1`

	log.Printf("[DB] GetUserByEmail - Querying for email: %s", email)
	log.Printf("[DB] Query: %s", query)

	rows, err := models.Pool.Query(ctx, query, email)
	if err != nil {
		log.Printf("[DB ERROR] Query failed for email %s: %v", email, err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		log.Printf("[DB] No user found for email: %s", email)
		return nil, nil // User not found
	}

	log.Printf("[DB] User found for email: %s", email)

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

// IncrementModelViews increments the view count for a published model
func IncrementModelViews(ctx context.Context, modelID int) error {
	if models.Pool == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
		UPDATE published_models
		SET views_count = views_count + 1
		WHERE id = $1
	`

	_, err := models.Pool.Exec(ctx, query, modelID)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	log.Printf("Incremented views for published model ID: %d", modelID)
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

// GetUserByUsername retrieves a user by username
func GetUserByApiKey(ctx context.Context, apiKey string) (*map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `SELECT id, email, username, api_key, subscription_tier, subscription_status, training_credits FROM users WHERE api_key = $1`

	log.Printf("[DB] GetUserByApiKey - Querying for API key")

	rows, err := models.Pool.Query(ctx, query, apiKey)
	if err != nil {
		log.Printf("[DB ERROR] Query failed for API key: %v", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		log.Printf("[DB] No user found with provided API key")
		return nil, nil
	}

	var user map[string]interface{} = make(map[string]interface{})
	var id int
	var email, username, apiKeyDb string
	var subscriptionTier, subscriptionStatus *string
	var trainingCredits *int

	if err := rows.Scan(&id, &email, &username, &apiKeyDb, &subscriptionTier, &subscriptionStatus, &trainingCredits); err != nil {
		log.Printf("[DB ERROR] Failed to scan row: %v", err)
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

	log.Printf("[DB] User found with API key: %s", email)
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

	query := `
		INSERT INTO users (email, password, username)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int
	err := models.Pool.QueryRow(ctx, query, email, password, username).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	log.Printf("Inserted user with ID: %d (username: %s)", id, username)
	return id, nil
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
