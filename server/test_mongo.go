package main
//
// import (
// 	"fmt"
// 	"log"
// 	"server/internal/models"
// )
//
// func main() {
// 	fmt.Println("Testing MongoDB connection...")
// 	
// 	err := models.ConnectDB()
// 	if err != nil {
// 		log.Fatal("❌ Connection failed:", err)
// 	}
// 	
// 	fmt.Println("✅ MongoDB connection successful!")
// 	
// 	// Test basic operation
// 	collection := models.GetCollection()
// 	if collection == nil {
// 		log.Fatal("❌ Failed to get collection")
// 	}
// 	
// 	fmt.Println("✅ Collection access successful!")
// }
