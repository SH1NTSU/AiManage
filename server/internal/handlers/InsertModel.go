package handlers

import (
	"io"
	"net/http"
	"os"
	"server/internal/models"
	"server/internal/types"
)







func InsertHandler(w http.ResponseWriter, r *http.Request) {

    const CollectionName = "Models"
    // w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
    // w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    // w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    //
    // Handle preflight (OPTIONS request)
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }
    err := r.ParseMultipartForm(100 << 20) // 10 MB
    	if err != nil {
        	http.Error(w, "Could not parse multipart form: "+err.Error(), http.StatusBadRequest)
        	return
	}

    name := r.FormValue("name")

    pictureFile, pictureHeader, err := r.FormFile("picture")
    var picturePath string
    if err == nil {
        defer pictureFile.Close()

        picturePath = "./uploads/" + pictureHeader.Filename
        out, err := os.Create(picturePath)
        if err != nil {
            http.Error(w, "Could not save picture: "+err.Error(), http.StatusInternalServerError)
            return
        }
        defer out.Close()
        io.Copy(out, pictureFile)
    }

    // Handle multiple folder files
    folderFiles := r.MultipartForm.File["folder"]
    folderPaths := []string{}
    for _, fileHeader := range folderFiles {
        file, err := fileHeader.Open()
        if err != nil {
            http.Error(w, "Could not open folder file: "+err.Error(), http.StatusInternalServerError)
            return
        }
        defer file.Close()

        path := "./uploads/" + fileHeader.Filename
        out, err := os.Create(path)
        if err != nil {
            http.Error(w, "Could not save folder file: "+err.Error(), http.StatusInternalServerError)
            return
        }
        defer out.Close()
        io.Copy(out, file)

        folderPaths = append(folderPaths, path)
    }

    m := types.Model{
        Name:    name,
        Picture: picturePath,
        Folder:  folderPaths,
    }

    err = models.Insert( CollectionName, m)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Model added successfully!"))
}
