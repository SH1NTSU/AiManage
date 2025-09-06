package handlers

import (
	"encoding/json"
	"net/http"
	"server/internal/models"
	"server/internal/types"
)







func InsertHandler(w http.ResponseWriter, r *http.Request) {
	var m types.Model

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} 


	err = models.Insert(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}


	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Model added succesfully!"))
	
}
