package repository

import ( "context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"server/internal/models"
)

// GetModelsByUserID retrieves all models for a specific user
func GetModelsByUserID(ctx context.Context, userID int) ([]map[string]interface{}, error) {
	if models.Pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, user_id, name, picture, folder, training_script, created_at, updated_at
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
			row[string(fieldDescriptions[i].Name)] = v
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
		SELECT id, user_id, name, picture, folder, training_script, created_at, updated_at
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
			row[string(fieldDescriptions[i].Name)] = v
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

	query := `SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1`

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

// InsertUser inserts a new user
func InsertUser(ctx context.Context, email, password string) (int, error) {
	if models.Pool == nil {
		return 0, fmt.Errorf("database connection not initialized")
	}

	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int
	err := models.Pool.QueryRow(ctx, query, email, password).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	log.Printf("Inserted user with ID: %d", id)
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
