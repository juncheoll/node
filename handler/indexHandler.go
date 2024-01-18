package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"node/database"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	files, err := database.GetFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filesJSON, err := json.Marshal(files)
	if err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 파일 목록 전달
	w.Header().Set("Content_Type", "application/json")
	tmpl.Execute(w, string(filesJSON))
}
