package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"node/database"
	"node/tcpserver"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	uploadType := r.FormValue("uploadType")

	if uploadType == "single" {
		UploadToSingleNodeHandler(w, r)
	} else if uploadType == "all" {
		UploadToAllNodeHandler(w, r)
	} else {
		http.Error(w, "Invalid upload type", http.StatusBadRequest)
		return
	}
}

func UploadToSingleNodeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("singleupload 호출")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads/"
	os.MkdirAll(uploadDir, os.ModePerm)
	filePath := uploadDir + header.Filename
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	_, err = database.SaveFileToDB(header.Filename, filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func UploadToAllNodeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("allupload 호출")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads/"
	os.MkdirAll(uploadDir, os.ModePerm)
	filePath := uploadDir + header.Filename
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	fileInfo, err := database.SaveFileToDB(header.Filename, filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 다른 노드들에게 파일 전송

	tcpserver.SendFileToOtherNodes(file, fileInfo)

	http.Redirect(w, r, "/", http.StatusSeeOther)

	fmt.Println("전체 업로드 완료")
}
