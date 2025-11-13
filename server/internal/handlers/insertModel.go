package handlers

import (
	"io"
	"log"
	"net/http"
	"os"

	"server/helpers"
	"server/internal/middlewares"
	"server/internal/repository"
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

	// Debug: Log all form fields
	log.Println("📋 Form fields received:")
	for key, values := range r.Form {
		log.Printf("  - %s: %v", key, values)
	}

	// Debug: Log all files
	log.Println("📁 Files received:")
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for key, files := range r.MultipartForm.File {
			log.Printf("  - %s: %d file(s)", key, len(files))
			for i, file := range files {
				log.Printf("    [%d] %s (%d bytes)", i, file.Filename, file.Size)
			}
		}
	} else {
		log.Println("  - No files in multipart form")
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}
	log.Println("📄 Received model name:", name)

	// Check if this is local mode or server mode
	folderPath := r.FormValue("folder_path")
	isLocalMode := folderPath != ""

	log.Printf("📍 Mode: %s", map[bool]string{true: "Local", false: "Server"}[isLocalMode])

	var modelDir string
	if isLocalMode {
		// Local mode: use the provided path
		modelDir = folderPath
		log.Printf("📂 Using local folder path: %s", modelDir)
	} else {
		// Server mode: create uploads directory
		modelDir = "./uploads/" + name
		if err := os.MkdirAll(modelDir, os.ModePerm); err != nil {
			log.Println("❌ Failed to create model directory:", err)
			http.Error(w, "Could not create model directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("📁 Created server directory: %s", modelDir)
	}

	// Handle picture upload (optional)
	var picturePath string
	pictureFile, pictureHeader, err := r.FormFile("picture")
	if err == nil {
		defer pictureFile.Close()

		picturePath = modelDir + "/" + pictureHeader.Filename
		pictureOut, err := os.Create(picturePath)
		if err != nil {
			log.Println("❌ Could not create picture file:", err)
			http.Error(w, "Could not save picture: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer pictureOut.Close()

		if _, err := io.Copy(pictureOut, pictureFile); err != nil {
			log.Println("❌ Could not write picture file:", err)
			http.Error(w, "Could not save picture: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("✅ Picture saved:", picturePath)
	} else {
		log.Println("ℹ️ No picture provided (optional)")
	}

	// Handle folder/model zip upload (only for server mode)
	if !isLocalMode {
		zipFile, zipHeader, err := r.FormFile("folder")
		if err != nil {
			log.Println("❌ No model zip file provided:", err)
			log.Println("💡 Expected file field name: 'folder'")

			// Suggest available field names
			if r.MultipartForm != nil && r.MultipartForm.File != nil && len(r.MultipartForm.File) > 0 {
				log.Println("💡 Available file fields:")
				for key := range r.MultipartForm.File {
					log.Printf("   - '%s'", key)
				}
			}

			http.Error(w, "You must provide a model zip file with field name 'folder' for server mode", http.StatusBadRequest)
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
		log.Println("✅ Model zip saved:", zipPath)

		// Extract zip
		if err := helpers.Unzip(zipPath, modelDir); err != nil {
			log.Println("❌ Could not unzip file:", err)
			http.Error(w, "Could not unzip model: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("✅ Model unzipped to:", modelDir)

		// Optional: remove the zip after extraction
		os.Remove(zipPath)
	} else {
		log.Println("ℹ️ Local mode: Skipping file upload, using local path")
	}

	// Get user email from context
	email, ok := r.Context().Value(middlewares.UserEmailKey).(string)
	if !ok || email == "" {
		log.Println("❌ User email not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := repository.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Println("❌ Failed to get user:", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Println("❌ User not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userID, ok := (*user)["id"].(int32)
	if !ok {
		log.Println("❌ Failed to get user ID")
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Get training script path (optional, defaults to "train.py")
	trainingScript := r.FormValue("training_script")
	if trainingScript == "" {
		trainingScript = "train.py"
		log.Println("ℹ️ No training_script specified, defaulting to 'train.py'")
	} else {
		log.Printf("📜 Training script: %s", trainingScript)
	}

	// Insert model into database
	log.Printf("📦 Inserting into PostgreSQL for user %d: name=%s, picture=%s, training_script=%s\n", userID, name, picturePath, trainingScript)
	modelID, err := repository.InsertModel(r.Context(), int(userID), name, picturePath, []string{modelDir}, trainingScript)
	if err != nil {
		log.Println("❌ PostgreSQL insert failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Insert successful! Model ID: %d", modelID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Model added successfully!"))
}
