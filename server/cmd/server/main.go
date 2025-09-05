package main

import (
	
	"log"
	"net/http"
	"server/internal/service"
)

func  main() {
 	r := service.NewRouter()

	port := ":8080"
	log.Println("Server running on port", port)
	if err := http.ListenAndServe(port, r); err != nil {
			log.Fatalf("Server failed: %v", err)
			}
		}

