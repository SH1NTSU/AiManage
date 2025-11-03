// types/main.go
package types

import "time"

type User struct {
    ID        int       `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Password  string    `json:"-" db:"password"` // "-" prevents password from being exposed in JSON responses
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Session struct {
    ID           int       `json:"id" db:"id"`
    UserID       int       `json:"user_id" db:"user_id"`
    Email        string    `json:"email" db:"email"`
    RefreshToken string    `json:"refresh_token" db:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Model struct {
    ID        int       `json:"id" db:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    Name      string    `json:"name" db:"name"`
    Picture   string    `json:"picture" db:"picture"`
    Folder    []string  `json:"folder" db:"folder"` // PostgreSQL array support via pgx
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
