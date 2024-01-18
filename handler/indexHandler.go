package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"

	"node/database"
	"node/tcpserver"
)

type Message struct {
	NodeName  string          `json:"nodename"`
	NodeCount int             `json:"nodecnt"`
	Files     []database.File `json:"files"`
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	files, err := database.GetFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message := Message{
		NodeName:  os.Getenv("HOSTNAME"),
		NodeCount: len(tcpserver.Nodes),
		Files:     files,
	}

	messageJSON, err := json.Marshal(message)
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
	tmpl.Execute(w, string(messageJSON))
}
