package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"node/database"

	"github.com/gorilla/mux"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["fileName"]

	file, err := database.GetFileByName(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if file == nil {
		http.Error(w, "File not found", http.StatusInternalServerError)
		return
	}

	f, err := os.Open(file.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename+%s", file.Name))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	_, err = io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("File '%s' downloaded\n", file.Name)
}
