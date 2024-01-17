package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"text/template"

	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHTML)
	mux.HandleFunc("/upload", handleUpload)
	handler := cors.Default().Handler(mux)
	fmt.Println("Server running on port 8082")
	http.ListenAndServe(":8082", handler)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	// Đọc tệp HTML
	html, err := template.ParseFiles("./index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading HTML file: %v", err), http.StatusInternalServerError)
		return
	}

	// Phục vụ tệp HTML
	html.Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving the file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()
	fmt.Println("upload")
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	fileWriter, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating form file: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error copying file content: %v", err), http.StatusInternalServerError)
		return
	}
	writer.Close()

	apiURL := "http://127.0.0.1:8502/predict/"
	response, err := http.Post(apiURL, writer.FormDataContentType(), &buffer)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error making request to prediction server: %v", err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		var latexCode string
		json.NewDecoder(response.Body).Decode(&latexCode)

		fmt.Fprintf(w, "%s", latexCode)
	} else {
		http.Error(w, fmt.Sprintf("Error from prediction server: %s", response.Status), http.StatusInternalServerError)
	}
}
