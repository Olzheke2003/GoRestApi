package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	model "github.com/Olzheke2003/GoRestApi/internal/app/model"
)

// ErrorResponse описывает формат ошибки для фронтенда

func HandleArchiveInformation(w http.ResponseWriter, r *http.Request) {
	// Логирование заголовков запроса
	log.Println("Received request with headers:")
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("%s: %s\n", name, value)
		}
	}

	if r.Method != http.MethodPost {
		respondWithError(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(50 << 20) // 50MB

	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusBadRequest)
		log.Printf("Error reading file: %v\n", err)
		return
	}
	defer file.Close()

	log.Printf("Processing file: %s, Size: %d bytes\n", header.Filename, header.Size)

	if !isArchive(file) {
		respondWithError(w, "File is not a valid archive", http.StatusBadRequest)
		log.Println("Uploaded file is not a valid archive")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		respondWithError(w, fmt.Sprintf("Failed to seek file: %v", err), http.StatusInternalServerError)
		log.Printf("Error seeking file: %v\n", err)
		return
	}

	tempFile, err := os.CreateTemp("", "uploaded-*.zip")
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to create temp file: %v", err), http.StatusInternalServerError)
		log.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		respondWithError(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		log.Printf("Error saving file: %v\n", err)
		return
	}

	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		respondWithError(w, "Uploaded file is not a valid ZIP archive", http.StatusBadRequest)
		log.Printf("Error opening zip file: %v\n", err)
		return
	}
	defer zipReader.Close()

	var archiveInfo model.ArchiveInfo
	archiveInfo.FileName = header.Filename
	archiveInfo.ArchiveSize = float64(header.Size)
	archiveInfo.Files = make([]model.FileInfo, 0)

	var totalSize float64
	for _, f := range zipReader.File {
		fileInfo := model.FileInfo{
			FilePath: f.Name,
			Size:     float64(f.UncompressedSize64),
			MimeType: detectMimeType(f),
		}
		archiveInfo.Files = append(archiveInfo.Files, fileInfo)
		totalSize += fileInfo.Size
	}

	archiveInfo.TotalFiles = len(archiveInfo.Files)
	archiveInfo.TotalSize = totalSize

	log.Println("Form data:", r.MultipartForm)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(archiveInfo); err != nil {
		respondWithError(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		log.Printf("Error encoding response: %v\n", err)
	}
}

// respondWithError отправляет ответ с ошибкой в формате JSON
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResponse := ErrorResponse{Message: message}
	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		log.Printf("Error encoding error response: %v\n", err)
	}
}

func isArchive(file io.Reader) bool {
	buf := make([]byte, 4)
	_, err := file.Read(buf)
	if err != nil {
		return false
	}

	return strings.HasPrefix(string(buf), "PK\x03\x04")
}

func detectMimeType(file *zip.File) string {
	reader, err := file.Open()
	if err != nil {
		return "unknown"
	}
	defer reader.Close()

	buffer := make([]byte, 512)
	_, err = reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "unknown"
	}

	return mime.TypeByExtension(filepath.Ext(file.Name))
}
