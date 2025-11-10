package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	godotenv.Load()

	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		log.Fatal("DB_URI not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	rows, err := pool.Query(context.Background(), "SELECT id, email, username, created_at FROM users ORDER BY id")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n=== All Users in Database ===")
	fmt.Printf("%-5s %-30s %-20s %-25s\n", "ID", "Email", "Username", "Created At")
	fmt.Println("------------------------------------------------------------------------------------")

	for rows.Next() {
		var id int
		var email string
		var username *string
		var createdAt interface{}

		if err := rows.Scan(&id, &email, &username, &createdAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		usernameStr := "NULL"
		if username != nil {
			usernameStr = *username
		}

		fmt.Printf("%-5d %-30s %-20s %-25v\n", id, email, usernameStr, createdAt)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Rows iteration error: %v", err)
	}

	fmt.Println("\n")
}
