package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", handleUpload)
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)
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
