package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"server/internal/models"
	"server/internal/types"
	"server/helpers"
)







func InsertHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📩 InsertHandler called")
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse form
	err := r.ParseMultipartForm(200 << 20) // 200 MB for bigger zips
	if err != nil {
		log.Println("❌ ParseMultipartForm error:", err)
		http.Error(w, "Could not parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}
	log.Println("📄 Received model name:", name)

	modelDir := "./uploads/" + name
	if err := os.MkdirAll(modelDir, os.ModePerm); err != nil {
		log.Println("❌ Failed to create model directory:", err)
		http.Error(w, "Could not create model directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle zip upload
	zipFile, zipHeader, err := r.FormFile("model_zip")
	if err != nil {
		log.Println("❌ No zip file provided:", err)
		http.Error(w, "You must provide a model zip file", http.StatusBadRequest)
		return
	}
	defer zipFile.Close()

	zipPath := modelDir + "/" + zipHeader.Filename
	out, err := os.Create(zipPath)
	if err != nil {
		log.Println("❌ Could not create zip file:", err)
		http.Error(w, "Could not save zip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, zipFile); err != nil {
		log.Println("❌ Could not write zip file:", err)
		http.Error(w, "Could not save zip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract zip
	if err := helpers.Unzip(zipPath, modelDir); err != nil {
		
		log.Println("❌ Could not unzip file:", err)
		http.Error(w, "Could not unzip model: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("✅ Model unzipped to:", modelDir)

	// Optional: remove the zip after extraction
	os.Remove(zipPath)

	m := types.Model{
		Name:   name,
		Folder: []string{modelDir},
	}

	log.Printf("📦 Inserting into MongoDB: %+v\n", m)
	err = models.Insert("Models", m)
	if err != nil {
		log.Println("❌ MongoDB insert failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("✅ Insert successful!")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Model added successfully!"))
}
