package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// AllowedMimeTypes содержит список допустимых MIME-типов
var AllowedMimeTypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/xml": true,
	"image/jpeg":      true,
	"image/png":       true,
}

// ErrorResponse структура для ответа с ошибкой
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HandleCreateArchive обрабатывает запрос на создание ZIP-архива из списка файлов
func HandleCreateArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	// Ограничиваем размер загружаемого файла (например, до 50MB)
	err := r.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		log.Printf("Error parsing multipart form: %v\n", err)
		return
	}

	// Получаем файлы из multipart/form-data
	files, ok := r.MultipartForm.File["files[]"]
	if !ok || len(files) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "No files provided or invalid key used")
		log.Println("No files provided or invalid key used")
		return
	}

	// Создаем временный файл для ZIP-архива
	tempFile, err := os.CreateTemp("", "archive-*.zip")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create temp file: "+err.Error())
		log.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tempFile.Name()) // Удаляем файл после обработки
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	for _, header := range files {
		contentType := header.Header.Get("Content-Type")
		if !AllowedMimeTypes[contentType] {
			errMsg := fmt.Sprintf("File %s has invalid MIME type: %s", header.Filename, contentType)
			log.Printf("Unsupported MIME type: %s\n", contentType)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			continue // Пропускаем некорректный файл
		}

		file, err := header.Open()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error opening file %s: %v", header.Filename, err))
			log.Printf("Error opening file %s: %v\n", header.Filename, err)
			continue
		}
		defer file.Close()

		zipFile, err := zipWriter.Create(header.Filename)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error creating entry for %s in zip file: %v", header.Filename, err))
			log.Printf("Error creating entry for %s in zip file: %v\n", header.Filename, err)
			continue
		}

		if _, err := io.Copy(zipFile, file); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error writing file %s to zip: %v", header.Filename, err))
			log.Printf("Error writing file %s to zip: %v\n", header.Filename, err)
			continue
		}
	}

	// Закрытие zipWriter
	if err := zipWriter.Close(); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to close zip writer: "+err.Error())
		log.Printf("Error closing zip writer: %v\n", err)
		return
	}

	// Отправляем архив клиенту
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")

	http.ServeFile(w, r, tempFile.Name())
}

// writeErrorResponse записывает ошибку в формате JSON
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := ErrorResponse{
		Message: message,
		Code:    statusCode,
	}
	json.NewEncoder(w).Encode(errorResponse)
}
